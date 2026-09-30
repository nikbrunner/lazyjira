package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nikbrunner/lazyjira/pkg/tui/theme"
)

const (
	keyEsc   = "esc"
	keyEnter = "enter"
	keyDown  = "down"
	keyCtrlD = "ctrl+d"
	keyCtrlU = "ctrl+u"
)

// CreatePanel identifies which sub-panel of the create form is focused
type CreatePanel int

const (
	CreatePanelSummary CreatePanel = iota
	CreatePanelDescription
	CreatePanelFields
	createPanelCount = 3
)

// CreateFormField holds one field in the create issue form
type CreateFormField struct {
	Name          string
	FieldID       string
	Type          int
	Value         any
	DisplayValue  string
	Required      bool
	AllowedValues []ModalItem
	HasError      bool
	SchemaItems   string
}

const (
	CFFieldSingleSelect = iota
	CFFieldMultiSelect
	CFFieldPerson
	CFFieldSingleText
	CFFieldMultiText
)

type CreateFormTypeSelectedMsg struct {
	TypeID   string
	TypeName string
}
type CreateFormEditTextMsg struct{ FieldIndex int }
type CreateFormEditExternalMsg struct{ FieldIndex int }
type CreateFormPickerMsg struct {
	FieldIndex int
	Items      []ModalItem
}
type CreateFormChecklistMsg struct {
	FieldIndex int
	Items      []ModalItem
}
type CreateFormSubmitMsg struct{ Fields map[string]any }
type CreateFormCancelMsg struct{}

// CreateForm is a 3-panel overlay for issue creation
type CreateForm struct {
	visible bool
	width   int
	height  int

	issueTypeName string
	projectKey    string
	focusedPanel  CreatePanel

	allFields []CreateFormField

	summaryText   []rune
	summaryCursor int
	summaryIdx    int

	descIdx int
	// desc is nil until ShowForm; Hide drops it with the rest of the form state.
	desc *textarea.Model

	fieldIndices  []int
	fieldCursor   int
	fieldOffset   int
	fieldDblClick DblClickDetector

	paused      bool
	errorMsg    string
	loading     bool
	filterInput TextInput
	filtering   bool
}

// NewCreateForm constructs a hidden CreateForm.
func NewCreateForm() CreateForm {
	return CreateForm{summaryIdx: -1, descIdx: -1}
}

func newDescArea() textarea.Model {
	ta := textarea.New()
	ta.Prompt = " "
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.MaxHeight = 0
	// ctrl+v would emit textarea's private paste message, which never reaches
	// the form; terminal paste arrives as a KeyMsg instead.
	ta.KeyMap.Paste.SetEnabled(false)
	ta.Cursor.SetMode(cursor.CursorStatic)
	ta.Cursor.Style = lipgloss.NewStyle().Foreground(theme.ColorCyan)
	style, _ := textarea.DefaultStyles()
	style.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle = style
	ta.BlurredStyle = style
	ta.Blur()
	return ta
}

// Pause stops intercepting keys so sub-overlays can receive input
func (f *CreateForm) Pause() { f.paused = true }

// Resume resumes key interception after sub-overlay closes
func (f *CreateForm) Resume() { f.paused = false }

// FocusedPanel returns which sub-panel is currently focused
func (f *CreateForm) FocusedPanel() CreatePanel { return f.focusedPanel }

// IsFiltering returns true when the fields filter input is active
func (f *CreateForm) IsFiltering() bool { return f.filtering }

// FilterQuery returns the current filter text
func (f *CreateForm) FilterQuery() string { return f.filterInput.Value() }

// FilterBarView renders the filter bar with cursor positioning
func (f *CreateForm) FilterBarView() string { return RenderFilterBarInput(&f.filterInput) }

// DescriptionText returns the raw Markdown in the Description textarea.
func (f *CreateForm) DescriptionText() string {
	if f.desc == nil {
		return ""
	}
	return f.desc.Value()
}

