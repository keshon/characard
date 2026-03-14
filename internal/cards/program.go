package cards

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/keshon/charapng"
)

type programModel struct {
	root                  Model
	cardList              list.Model
	actionList            list.Model
	templateInput         textinput.Model
	previewViewport       viewport.Model
	fieldsViewport        viewport.Model
	renameViewportsInited bool
	focusedRenamePanel    int // 0=template input, 1=preview viewport, 2=fields viewport
	width                 int
	height                int
	selectedScrollOffset  int
	filterMode            int // 0=by filename/name, 1=by tags
}

const (
	filterModeName = 0
	filterModeTags = 1
)

type renameDoneMsg struct {
	err   error
	count int
}

type cardItem struct {
	path      string
	card      *charapng.Card
	filterMode int // 0=by name, 1=by tags (used in FilterValue)
}

func (i cardItem) Title() string {
	return DisplayName(i.path, i.card)
}

func (i cardItem) Description() string {
	if i.card != nil && i.card.Creator != "" {
		return i.card.Creator
	}
	return filepath.Base(i.path)
}

func (i cardItem) FilterValue() string {
	var s string
	switch i.filterMode {
	case filterModeTags:
		if i.card != nil && len(i.card.Tags) > 0 {
			s = strings.Join(i.card.Tags, " ")
		} else {
			s = filepath.Base(i.path)
		}
	default:
		// filterModeName: path + character name
		s = i.path
		if i.card != nil && i.card.Name != "" {
			s = s + " " + i.card.Name
		}
	}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		out = append(out, unicode.ToLower(r))
	}
	return string(out)
}

type actionItem struct {
	action actionType
	name   string
	desc   string
}

func (i actionItem) Title() string       { return i.name }
func (i actionItem) Description() string { return i.desc }
func (i actionItem) FilterValue() string { return i.name }

// filterByContains returns only items whose FilterValue contains the term (case-insensitive).
func filterByContains(term string, targets []string) []list.Rank {
	term = strings.ToLower(term)
	var out []list.Rank
	for i, t := range targets {
		if strings.Contains(strings.ToLower(t), term) {
			out = append(out, list.Rank{Index: i})
		}
	}
	return out
}

func buildCardItems(files []charapng.CardFile, filterMode int) []list.Item {
	items := make([]list.Item, len(files))
	for i, f := range files {
		items[i] = cardItem{path: f.Path, card: f.Card, filterMode: filterMode}
	}
	return items
}

// wrapLine breaks line at or before maxWidth, continuing with indent on the next line.
// If maxWidth <= 0, returns line unchanged.
func wrapLine(line string, maxWidth int, indent string) string {
	if maxWidth <= 0 || len(line) <= maxWidth {
		return line
	}
	var b strings.Builder
	remaining := line
	for len(remaining) > maxWidth {
		// Find last comma or space before maxWidth to break at
		chunk := remaining[:maxWidth]
		idx := strings.LastIndexAny(chunk, ", ")
		if idx <= 0 {
			idx = len(chunk)
		} else {
			idx++ // include the comma/space
		}
		b.WriteString(remaining[:idx])
		b.WriteString("\n")
		b.WriteString(indent)
		remaining = strings.TrimLeft(remaining[idx:], " ")
	}
	if len(remaining) > 0 {
		b.WriteString(remaining)
	}
	return b.String()
}

