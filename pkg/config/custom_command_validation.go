package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"text/template/parse"

	"mvdan.cc/sh/v3/syntax"
)

const maxCommandTemplateVariants = 64

type shellPlaceholder struct {
	marker string
	raw    bool
}

type shellTemplateSource struct {
	text         string
	placeholders []shellPlaceholder
	hasRaw       bool
}

type shellRenderState struct {
	prefix  string
	values  map[string]string
	markers []string
}

type shellTemplateValue struct {
	value string
	state *shellRenderState
}

func (v shellTemplateValue) String() string {
	return v.state.add(v.value)
}

func newShellRenderState() (*shellRenderState, error) {
	prefix, err := randomMarkerPrefix()
	if err != nil {
		return nil, err
	}
	return &shellRenderState{prefix: prefix, values: make(map[string]string)}, nil
}

func (s *shellRenderState) add(value string) string {
	marker := fmt.Sprintf("__LAZYJIRA_%s_%d__", s.prefix, len(s.markers))
	s.values[marker] = value
	s.markers = append(s.markers, marker)
	return marker
}

func randomMarkerPrefix() (string, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", fmt.Errorf("generate template validation token: %w", err)
	}
	return hex.EncodeToString(nonce[:]), nil
}

func validateCustomCommandTemplate(tmpl *template.Template, shell string) error {
	prefix, err := randomMarkerPrefix()
	if err != nil {
		return err
	}
	variants, err := expandCommandTemplate(tmpl.Root.Nodes, prefix, new(int))
	if err != nil {
		return err
	}

	needsShell := false
	for _, source := range variants {
		for _, placeholder := range source.placeholders {
			needsShell = needsShell || !placeholder.raw
		}
	}
	if needsShell {
		lang, ok := shellLanguage(shell)
		if !ok {
			return fmt.Errorf("safe template values require sh, bash, zsh, or mksh; configured shell is %q", filepath.Base(shell))
		}
		for _, source := range variants {
			if source.hasRaw || len(source.placeholders) == 0 {
				continue
			}
			file, err := syntax.NewParser(syntax.Variant(lang)).Parse(strings.NewReader(source.text), "custom command")
			if err != nil {
				return fmt.Errorf("cannot validate template values for %s: %w", lang, err)
			}
			for _, placeholder := range source.placeholders {
				if !placeholder.raw && !isUnquotedCommandArgument(file, source.text, placeholder.marker) {
					return errors.New("template values must be entire, unquoted command arguments")
				}
			}
		}
	}

	protectTemplateActions(tmpl.Root.Nodes, tmpl.Tree)
	return nil
}

func shellLanguage(shell string) (syntax.LangVariant, bool) {
	switch filepath.Base(shell) {
	case "sh", "dash", "ash":
		return syntax.LangPOSIX, true
	case "bash":
		return syntax.LangBash, true
	case "zsh":
		return syntax.LangZsh, true
	case "mksh":
		return syntax.LangMirBSDKorn, true
	default:
		return 0, false
	}
}

func expandCommandTemplate(nodes []parse.Node, prefix string, nextMarker *int) ([]shellTemplateSource, error) {
	variants := []shellTemplateSource{{}}
	for _, node := range nodes {
		var options []shellTemplateSource
		var err error
		switch node := node.(type) {
		case *parse.TextNode:
			options = []shellTemplateSource{{text: string(node.Text)}}
		case *parse.ActionNode:
			raw, err := validateTemplateAction(node.Pipe)
			if err != nil {
				return nil, err
			}
			marker := fmt.Sprintf("__LAZYJIRA_%s_%d__", prefix, *nextMarker)
			(*nextMarker)++
			options = []shellTemplateSource{{
				text:         marker,
				placeholders: []shellPlaceholder{{marker: marker, raw: raw}},
				hasRaw:       raw,
			}}
		case *parse.IfNode:
			options, err = expandCommandTemplate(node.List.Nodes, prefix, nextMarker)
			if err != nil {
				return nil, err
			}
			if node.ElseList != nil {
				elseOptions, err := expandCommandTemplate(node.ElseList.Nodes, prefix, nextMarker)
				if err != nil {
					return nil, err
				}
				options = append(options, elseOptions...)
			} else {
				options = append(options, shellTemplateSource{})
			}
		case *parse.CommentNode:
			continue
		default:
			return nil, fmt.Errorf("unsupported Go template construct %T; use field values and approved helpers", node)
		}

		if len(variants)*len(options) > maxCommandTemplateVariants {
			return nil, fmt.Errorf("custom command template has more than %d possible branches", maxCommandTemplateVariants)
		}
		variants = combineShellSources(variants, options)
	}
	return variants, nil
}

func combineShellSources(prefixes, suffixes []shellTemplateSource) []shellTemplateSource {
	out := make([]shellTemplateSource, 0, len(prefixes)*len(suffixes))
	for _, prefix := range prefixes {
		for _, suffix := range suffixes {
			var text strings.Builder
			text.Grow(len(prefix.text))
			text.WriteString(prefix.text)
			text.WriteString(suffix.text)
			combined := shellTemplateSource{
				text:   text.String(),
				hasRaw: prefix.hasRaw || suffix.hasRaw,
			}
			combined.placeholders = append(combined.placeholders, prefix.placeholders...)
			combined.placeholders = append(combined.placeholders, suffix.placeholders...)
			out = append(out, combined)
		}
	}
	return out
}