// SetDescriptionText replaces the Description textarea content.
func (f *CreateForm) SetDescriptionText(text string) {
	if f.desc != nil {
		f.desc.SetValue(text)
	}
}

func (f *CreateForm) ShowForm(fields []CreateFormField, issueTypeName, projectKey string) {
	f.visible = true
	f.issueTypeName = issueTypeName
	f.projectKey = projectKey
	f.allFields = fields
	f.summaryIdx = -1
	f.descIdx = -1
	f.fieldIndices = nil
	f.fieldCursor = 0
	f.fieldOffset = 0
	f.errorMsg = ""
	f.loading = false
	f.filterInput.SetValue("")
	f.filtering = false
	desc := newDescArea()
	f.desc = &desc

	for i, fld := range fields {
		switch fld.FieldID {
		case "summary":
			f.summaryIdx = i
			f.summaryText = []rune(fld.DisplayValue)
			f.summaryCursor = len(f.summaryText)
		case "description":
			f.descIdx = i
			f.desc.SetValue(fld.DisplayValue)
		default:
			f.fieldIndices = append(f.fieldIndices, i)
		}
	}
	f.setFocus(CreatePanelSummary)
	f.sizeDesc(f.width, f.height)
}

func (f *CreateForm) setFocus(p CreatePanel) {
	f.focusedPanel = p
	if f.desc == nil {
		return
	}
	if p == CreatePanelDescription {
		f.desc.Focus()
	} else {
		f.desc.Blur()
	}
}

func (f *CreateForm) Hide() {
	f.visible = false
	f.allFields = nil
	f.fieldIndices = nil
	f.summaryText = nil
	f.desc = nil
	f.errorMsg = ""
	f.loading = false
	f.filterInput.SetValue("")
	f.filtering = false
}

func (f *CreateForm) SetFieldValue(index int, value any, display string) {
	if index < 0 || index >= len(f.allFields) {
		return
	}
	f.allFields[index].Value = value
	if display == "" && !f.allFields[index].Required {
		display = "None"
	}
	f.allFields[index].DisplayValue = display
	f.allFields[index].HasError = false

	if index == f.summaryIdx {
		f.summaryText = []rune(display)
		f.summaryCursor = len(f.summaryText)
	}
}

func (f *CreateForm) SetError(msg string) {
	f.errorMsg = msg
	f.loading = false
}

func (f *CreateForm) SetLoading(loading bool) {
	f.loading = loading
	if loading {
		f.visible = true
	}
}

func (f *CreateForm) FieldAt(index int) *CreateFormField {
	if index < 0 || index >= len(f.allFields) {
		return nil
	}
	return &f.allFields[index]
}

func (f *CreateForm) IsVisible() bool { return f.visible }

func (f *CreateForm) SetSize(w, h int) {
	f.width = w
	f.height = h
	f.sizeDesc(w, h)
}

// Intercept handles keyboard and mouse input for the 3-panel form
func (f *CreateForm) Intercept(msg tea.Msg) (tea.Cmd, bool) {
	if !f.visible || f.paused {
		return nil, false
	}

	if mm, isMouse := msg.(tea.MouseMsg); isMouse {
		return f.interceptMouse(mm)
	}

	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil, false
	}
	if f.loading {
		if km.String() == keyEsc {
			f.Hide()
			return func() tea.Msg { return CreateFormCancelMsg{} }, true
		}
		return nil, true
	}
	if f.filtering {
		return f.interceptFilter(km)
	}

	switch km.Type { //nolint:exhaustive
	case tea.KeyTab:
		f.setFocus(CreatePanel((int(f.focusedPanel) + 1) % createPanelCount))
		return nil, true
	case tea.KeyShiftTab:
		f.setFocus(CreatePanel((int(f.focusedPanel) + createPanelCount - 1) % createPanelCount))
		return nil, true
	case tea.KeyCtrlS:
		return f.submitForm()
	}

	switch f.focusedPanel {
	case CreatePanelSummary:
		return f.interceptSummary(km)
	case CreatePanelDescription:
		return f.interceptDescription(km)
	case CreatePanelFields:
		return f.interceptFields(km)
	}
	return nil, true
}

