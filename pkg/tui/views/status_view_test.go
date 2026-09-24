package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nikbrunner/lazyjira/v2/pkg/internal/testkit"
)

const testEmail = "user@example.com"
const testHost = "https://jira.example.com"
const testProject = "PLAT"

func makeStatusPanel() *StatusPanel {
	return NewStatusPanel(testProject, testEmail, testHost)
}

func TestStatusPanel_NewStatusPanel_SetsDefaults(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	testkit.AssertEqual(t, "error empty", panel.ErrorMessage(), "")
}

func TestStatusPanel_SetProject_DoesNotReplaceAccountStatus(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	panel.SetProject("NEWPROJ")
	panel.SetSize(80, 10)
	output := stripANSI(panel.View())
	if strings.Contains(output, "NEWPROJ") {
		t.Errorf("View() = %q, should not display project as account status", output)
	}
}

func TestStatusPanel_SetOnline_OfflineShowsX(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	panel.SetOnline(false)
	panel.SetSize(80, 10)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "✗") {
		t.Errorf("offline panel output = %q, want ✗ indicator", output)
	}
}

func TestStatusPanel_SetOnline_OnlineShowsCheck(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	panel.SetOnline(true)
	panel.SetSize(80, 10)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "✓") {
		t.Errorf("online panel output = %q, want ✓ indicator", output)
	}
}

func TestStatusPanel_SetError_ShowsInView(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	panel.SetError("connection failed")
	panel.SetSize(80, 10)
	testkit.AssertEqual(t, "error message", panel.ErrorMessage(), "connection failed")
	output := stripANSI(panel.View())
	if !strings.Contains(output, "connection failed") {
		t.Errorf("View() = %q, want to contain error text", output)
	}
}

func TestStatusPanel_SetFocused_ReturnsNonEmptyView(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	panel.SetSize(80, 10)
	panel.SetFocused(true)
	if panel.View() == "" {
		t.Error("expected non-empty View() when focused")
	}
}

func TestStatusPanel_Init_ReturnsNil(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	if cmd := panel.Init(); cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestStatusPanel_Update_ReturnsSelf(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	result, cmd := panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if result != panel {
		t.Error("Update() should return same panel pointer")
	}
	if cmd != nil {
		t.Error("Update() should return nil cmd")
	}
}

func TestStatusPanel_View_CollapsedBar_WhenHeightOne(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	panel.SetSize(80, 1)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "Status") {
		t.Errorf("collapsed bar = %q, want Status label", output)
	}
}

func TestStatusPanel_View_ShowsEmail(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	panel.SetSize(80, 10)
	output := stripANSI(panel.View())
	if !strings.Contains(output, testEmail) {
		t.Errorf("View() = %q, want to contain email %s", output, testEmail)
	}
}

func TestStatusPanel_View_ShowsConnectionDetails(t *testing.T) {
	t.Parallel()
	panel := makeStatusPanel()
	panel.SetAuthMethod("Environment variables")
	panel.SetVersion("v1.2.3")
	panel.SetSize(80, 10)
	output := stripANSI(panel.View())
	for _, want := range []string{testEmail, testHost, "Environment variables", "v1.2.3", "Connected"} {
		if !strings.Contains(output, want) {
			t.Errorf("View() = %q, want to contain %q", output, want)
		}
	}
}

func TestStatusPanel_View_UsesHostWhenEmailMissing(t *testing.T) {
	t.Parallel()
	panel := NewStatusPanel("", "", testHost)
	panel.SetSize(80, 10)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "Account: "+testHost) {
		t.Errorf("View() = %q, want host account fallback", output)
	}
}

func TestStatusPanel_View_NarrowPrioritizesConnectionAndErrorAndFitsWidth(t *testing.T) {
	t.Parallel()
	panel := NewStatusPanel("", strings.Repeat("用", 20), testHost)
	panel.SetOnline(false)
	panel.SetError(strings.Repeat("界", 20))
	panel.SetSize(18, 10)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "Disconnected") || !strings.Contains(output, "Error:") {
		t.Fatalf("View() = %q, want connection state and error", output)
	}
	if !strings.Contains(output, "✗ Disconnected") {
		t.Fatalf("View() = %q, want complete connection state", output)
	}
	for _, line := range strings.Split(output, "\n") {
		if got := lipgloss.Width(line); got > 18 {
			t.Errorf("line width = %d, exceeds panel width 18: %q", got, line)
		}
	}
}

func TestStatusPanel_View_LongEmailTruncated(t *testing.T) {
	t.Parallel()
	longEmail := strings.Repeat("a", 100) + "@example.com"
	panel := NewStatusPanel("PROJ", longEmail, testHost)
	panel.SetSize(40, 10)
	output := stripANSI(panel.View())
	if strings.Contains(output, longEmail) {
		t.Errorf("expected long email to be truncated in View()")
	}
	if !strings.Contains(output, "…") {
		t.Errorf("expected ellipsis in truncated email, got %q", output)
	}
}