// buildFieldsPanelContent returns a static description of template fields and their values
// across the given card files. Used for the rename screen fields viewport.
// maxLineWidth: wrap long lines with indent (e.g. 2 spaces); 0 = no wrap.
func buildFieldsPanelContent(files []charapng.CardFile, maxLineWidth int) string {
	type fieldDef struct {
		key  string
		desc string
	}
	maxTag := 0
	for _, f := range files {
		if f.Card != nil && len(f.Card.Tags) > maxTag {
			maxTag = len(f.Card.Tags)
		}
	}
	var defs []fieldDef
	defs = append(defs, fieldDef{"name", "character name"}, fieldDef{"creator", "creator"}, fieldDef{"spec", "card spec version (v1/v2/v3)"})
	for i := 1; i <= maxTag; i++ {
		defs = append(defs, fieldDef{fmt.Sprintf("tag%d", i), ordinalTag(i) + " tag"})
	}
	defs = append(defs,
		fieldDef{"year", "file modification year"},
		fieldDef{"month", "file modification month"},
		fieldDef{"day", "file modification day"},
		fieldDef{"index", "file index (1-based)"},
	)
	var b strings.Builder
	for _, d := range defs {
		var values []string
		for i, f := range files {
			var v string
			fileTime := time.Now()
			if f.Path != "" {
				if info, err := os.Stat(f.Path); err == nil {
					fileTime = info.ModTime()
				}
			}
			switch d.key {
			case "index":
				v = strconv.Itoa(i + 1)
			case "year":
				v = strconv.Itoa(fileTime.Year())
			case "month":
				v = fmt.Sprintf("%02d", fileTime.Month())
			case "day":
				v = fmt.Sprintf("%02d", fileTime.Day())
			}
			if v == "" && f.Card != nil {
				switch d.key {
				case "name":
					v = f.Card.Name
				case "creator":
					v = f.Card.Creator
				case "spec":
					v = f.Card.SpecVersion
				default:
					if strings.HasPrefix(d.key, "tag") {
						n, _ := strconv.Atoi(strings.TrimPrefix(d.key, "tag"))
						if n >= 1 && n <= len(f.Card.Tags) {
							v = f.Card.Tags[n-1]
						}
					}
				}
			}
			if v != "" {
				values = append(values, v)
			}
		}
		line := fmt.Sprintf("[%s] — %s:", d.key, d.desc)
		if len(values) > 0 {
			line += " " + strings.Join(values, ", ")
		} else {
			line += " (empty)"
		}
		line = wrapLine(line, maxLineWidth, "  ")
		b.WriteString(line)
		b.WriteString("\n\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func ordinalTag(n int) string {
	switch n {
	case 1:
		return "first"
	case 2:
		return "second"
	case 3:
		return "third"
	default:
		return fmt.Sprintf("%dth", n)
	}
}

func buildPreviewLines(previews []charapng.RenamePreview) string {
	var b strings.Builder
	for _, pr := range previews {
		b.WriteString(filepath.Base(pr.Path))
		b.WriteString("  →  ")
		b.WriteString(newNameStyle.Render(pr.NewName))
		b.WriteString("\n")
	}
	if len(previews) == 0 {
		b.WriteString("(enter a template to see preview)")
	}
	return b.String()
}

func newProgramModel(m Model) programModel {
	files := m.Files()
	pm := programModel{root: m, filterMode: filterModeName}
	items := buildCardItems(files, pm.filterMode)

	l := list.New(items, cardDelegate{selected: pm.root.selected}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowFilter(true)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Filter = filterByContains
	pm.cardList = l

	actionItems := []list.Item{
		actionItem{action: actionBatchRename, name: "Batch Rename", desc: "Rename selected files by template ([name], [tag1], [year], ...)"},
	}
	al := list.New(actionItems, list.NewDefaultDelegate(), 0, 0)
	al.SetShowTitle(false)
	al.SetShowFilter(false)
	al.SetShowHelp(false)
	al.SetShowStatusBar(false)
	pm.actionList = al

	ti := textinput.New()
	ti.Placeholder = "e.g. [name]-[tag1]-[year].png"
	ti.CharLimit = 256
	ti.Width = 50
	pm.templateInput = ti

	return pm
}

func (p programModel) Init() tea.Cmd { return nil }

func (p programModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.cardList.SetSize(msg.Width, msg.Height-4)
		listHeight := msg.Height - 18
		if listHeight < 5 {
			listHeight = 5
		}
		p.actionList.SetSize(msg.Width, listHeight)
		if p.root.state == stateRenameTemplate && p.renameViewportsInited {
			previewH := (p.height - 6) / 2
			if previewH < 3 {
				previewH = 3
			}
			fieldsH := p.height - 6 - previewH
			if fieldsH < 3 {
				fieldsH = 3
			}
			p.previewViewport.Width = p.width - 2
			p.previewViewport.Height = previewH
			p.fieldsViewport.Width = p.width - 2
			p.fieldsViewport.Height = fieldsH
		}
		return p, nil
	}

	switch p.root.state {
	case stateCards:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			key := msg.String()
			filterState := p.cardList.FilterState()
			isFiltering := filterState == list.Filtering
			if isFiltering && key == "tab" {
				// Toggle filter mode: name <-> tags
				p.filterMode = (p.filterMode + 1) % 2
				p.cardList.SetItems(buildCardItems(p.root.Files(), p.filterMode))
				return p, nil
			}
			if !isFiltering {
				switch key {
				case "q", "ctrl+c":
					return p, tea.Quit
				case " ", "insert":
					p.toggleSelectedVisible()
					return p, nil
				case "+":
					p.toggleSelectedAll(true)
					return p, nil
				case "-":
					p.toggleSelectedAll(false)
					return p, nil
				case "t":
					p.sortSelectedTop()
					return p, nil
				case "enter":
					if len(p.root.SelectedFiles()) > 0 {
						p.root.state = stateChooseAction
					}
					return p, nil
				}
			}
		}
		var cmd tea.Cmd
		p.cardList, cmd = p.cardList.Update(msg)
		return p, cmd

	case stateChooseAction:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				p.root.state = stateCards
				return p, nil
			case "left":
				if p.selectedScrollOffset > 0 {
					p.selectedScrollOffset--
				}
				return p, nil
			case "right":
				maxOffset := len(p.root.SelectedFiles()) - p.maxVisibleSelected()
				if maxOffset < 0 {
					maxOffset = 0
				}
				if p.selectedScrollOffset < maxOffset {
					p.selectedScrollOffset++
				}
				return p, nil
			case "enter":
				if i, ok := p.actionList.SelectedItem().(actionItem); ok {
					p.root.currentAction = i.action
					if i.action == actionBatchRename {
						p.root.renameTemplate = ""
						p.root.renamePreviews = nil
						p.root.renameErr = nil
						maxW := p.width - 4
						if maxW < 40 {
							maxW = 40
						}
						p.root.renameFieldsContent = buildFieldsPanelContent(p.root.SelectedFiles(), maxW)
						p.templateInput.SetValue("")
						p.templateInput.Focus()
						p.focusedRenamePanel = 0
						p.root.state = stateRenameTemplate
						return p, textinput.Blink
					}
				}
				return p, nil
			}
		}
		var cmd tea.Cmd
		p.actionList, cmd = p.actionList.Update(msg)
		return p, cmd

	case stateRenameTemplate:
		// Init viewports on first run in this state when we have dimensions
		if !p.renameViewportsInited && p.height > 10 {
			previewH := (p.height - 6) / 2
			if previewH < 3 {
				previewH = 3
			}
			fieldsH := p.height - 6 - previewH
			if fieldsH < 3 {
				fieldsH = 3
			}
			p.previewViewport = viewport.New(p.width-2, previewH)
			p.previewViewport.SetContent(buildPreviewLines(p.root.renamePreviews))
			p.fieldsViewport = viewport.New(p.width-2, fieldsH)
			p.fieldsViewport.SetContent(p.root.renameFieldsContent)
			p.renameViewportsInited = true
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				p.root.state = stateChooseAction
				p.templateInput.Blur()
				p.renameViewportsInited = false
				return p, nil
			case "tab":
				p.focusedRenamePanel = (p.focusedRenamePanel + 1) % 3
				if p.focusedRenamePanel == 0 {
					p.templateInput.Focus()
				} else {
					p.templateInput.Blur()
				}
				return p, nil
			case "shift+tab":
				p.focusedRenamePanel--
				if p.focusedRenamePanel < 0 {
					p.focusedRenamePanel = 2
				}
				if p.focusedRenamePanel == 0 {
					p.templateInput.Focus()
				} else {
					p.templateInput.Blur()
				}
				return p, nil
			case "enter":
				if p.focusedRenamePanel != 0 {
					return p, nil
				}
				tpl := strings.TrimSpace(p.templateInput.Value())
				if tpl == "" {
					return p, nil
				}
				p.root.renameTemplate = tpl
				sel := p.root.SelectedFiles()
				previews, err := charapng.BatchRenamePreview(sel, tpl)
				p.root.renamePreviews = previews
				p.root.renameErr = err
				p.templateInput.Blur()
				p.root.state = stateRenameConfirm
				return p, nil
			case "up", "down", "pageup", "pagedown":
				if p.focusedRenamePanel == 1 {
					var cmd tea.Cmd
					p.previewViewport, cmd = p.previewViewport.Update(msg)
					return p, cmd
				}
				if p.focusedRenamePanel == 2 {
					var cmd tea.Cmd
					p.fieldsViewport, cmd = p.fieldsViewport.Update(msg)
					return p, cmd
				}
			}
		}
		var cmd tea.Cmd
		p.templateInput, cmd = p.templateInput.Update(msg)
		// Live-update preview when template changes
		sel := p.root.SelectedFiles()
		tpl := strings.TrimSpace(p.templateInput.Value())
		if previews, err := charapng.BatchRenamePreview(sel, tpl); err == nil {
			p.root.renamePreviews = previews
		}
		if p.renameViewportsInited {
			p.previewViewport.SetContent(buildPreviewLines(p.root.renamePreviews))
		}
		return p, cmd

	case stateRenameConfirm:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch strings.ToLower(msg.String()) {
			case "y":
				if p.root.renameErr != nil {
					p.root.state = stateChooseAction
					return p, nil
				}
				sel := p.root.SelectedFiles()
				p.root.state = stateApplying
				return p, tea.Cmd(func() tea.Msg {
					err := charapng.BatchRename(sel, p.root.renameTemplate)
					n := 0
					if err == nil {
						n = len(sel)
					}
					return renameDoneMsg{err: err, count: n}
				})
			case "n", "esc":
				p.root.state = stateRenameTemplate
				p.focusedRenamePanel = 0
				p.templateInput.Focus()
				return p, textinput.Blink
			}
		}
		return p, nil

	case stateApplying:
		switch msg := msg.(type) {
		case renameDoneMsg:
			p.root.renameErr = msg.err
			p.root.renameDoneCount = msg.count
			p.root.state = stateDone
			return p, nil
		}
		return p, nil

	case stateDone:
		if _, ok := msg.(tea.KeyMsg); ok {
			return p, tea.Quit
		}
		return p, nil
	}

	return p, nil
}