func (f *CreateForm) interceptMouse(mm tea.MouseMsg) (tea.Cmd, bool) {
	switch mm.Button { //nolint:exhaustive
	case tea.MouseButtonWheelUp:
		f.scrollFocused(-3)
		return nil, true
	case tea.MouseButtonWheelDown:
		f.scrollFocused(3)
		return nil, true
	case tea.MouseButtonLeft:
		if mm.Action != tea.MouseActionPress {
			return nil, true
		}
	default:
		return nil, true
	}

	l := f.layout(f.width, f.height)
	switch {
	case l.summary.contains(mm.X, mm.Y):
		f.setFocus(CreatePanelSummary)
	case l.desc.contains(mm.X, mm.Y):
		f.setFocus(CreatePanelDescription)
	case l.fields.contains(mm.X, mm.Y):
		f.setFocus(CreatePanelFields)
		rowInPanel := mm.Y - l.fields.y - 1
		innerH := max(l.fields.h-2, 1)
		if rowInPanel >= 0 && rowInPanel < innerH {
			filtered := f.filteredFields()
			idx := f.fieldOffset + rowInPanel
			if idx >= 0 && idx < len(filtered) {
				f.fieldCursor = idx
				if f.fieldDblClick.Click(idx) {
					return f.editCurrentField(filtered)
				}
			}
		}
	}

	return nil, true
}

// scrollFocused scrolls the currently focused panel by delta lines
func (f *CreateForm) scrollFocused(delta int) {
	switch f.focusedPanel {
	case CreatePanelSummary:
		// move cursor by delta chars so the view follows
		f.summaryCursor += delta * 10
		if f.summaryCursor < 0 {
			f.summaryCursor = 0
		}
		if f.summaryCursor > len(f.summaryText) {
			f.summaryCursor = len(f.summaryText)
		}
	case CreatePanelDescription:
		f.scrollDesc(delta)
	case CreatePanelFields:
		filtered := f.filteredFields()
		f.fieldCursor += delta
		if f.fieldCursor < 0 {
			f.fieldCursor = 0
		}
		if f.fieldCursor >= len(filtered) {
			f.fieldCursor = max(len(filtered)-1, 0)
		}
		f.ensureFieldVisible()
	}
}

// scrollDesc moves the textarea cursor by delta lines through Update, which is
// what makes the textarea scroll its view to follow the cursor.
func (f *CreateForm) scrollDesc(delta int) {
	if f.desc == nil {
		return
	}
	step := tea.KeyMsg{Type: tea.KeyDown}
	if delta < 0 {
		step = tea.KeyMsg{Type: tea.KeyUp}
		delta = -delta
	}
	for range delta {
		*f.desc, _ = f.desc.Update(step)
	}
}

func (f *CreateForm) interceptFilter(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case keyEsc:
		f.filtering = false
		f.filterInput.SetValue("")
		f.fieldCursor = 0
		f.fieldOffset = 0
	case keyEnter:
		f.confirmFilter()
	case keyDown, KeyCtrlJ:
		filtered := f.filteredFields()
		if f.fieldCursor < len(filtered)-1 {
			f.fieldCursor++
			f.ensureFieldVisible()
		}
	case "up", KeyCtrlK:
		if f.fieldCursor > 0 {
			f.fieldCursor--
			f.ensureFieldVisible()
		}
	default:
		updated, changed := f.filterInput.Update(msg)
		f.filterInput = updated
		if changed {
			f.fieldCursor = 0
			f.fieldOffset = 0
		}
	}
	return nil, true
}