func validateTemplateAction(pipe *parse.PipeNode) (bool, error) {
	if pipe == nil || len(pipe.Decl) != 0 || len(pipe.Cmds) == 0 {
		return false, errors.New("template actions must reference a command field")
	}
	first := pipe.Cmds[0]
	if len(first.Args) != 1 {
		return false, errors.New("template actions must reference one command field")
	}
	field, ok := first.Args[0].(*parse.FieldNode)
	if !ok || len(field.Ident) != 1 {
		return false, errors.New("template actions must reference one top-level command field")
	}
	if len(pipe.Cmds) == 1 {
		return false, nil
	}
	if len(pipe.Cmds) != 2 || len(pipe.Cmds[1].Args) != 1 {
		return false, errors.New("unsupported template formatting; use shellescape, slugify, or shellraw")
	}
	helper, ok := pipe.Cmds[1].Args[0].(*parse.IdentifierNode)
	if !ok {
		return false, errors.New("unsupported template formatting; use shellescape, slugify, or shellraw")
	}
	switch helper.Ident {
	case "shellescape", "slugify":
		return false, nil
	case "shellraw":
		return true, nil
	default:
		return false, fmt.Errorf("unsupported template helper %q; use shellescape, slugify, or shellraw", helper.Ident)
	}
}

func protectTemplateActions(nodes []parse.Node, tree *parse.Tree) {
	for _, node := range nodes {
		switch node := node.(type) {
		case *parse.ActionNode:
			raw, _ := validateTemplateAction(node.Pipe)
			if !raw {
				identifier := parse.NewIdentifier("__shellarg").SetTree(tree).SetPos(node.Pipe.Pos)
				node.Pipe.Cmds = append(node.Pipe.Cmds, &parse.CommandNode{
					NodeType: parse.NodeCommand,
					Pos:      node.Pipe.Pos,
					Args:     []parse.Node{identifier},
				})
			}
		case *parse.IfNode:
			protectTemplateActions(node.List.Nodes, tree)
			if node.ElseList != nil {
				protectTemplateActions(node.ElseList.Nodes, tree)
			}
		}
	}
}

func renderAndValidateShell(source string, state *shellRenderState, shell string) (string, error) {
	if len(state.markers) == 0 {
		return source, nil
	}
	lang, ok := shellLanguage(shell)
	if !ok {
		return "", fmt.Errorf("safe template values require sh, bash, zsh, or mksh; configured shell is %q", filepath.Base(shell))
	}
	file, err := syntax.NewParser(syntax.Variant(lang)).Parse(strings.NewReader(source), "custom command")
	if err != nil {
		return "", fmt.Errorf("cannot validate rendered command for %s: %w", lang, err)
	}
	pairs := make([]string, 0, len(state.markers)*2)
	for _, marker := range state.markers {
		if !isUnquotedCommandArgument(file, source, marker) {
			return "", errors.New("template values must be entire, unquoted command arguments")
		}
		pairs = append(pairs, marker, shellescape(state.values[marker]))
	}
	return strings.NewReplacer(pairs...).Replace(source), nil
}

func isUnquotedCommandArgument(file *syntax.File, source, marker string) bool {
	if strings.Count(source, marker) != 1 {
		return false
	}
	markerStart, markerEnd, ok := markerByteOffsets(source, marker)
	if !ok {
		return false
	}

	var ancestors []syntax.Node
	matches, argument, unsafe := 0, false, false
	syntax.Walk(file, func(node syntax.Node) bool {
		if node == nil {
			if len(ancestors) > 0 {
				ancestors = ancestors[:len(ancestors)-1]
			}
			return true
		}
		ancestors = append(ancestors, node)
		lit, ok := node.(*syntax.Lit)
		if !ok || lit.Value != marker {
			return true
		}
		matches++
		var word *syntax.Word
		var call *syntax.CallExpr
		for _, ancestor := range slices.Backward(ancestors) {
			switch ancestor := ancestor.(type) {
			case *syntax.DblQuoted, *syntax.SglQuoted, *syntax.ParamExp,
				*syntax.CmdSubst, *syntax.ArithmExp, *syntax.ArithmCmd,
				*syntax.ProcSubst, *syntax.ExtGlob, *syntax.BraceExp:
				unsafe = true
			case *syntax.Word:
				if word == nil {
					word = ancestor
				}
			case *syntax.CallExpr:
				if call == nil {
					call = ancestor
				}
			}
		}
		if word == nil || call == nil || len(word.Parts) != 1 || word.Lit() != marker {
			unsafe = true
		} else {
			part, ok := word.Parts[0].(*syntax.Lit)
			if !ok || part.Value != marker ||
				word.Pos().Offset() != markerStart || word.End().Offset() != markerEnd {
				unsafe = true
			}
			for i, arg := range call.Args {
				if arg != word {
					continue
				}
				if i == 0 {
					unsafe = true
				} else {
					argument = true
				}
			}
		}
		return true
	})
	return matches == 1 && argument && !unsafe
}

func markerByteOffsets(source, marker string) (uint, uint, bool) {
	if strings.Count(source, marker) != 1 {
		return 0, 0, false
	}
	var byteOffset uint
	for i := 0; i <= len(source)-len(marker); i++ {
		if source[i:i+len(marker)] == marker {
			end := byteOffset
			for range marker {
				end++
			}
			return byteOffset, end, true
		}
		byteOffset++
	}
	return 0, 0, false
}