func (p *programModel) toggleSelectedVisible() {
	visible := p.cardList.VisibleItems()
	idx := p.cardList.Index()
	if idx < 0 || idx >= len(visible) {
		return
	}
	item := visible[idx].(cardItem)
	p.root.ToggleSelection(item.path)
}

func (p *programModel) toggleSelectedAll(v bool) {
	for _, it := range p.cardList.VisibleItems() {
		p.root.SetSelected(it.(cardItem).path, v)
	}
}

func (p *programModel) sortSelectedTop() {
	files := p.root.Files()
	var selected, rest []list.Item
	for _, f := range files {
		it := cardItem{path: f.Path, card: f.Card}
		if p.root.IsSelected(f.Path) {
			selected = append(selected, it)
		} else {
			rest = append(rest, it)
		}
	}
	p.cardList.SetItems(append(selected, rest...))
}

var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	subtitleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	newNameStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	countStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	noMatchStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
)

func (p programModel) maxVisibleSelected() int {
	available := p.height - 15
	if available < 3 {
		available = 3
	}
	if available > 10 {
		available = 10
	}
	return available
}

func (p programModel) renderSelectedList(title string) string {
	sel := p.root.SelectedFiles()
	count := len(sel)
	var b strings.Builder
	b.WriteString(subtitleStyle.Render(title) + " ")
	b.WriteString(countStyle.Render(fmt.Sprintf("[%d files]", count)))
	if count > 0 {
		b.WriteString(subtitleStyle.Render(" (←/→ scroll)"))
	}
	b.WriteString("\n")
	if count == 0 {
		b.WriteString("  ")
		b.WriteString(noMatchStyle.Render("No files selected"))
		b.WriteString("\n")
		return b.String()
	}
	maxVisible := p.maxVisibleSelected()
	start := p.selectedScrollOffset
	end := start + maxVisible
	if end > count {
		end = count
	}
	if start > 0 {
		b.WriteString("  ")
		b.WriteString(subtitleStyle.Render(fmt.Sprintf("↑ %d more above", start)))
		b.WriteString("\n")
	}
	for i := start; i < end; i++ {
		f := sel[i]
		b.WriteString("  • ")
		b.WriteString(selectedItemStyle.Render(DisplayName(f.Path, f.Card)))
		b.WriteString("\n")
	}
	if end < count {
		b.WriteString("  ")
		b.WriteString(subtitleStyle.Render(fmt.Sprintf("↓ %d more below", count-end)))
	}
	return b.String()
}