func (f *CreateForm) interceptSummary(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type { //nolint:exhaustive
	case tea.KeyEnter:
		return f.submitForm()
	case tea.KeyEsc:
		f.Hide()
		return func() tea.Msg { return CreateFormCancelMsg{} }, true
	case tea.KeyBackspace:
		if f.summaryCursor > 0 {
			f.summaryText = append(f.summaryText[:f.summaryCursor-1], f.summaryText[f.summaryCursor:]...)
			f.summaryCursor--
		}
	case tea.KeyDelete:
		if f.summaryCursor < len(f.summaryText) {
			f.summaryText = append(f.summaryText[:f.summaryCursor], f.summaryText[f.summaryCursor+1:]...)
		}
	case tea.KeyLeft:
		if f.summaryCursor > 0 {
			f.summaryCursor--
		}
	case tea.KeyRight:
		if f.summaryCursor < len(f.summaryText) {
			f.summaryCursor++
		}
	case tea.KeyHome, tea.KeyCtrlA:
		f.summaryCursor = 0
	case tea.KeyEnd, tea.KeyCtrlE:
		f.summaryCursor = len(f.summaryText)
	case tea.KeyCtrlU:
		f.summaryText = f.summaryText[f.summaryCursor:]
		f.summaryCursor = 0
	case tea.KeyCtrlK:
		f.summaryText = f.summaryText[:f.summaryCursor]
	case tea.KeySpace:
		f.insertSummaryRunes([]rune{' '})
	case tea.KeyRunes:
		f.insertSummaryRunes(msg.Runes)
	}
	return nil, true
}

func (f *CreateForm) insertSummaryRunes(runes []rune) {
	newText := make([]rune, 0, len(f.summaryText)+len(runes))
	newText = append(newText, f.summaryText[:f.summaryCursor]...)
	newText = append(newText, runes...)
	newText = append(newText, f.summaryText[f.summaryCursor:]...)
	f.summaryText = newText
	f.summaryCursor += len(runes)
}

func (f *CreateForm) interceptDescription(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.Type { //nolint:exhaustive
	case tea.KeyEsc:
		f.Hide()
		return func() tea.Msg { return CreateFormCancelMsg{} }, true
	case tea.KeyCtrlG:
		if f.descIdx >= 0 {
			idx := f.descIdx
			return func() tea.Msg { return CreateFormEditExternalMsg{FieldIndex: idx} }, true
		}
		return nil, true
	}
	if f.desc == nil {
		return nil, true
	}
	var cmd tea.Cmd
	*f.desc, cmd = f.desc.Update(msg)
	return cmd, true
}

func (f *CreateForm) interceptFields(msg tea.KeyMsg) (tea.Cmd, bool) {
	filtered := f.filteredFields()
	switch msg.String() {
	case "j", keyDown, KeyCtrlJ:
		if f.fieldCursor < len(filtered)-1 {
			f.fieldCursor++
			f.ensureFieldVisible()
		}
	case "k", "up", KeyCtrlK:
		if f.fieldCursor > 0 {
			f.fieldCursor--
			f.ensureFieldVisible()
		}
	case "g":
		f.fieldCursor = 0
		f.fieldOffset = 0
	case "G":
		if len(filtered) > 0 {
			f.fieldCursor = len(filtered) - 1
			f.ensureFieldVisible()
		}
	case keyCtrlD:
		half := f.fieldsInnerH() / 2
		f.fieldCursor = min(f.fieldCursor+half, max(len(filtered)-1, 0))
		f.ensureFieldVisible()
	case keyCtrlU:
		half := f.fieldsInnerH() / 2
		f.fieldCursor = max(f.fieldCursor-half, 0)
		f.ensureFieldVisible()
	case "/":
		f.filtering = true
		f.filterInput.SetValue("")
	case "e", " ":
		return f.editCurrentField(filtered)
	case keyEnter:
		return f.submitForm()
	case keyEsc, "q":
		f.Hide()
		return func() tea.Msg { return CreateFormCancelMsg{} }, true
	}
	return nil, true
}

