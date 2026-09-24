package config

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"text/template"

	"github.com/nikbrunner/lazyjira/v2/pkg/ascii"
)

// Context identifies a UI state in which a custom command may fire.
type Context string

const (
	CtxIssues         Context = "issues"
	CtxInfo           Context = "info"
	CtxProjects       Context = "projects"
	CtxDetail         Context = "detail"
	CtxDetailComments Context = "detail.comments"
)

// DefaultCommandContexts is applied when a command omits `contexts:`.
var DefaultCommandContexts = []Context{CtxIssues, CtxInfo, CtxDetail}

// ScopeMask is a bitmask of data scopes a command expects.
type ScopeMask uint8

const (
	ScopeIssue ScopeMask = 1 << iota
	ScopeProject
	ScopeComment
)

var contextScopes = map[Context]ScopeMask{
	CtxIssues:         ScopeIssue,
	CtxInfo:           ScopeIssue,
	CtxProjects:       ScopeProject,
	CtxDetail:         ScopeIssue,
	CtxDetailComments: ScopeIssue | ScopeComment,
}

// ResolvedCustomCommand is a validated, pre-computed custom command.
type ResolvedCustomCommand struct {
	Key      string
	Name     string
	Command  string
	Suspend  *bool
	Refresh  bool
	Contexts []Context
	Scopes   ScopeMask
	template *template.Template
	shell    string
}

// ShouldSuspend mirrors CustomCommandConfig.ShouldSuspend.
func (r ResolvedCustomCommand) ShouldSuspend() bool {
	return r.Suspend == nil || *r.Suspend
}

// HasContext reports whether the command is bound to the given context.
func (r ResolvedCustomCommand) HasContext(c Context) bool {
	return slices.Contains(r.Contexts, c)
}

// Shell returns the executable path used to validate and render this command.
func (r ResolvedCustomCommand) Shell() string { return r.shell }

func (r ResolvedCustomCommand) Render(values map[string]string) (string, error) {
	if r.template == nil {
		return "", fmt.Errorf("custom command %q has no validated template", r.Name)
	}
	state, err := newShellRenderState()
	if err != nil {
		return "", err
	}
	tmpl, err := r.template.Clone()
	if err != nil {
		return "", err
	}
	tmpl.Funcs(template.FuncMap{
		"__shellarg": func(value any) any {
			return shellTemplateValue{value: templateString(value), state: state}
		},
	})
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, values); err != nil {
		return "", err
	}
	return renderAndValidateShell(rendered.String(), state, r.shell)
}

func shellescape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// slugify produces an ASCII slug: lowercase [a-z0-9], other runes
// collapsed to '-'. Delegates ASCII normalization to ascii.Convert so
// the transliteration rules (ä->ae, ß->ss, NFD accent strip) stay
// shared with branch-name sanitization.
func slugify(s string) string {
	s = ascii.Convert(s)

	var b strings.Builder
	b.Grow(len(s))
	prevDash := true // suppress leading dashes
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}

func templateString(value any) string {
	switch value := value.(type) {
	case string:
		return value
	default:
		return fmt.Sprint(value)
	}
}

func shellescapeTemplateValue(value any) string { return templateString(value) }
func shellraw(value any) string                 { return templateString(value) }
func slugifyTemplateValue(value any) string     { return slugify(templateString(value)) }

var commandFuncMap = template.FuncMap{
	"__shellarg":  func(value any) any { return value },
	"shellescape": shellescapeTemplateValue,
	"shellraw":    shellraw,
	"slugify":     slugifyTemplateValue,
}

// ResolveCustomCommands validates the flat CustomCommands list and returns
// pre-computed entries with typed contexts, scope mask and parsed template.
func (c *Config) ResolveCustomCommands() ([]ResolvedCustomCommand, error) {
	return c.ResolveCustomCommandsForShell("sh")
}

// ResolveCustomCommandsForShell validates custom command templates for the shell that executes them.
func (c *Config) ResolveCustomCommandsForShell(shell string) ([]ResolvedCustomCommand, error) {
	out := make([]ResolvedCustomCommand, 0, len(c.CustomCommands))
	type keyCtx struct {
		key string
		ctx Context
	}
	seen := make(map[keyCtx]bool)

	for i, entry := range c.CustomCommands {
		if entry.Key == "" || entry.Name == "" || entry.Command == "" {
			return nil, fmt.Errorf("customCommands[%d] (%q): key, name and command must all be non-empty", i, entry.Name)
		}

		var ctxs []Context
		if len(entry.Contexts) == 0 {
			ctxs = append(ctxs, DefaultCommandContexts...)
		} else {
			for _, raw := range entry.Contexts {
				ctx := Context(raw)
				if _, ok := contextScopes[ctx]; !ok {
					return nil, fmt.Errorf("customCommands[%d] (%q): unknown context %q", i, entry.Key, raw)
				}
				ctxs = append(ctxs, ctx)
			}
		}

		var scopes ScopeMask
		for _, ctx := range ctxs {
			scopes |= contextScopes[ctx]
			k := keyCtx{entry.Key, ctx}
			if seen[k] {
				return nil, fmt.Errorf("customCommands: duplicate key %q in context %q", entry.Key, ctx)
			}
			seen[k] = true
		}

		tmpl, err := template.New("customCommand").
			Option("missingkey=error").
			Funcs(commandFuncMap).
			Parse(entry.Command)
		if err != nil {
			return nil, fmt.Errorf("customCommands[%d] (%q): template parse error: %w", i, entry.Key, err)
		}
		if err := validateCustomCommandTemplate(tmpl, shell); err != nil {
			return nil, fmt.Errorf("customCommands[%d] (%q): template validation error: %w", i, entry.Key, err)
		}

		out = append(out, ResolvedCustomCommand{
			Key:      entry.Key,
			Name:     entry.Name,
			Command:  entry.Command,
			Suspend:  entry.Suspend,
			Refresh:  entry.Refresh,
			Contexts: ctxs,
			Scopes:   scopes,
			template: tmpl,
			shell:    shell,
		})
	}
	return out, nil
}