func (p programModel) View() string {
	switch p.root.state {
	case stateCards:
		selectedCount := p.root.SelectedCount()
		totalCount := p.root.TotalCount()
		filterModeHint := "name"
		if p.filterMode == filterModeTags {
			filterModeHint = "tags"
		}
		var b strings.Builder
		b.WriteString(titleStyle.Render("Character Cards"))
		b.WriteString(" ")
		b.WriteString(countStyle.Render(fmt.Sprintf("[%d/%d selected]", selectedCount, totalCount)))
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render(fmt.Sprintf("/ filter (%s) · Tab: name/tags · Space select · + all · - none · t top · Enter proceed", filterModeHint)))
		b.WriteString("\n\n")
		b.WriteString(p.cardList.View())
		return b.String()

	case stateChooseAction:
		var b strings.Builder
		b.WriteString(titleStyle.Render("Choose Action"))
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("↑/↓ select action | Enter confirm | ESC back"))
		b.WriteString("\n\n")
		b.WriteString(p.renderSelectedList("Selected:"))
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Available actions:") + "\n")
		b.WriteString(p.actionList.View())
		return b.String()

	case stateRenameTemplate:
		var b strings.Builder
		b.WriteString(titleStyle.Render("Batch Rename — Template"))
		b.WriteString("  ")
		b.WriteString(subtitleStyle.Render("TAB switch panel · ↑/↓ scroll · Enter confirm · ESC back"))
		b.WriteString("\n")
		if p.focusedRenamePanel == 0 {
			b.WriteString(highlightStyle.Render("▸ Template: "))
		} else {
			b.WriteString("  Template: ")
		}
		b.WriteString(p.templateInput.View())
		b.WriteString("\n\n")
		if p.renameViewportsInited {
			if p.focusedRenamePanel == 1 {
				b.WriteString(highlightStyle.Render("▸ Live preview:") + "\n")
			} else {
				b.WriteString(subtitleStyle.Render("  Live preview:") + "\n")
			}
			b.WriteString(p.previewViewport.View())
			b.WriteString("\n")
			if p.focusedRenamePanel == 2 {
				b.WriteString(highlightStyle.Render("▸ Template fields (from selected cards):") + "\n")
			} else {
				b.WriteString(subtitleStyle.Render("  Template fields (from selected cards):") + "\n")
			}
			b.WriteString(p.fieldsViewport.View())
		} else {
			b.WriteString(subtitleStyle.Render("Loading..."))
		}
		return b.String()

	case stateRenameConfirm:
		var b strings.Builder
		b.WriteString(titleStyle.Render("Confirm Batch Rename"))
		b.WriteString("\n\n")
		if p.root.renameErr != nil {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: " + p.root.renameErr.Error()))
			b.WriteString("\n\n")
			b.WriteString("[n] back to template")
			return b.String()
		}
		b.WriteString(highlightStyle.Render("Template: ") + p.root.renameTemplate + "\n\n")
		n := len(p.root.renamePreviews)
		b.WriteString(subtitleStyle.Render(fmt.Sprintf("Rename %d file(s) with this template?", n)))
		b.WriteString("\n\n")
		b.WriteString("[y]es / [n]o (ESC to go back)")
		return b.String()

	case stateApplying:
		return titleStyle.Render("Renaming...") + "\n\nPlease wait..."

	case stateDone:
		var b strings.Builder
		b.WriteString(titleStyle.Render("Results") + "\n\n")
		if p.root.renameErr != nil {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "))
			b.WriteString(subtitleStyle.Render(p.root.renameErr.Error()) + "\n\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("Renamed "))
			b.WriteString(highlightStyle.Render(fmt.Sprintf("%d", p.root.renameDoneCount)))
			b.WriteString(" file(s).\n\n")
		}
		b.WriteString("Press any key to exit...")
		return b.String()
	}

	return ""
}

type cardDelegate struct {
	selected map[string]bool
}

func (d cardDelegate) Height() int                               { return 1 }
func (d cardDelegate) Spacing() int                              { return 0 }
func (d cardDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d cardDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i := item.(cardItem)
	checkbox := "[ ]"
	if d.selected[i.path] {
		checkbox = "[x]"
	}
	line := lipgloss.NewStyle().PaddingLeft(2).Render(checkbox + " " + i.Title())
	if index == m.Index() {
		line = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(line)
	}
	fmt.Fprint(w, line)
}