func (f *CreateForm) ensureFieldVisible() {
	vh := f.fieldsInnerH()
	if f.fieldCursor < f.fieldOffset {
		f.fieldOffset = f.fieldCursor
	}
	if f.fieldCursor >= f.fieldOffset+vh {
		f.fieldOffset = f.fieldCursor - vh + 1
	}
}

func (f *CreateForm) fieldsInnerH() int {
	return max(f.layout(f.width, f.height).fields.h-2, 1)
}

func (f *CreateForm) filteredFields() []int {
	if f.filterInput.Value() == "" {
		return f.fieldIndices
	}
	lower := strings.ToLower(f.filterInput.Value())
	var indices []int
	for _, idx := range f.fieldIndices {
		fld := f.allFields[idx]
		if strings.Contains(strings.ToLower(fld.Name), lower) ||
			strings.Contains(strings.ToLower(fld.DisplayValue), lower) {
			indices = append(indices, idx)
		}
	}
	return indices
}

// confirmFilter restores full field list and places cursor on the matched field
func (f *CreateForm) confirmFilter() {
	filtered := f.filteredFields()
	var matchedIdx int
	if f.fieldCursor >= 0 && f.fieldCursor < len(filtered) {
		matchedIdx = filtered[f.fieldCursor]
	}
	f.filtering = false
	f.filterInput.SetValue("")
	f.fieldCursor = 0
	for i, idx := range f.fieldIndices {
		if idx == matchedIdx {
			f.fieldCursor = i
			break
		}
	}
	f.fieldOffset = 0
	f.ensureFieldVisible()
}

func (f *CreateForm) editCurrentField(filtered []int) (tea.Cmd, bool) {
	if f.fieldCursor < 0 || f.fieldCursor >= len(filtered) {
		return nil, true
	}
	idx := filtered[f.fieldCursor]
	field := f.allFields[idx]

	switch field.Type {
	case CFFieldSingleText:
		return func() tea.Msg { return CreateFormEditTextMsg{FieldIndex: idx} }, true
	case CFFieldMultiText:
		return func() tea.Msg { return CreateFormEditExternalMsg{FieldIndex: idx} }, true
	case CFFieldSingleSelect, CFFieldPerson:
		if len(field.AllowedValues) > 0 {
			return func() tea.Msg {
				return CreateFormPickerMsg{FieldIndex: idx, Items: field.AllowedValues}
			}, true
		}
		return func() tea.Msg {
			return CreateFormPickerMsg{FieldIndex: idx, Items: nil}
		}, true
	case CFFieldMultiSelect:
		return func() tea.Msg {
			return CreateFormChecklistMsg{FieldIndex: idx, Items: field.AllowedValues}
		}, true
	}
	return nil, true
}

func (f *CreateForm) submitForm() (tea.Cmd, bool) {
	// sync summary text back to allFields
	if f.summaryIdx >= 0 {
		text := strings.TrimSpace(string(f.summaryText))
		f.allFields[f.summaryIdx].Value = text
		f.allFields[f.summaryIdx].DisplayValue = text
	}
	if f.descIdx >= 0 && f.desc != nil {
		text := strings.TrimSpace(f.desc.Value())
		f.allFields[f.descIdx].DisplayValue = text
		f.allFields[f.descIdx].Value = nil
		if text != "" {
			f.allFields[f.descIdx].Value = text
		}
	}

	// validate required fields
	hasErrors := false
	for i := range f.allFields {
		if f.allFields[i].Required && f.allFields[i].DisplayValue == "" && f.allFields[i].Value == nil {
			f.allFields[i].HasError = true
			hasErrors = true
		}
	}
	if hasErrors {
		var names []string
		for _, fld := range f.allFields {
			if fld.HasError {
				name := fld.Name
				if name == "" {
					name = fld.FieldID
				}
				names = append(names, name)
			}
		}
		f.errorMsg = "Required field(s) empty: " + strings.Join(names, ", ")
		return nil, true
	}

	fieldsMap := make(map[string]any)
	for _, fld := range f.allFields {
		if fld.Value != nil {
			fieldsMap[fld.FieldID] = fld.Value
		}
	}
	return func() tea.Msg { return CreateFormSubmitMsg{Fields: fieldsMap} }, true
}

