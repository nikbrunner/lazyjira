package git

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
)

var worktreeSeparators = regexp.MustCompile(`[^a-z0-9]+`)

func GenerateWorktreeName(repoName, key, summary, format string) (string, error) {
	if format == "" {
		format = "{{if .RepoName}}{{.RepoName}}-{{end}}{{.Key}}-{{.Summary}}"
	}
	tmpl, err := template.New("worktree").Parse(format)
	if err != nil {
		return "", err
	}
	slug := func(s string) string {
		return strings.Trim(worktreeSeparators.ReplaceAllString(strings.ToLower(s), "-"), "-")
	}
	data := struct {
		RepoName string
		Key      string
		Summary  string
	}{RepoName: slug(repoName), Key: strings.ToLower(key), Summary: slug(summary)}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", err
	}
	name := []rune(strings.ToLower(out.String()))
	if len(name) > 50 {
		name = name[:50]
	}
	return strings.TrimRight(string(name), "-"), nil
}
