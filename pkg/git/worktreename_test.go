package git

import "testing"

func TestGenerateWorktreeName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		repo    string
		summary string
		format  string
		want    string
		wantErr bool
	}{
		{name: "default", repo: "Web UI", summary: "Audit skills", want: "web-ui-websdk-218-audit-skills"},
		{name: "without repository", summary: "Audit skills", want: "websdk-218-audit-skills"},
		{name: "custom format", repo: "web-ui", summary: "Audit skills", format: "{{.RepoName}}/{{.Key}}_{{.Summary}}", want: "web-ui/websdk-218_audit-skills"},
		{name: "lowercase literals", summary: "Audit skills", format: "FIX/{{.Key}}", want: "fix/websdk-218"},
		{name: "shell slug rules", repo: "Web UI", summary: "  Fix: login & tests! (v2)  ", want: "web-ui-websdk-218-fix-login-tests-v2"},
		{name: "ASCII shell slug", summary: "Größe", want: "websdk-218-gr-e"},
		{name: "50 character limit", repo: "web-ui", summary: "Web UI audit skills with skill creator", want: "web-ui-websdk-218-web-ui-audit-skills-with-skill-c"},
		{name: "trailing hyphen after truncation", summary: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa b", want: "websdk-218-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{name: "malformed format", format: "{{.Key", wantErr: true},
		{name: "unknown field", format: "{{.Unknown}}", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := GenerateWorktreeName(tt.repo, "WEBSDK-218", tt.summary, tt.format)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want error %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("name = %q, want %q", got, tt.want)
			}
		})
	}
}