// Layout

const (
	panelMinH     = 3 // 1 content line + 2 borders
	fieldsColMinW = 28
	fieldsColMaxW = 48
	descColMinW   = 40
)

type formRect struct{ x, y, w, h int }

func (r formRect) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}

// createLayout places the panels in a w×h screen. Wide screens put Fields in
// a left column beside Summary over Description; narrow ones stack Summary,
// Description, and Fields.
type createLayout struct {
	split                 bool
	summary, desc, fields formRect
}

func (f *CreateForm) layout(w, h int) createLayout {
	availH := h - 1 // help bar
	if f.errorMsg != "" {
		availH--
	}
	availH = max(availH, 2*panelMinH)

	if w >= fieldsColMinW+descColMinW {
		fieldsW := min(max(w*3/10, fieldsColMinW), fieldsColMaxW)
		rightW := w - fieldsW
		return createLayout{
			split:   true,
			fields:  formRect{0, 0, fieldsW, availH},
			summary: formRect{fieldsW, 0, rightW, panelMinH},
			desc:    formRect{fieldsW, panelMinH, rightW, availH - panelMinH},
		}
	}

	rest := availH - panelMinH
	fieldsNat := max(len(f.filteredFields())+2, panelMinH)
	fieldsH := min(fieldsNat, max(rest/2, panelMinH))
	descH := max(rest-fieldsH, panelMinH)
	return createLayout{
		summary: formRect{0, 0, w, panelMinH},
		desc:    formRect{0, panelMinH, w, descH},
		fields:  formRect{0, panelMinH + descH, w, fieldsH},
	}
}

func (f *CreateForm) sizeDesc(w, h int) {
	if f.desc == nil {
		return
	}
	r := f.layout(w, h).desc
	f.desc.SetWidth(max(r.w-2, 1))
	f.desc.SetHeight(max(r.h-2, 1))
}

// Render

func (f *CreateForm) Render(bg string, w, h int) string {
	if !f.visible {
		return bg
	}
	if f.loading && len(f.allFields) == 0 {
		popup := RenderPanelFull("Create issue", "", "\n  Loading...\n", 40, 3, true, nil)
		return Overlay(bg, popup, w, h)
	}
	return f.renderForm(bg, w, h)
}

func (f *CreateForm) renderForm(bg string, w, h int) string {
	f.sizeDesc(w, h)
	l := f.layout(w, h)

	summaryPanel := f.renderSummary(l.summary.w, l.summary.h)
	descPanel := f.renderDescription(l.desc.w, l.desc.h)
	fieldsPanel := f.renderFields(l.fields.w, l.fields.h)

	var combined string
	if l.split {
		right := lipgloss.JoinVertical(lipgloss.Left, summaryPanel, descPanel)
		combined = lipgloss.JoinHorizontal(lipgloss.Top, fieldsPanel, right)
	} else {
		combined = lipgloss.JoinVertical(lipgloss.Left, summaryPanel, descPanel, fieldsPanel)
	}

	if f.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(theme.ColorRed)
		errLine := errStyle.Render(TruncateEnd(" "+f.errorMsg, w))
		if lw := lipgloss.Width(errLine); lw < w {
			errLine += strings.Repeat(" ", w-lw)
		}
		combined = lipgloss.JoinVertical(lipgloss.Left, combined, errLine)
	}

	return OverlayAt(bg, combined, 0, 0, w, h)
}

func (f *CreateForm) renderSummary(formW, panelH int) string {
	focused := f.focusedPanel == CreatePanelSummary
	innerW := max(formW-2, 1)
	innerH := max(panelH-2, 1)

	title := "Summary"
	if f.summaryIdx >= 0 && f.allFields[f.summaryIdx].Required {
		title = "*Summary"
	}

	if focused {
		content := f.renderSummaryWithCursor(innerW, innerH)
		return RenderPanelFull(title, "", content, formW, innerH, true, nil)
	}

	// not focused: render same layout as cursor mode but without cursor
	content := f.renderSummaryPlain(innerW, innerH)
	return RenderPanelFull(title, "", content, formW, innerH, false, nil)
}

func (f *CreateForm) renderSummaryWithCursor(innerW, innerH int) string {
	cursorStyle := lipgloss.NewStyle().Foreground(theme.ColorCyan)
	allRunes := append([]rune{' '}, f.summaryText...)
	cursorPos := f.summaryCursor + 1 // +1 for leading space

	// wrap runes into display lines
	type wLine struct {
		runes []rune
		start int
	}
	var wrapped []wLine
	off := 0
	for off < len(allRunes) {
		cut := 0
		w := 0
		for i := off; i < len(allRunes); i++ {
			rw := lipgloss.Width(string(allRunes[i]))
			if w+rw > innerW {
				break
			}
			w += rw
			cut = i + 1
		}
		if cut <= off {
			cut = off + 1
		}
		wrapped = append(wrapped, wLine{runes: allRunes[off:cut], start: off})
		off = cut
	}
	if len(wrapped) == 0 {
		wrapped = append(wrapped, wLine{})
	}

	// find which line the cursor is on and auto-scroll to keep it visible
	cursorLine := 0
	for li, wl := range wrapped {
		lineEnd := wl.start + len(wl.runes)
		if cursorPos < lineEnd || (cursorPos == lineEnd && cursorPos >= len(allRunes)) {
			cursorLine = li
			break
		}
	}
	viewStart := 0
	if cursorLine >= innerH {
		viewStart = cursorLine - innerH + 1
	}

	var lines []string
	cursorPlaced := false
	for li := viewStart; li < len(wrapped) && len(lines) < innerH; li++ {
		wl := wrapped[li]
		lineEnd := wl.start + len(wl.runes)
		lineW := lipgloss.Width(string(wl.runes))

		if !cursorPlaced && (cursorPos < lineEnd || (cursorPos == lineEnd && cursorPos >= len(allRunes))) {
			col := cursorPos - wl.start
			switch {
			case col >= len(wl.runes) && lineW >= innerW:
				// cursor past end of a full line: put cursor on next line
				lines = append(lines, string(wl.runes))
				if len(lines) < innerH {
					lines = append(lines, cursorStyle.Render("█")+strings.Repeat(" ", max(innerW-1, 0)))
				}
			case col >= len(wl.runes):
				// cursor past end of a short line: append cursor block
				rendered := string(wl.runes) + cursorStyle.Render("█")
				lines = append(lines, rendered)
			default:
				before := string(wl.runes[:col])
				at := string(wl.runes[col : col+1])
				after := string(wl.runes[col+1:])
				lines = append(lines, before+cursorStyle.Render(at)+after)
			}
			cursorPlaced = true
		} else {
			lines = append(lines, string(wl.runes))
		}
	}

	return strings.Join(lines, "\n")
}

// renderSummaryPlain uses the same wrapping as cursor mode but without cursor styling
func (f *CreateForm) renderSummaryPlain(innerW, innerH int) string {
	allRunes := append([]rune{' '}, f.summaryText...)
	if len(allRunes) <= 1 {
		return ""
	}

	var lines []string
	off := 0
	for off < len(allRunes) {
		cut := 0
		w := 0
		for i := off; i < len(allRunes); i++ {
			rw := lipgloss.Width(string(allRunes[i]))
			if w+rw > innerW {
				break
			}
			w += rw
			cut = i + 1
		}
		if cut <= off {
			cut = off + 1
		}
		lines = append(lines, string(allRunes[off:cut]))
		off = cut
	}

	if len(lines) > innerH {
		lines = lines[:innerH]
	}
	return strings.Join(lines, "\n")
}

func (f *CreateForm) renderDescription(panelW, panelH int) string {
	focused := f.focusedPanel == CreatePanelDescription
	innerH := max(panelH-2, 1)
	content := ""
	if f.desc != nil {
		content = f.desc.View()
	}
	return RenderPanelFull("Description", "", content, panelW, innerH, focused, nil)
}

func (f *CreateForm) renderFields(formW, panelH int) string {
	focused := f.focusedPanel == CreatePanelFields
	innerW := max(formW-2, 1)
	innerH := max(panelH-2, 1)

	filtered := f.filteredFields()

	// clamp cursor
	if f.fieldCursor >= len(filtered) {
		f.fieldCursor = max(len(filtered)-1, 0)
	}

	// adjust scroll offset
	if f.fieldCursor < f.fieldOffset {
		f.fieldOffset = f.fieldCursor
	}
	if f.fieldCursor >= f.fieldOffset+innerH {
		f.fieldOffset = f.fieldCursor - innerH + 1
	}
	maxOffset := max(len(filtered)-innerH, 0)
	if f.fieldOffset > maxOffset {
		f.fieldOffset = maxOffset
	}

	selStyle := lipgloss.NewStyle().Background(theme.ColorHighlight)
	errStyle := lipgloss.NewStyle().Foreground(theme.ColorRed)
	reqMark := lipgloss.NewStyle().Foreground(theme.ColorRed).Bold(true).Render("*")

	// label column = longest field name + 1 space gap (+ 1 for req mark / leading space)
	labelW := 0
	for _, idx := range filtered {
		w := lipgloss.Width(f.allFields[idx].Name)
		if w > labelW {
			labelW = w
		}
	}
	labelW += 2 // leading marker + trailing space

	// cap label column to half the inner width so values always have room
	maxLabelW := innerW / 2
	if labelW > maxLabelW {
		labelW = maxLabelW
	}

	end := min(f.fieldOffset+innerH, len(filtered))
	var lines []string
	for ci := f.fieldOffset; ci < end; ci++ {
		idx := filtered[ci]
		fld := f.allFields[idx]

		label := fld.Name
		if fld.Required {
			label = reqMark + label
		} else {
			label = " " + label
		}
		// truncate long labels
		if lipgloss.Width(label) > labelW {
			label = TruncateEnd(label, labelW)
		}
		for lipgloss.Width(label) < labelW {
			label += " "
		}

		val := fld.DisplayValue
		maxVal := innerW - labelW - 1
		if maxVal > 0 && lipgloss.Width(val) > maxVal {
			val = TruncateEnd(val, maxVal)
		}

		var line string
		if focused && ci == f.fieldCursor {
			plain := " " + label + val
			for lipgloss.Width(plain) < innerW {
				plain += " "
			}
			line = selStyle.Render(plain)
		} else {
			if val != "" {
				val = styleFieldValue(fld, val)
			}
			line = " " + label + val
			if fld.HasError {
				line = " " + errStyle.Render(label) + val
			}
		}

		lines = append(lines, line)
	}

	for len(lines) < innerH {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	footer := ""
	if len(filtered) > 0 {
		footer = fmt.Sprintf("%d of %d", f.fieldCursor+1, len(filtered))
	}

	var scroll *ScrollInfo
	if len(filtered) > innerH {
		scroll = &ScrollInfo{
			Total:   len(filtered),
			Visible: innerH,
			Offset:  f.fieldOffset,
		}
	}

	title := "Fields"
	return RenderPanelFull(title, footer, content, formW, innerH, focused, scroll)
}

func noneStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.ColorGray)
}

func styleFieldValue(fld CreateFormField, val string) string {
	if val == "None" {
		return noneStyle().Render(val)
	}
	switch fld.FieldID {
	case "priority":
		return theme.PriorityStyled(val)
	default:
		if fld.Type == CFFieldPerson || fld.SchemaItems == "user" {
			return theme.AuthorRender(val)
		}
		return val
	}
}
