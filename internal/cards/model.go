package cards

import (
	"context"
	"path/filepath"
	"sort"

	"github.com/keshon/charapng"
)

type state int

const (
	stateCards state = iota
	stateChooseAction
	stateRenameTemplate
	stateRenameConfirm
	stateApplying
	stateDone
)

type actionType int

const (
	actionBatchRename actionType = iota
)

// Model holds the card list, selection, and current UI state.
type Model struct {
	ctx   context.Context
	dir   string
	files []charapng.CardFile

	state      state
	selected   map[string]bool // key: CardFile.Path
	sortAsc    bool
	currentAction actionType

	// Batch rename
	renameTemplate    string
	renamePreviews    []charapng.RenamePreview
	renameFieldsContent string // static panel content, built once when entering stateRenameTemplate
	renameErr         error
	renameDoneCount   int
}

// NewModel builds a model from scanned card files.
func NewModel(ctx context.Context, dir string, files []charapng.CardFile) Model {
	return Model{
		ctx:     ctx,
		dir:     dir,
		files:   files,
		state:   stateCards,
		selected: make(map[string]bool),
		sortAsc: true,
	}
}

// Files returns the list of card files (sorted by path).
func (m *Model) Files() []charapng.CardFile {
	out := make([]charapng.CardFile, len(m.files))
	copy(out, m.files)
	sort.Slice(out, func(i, j int) bool {
		if m.sortAsc {
			return out[i].Path < out[j].Path
		}
		return out[i].Path > out[j].Path
	})
	return out
}

// ToggleSelection toggles the selection for the given path.
func (m *Model) ToggleSelection(path string) {
	m.selected[path] = !m.selected[path]
}

// SetSelected sets selection for path to v.
func (m *Model) SetSelected(path string, v bool) {
	m.selected[path] = v
}

// IsSelected returns whether path is selected.
func (m *Model) IsSelected(path string) bool {
	return m.selected[path]
}

// SelectedFiles returns only the selected card files (order preserved from Files()).
func (m *Model) SelectedFiles() []charapng.CardFile {
	var out []charapng.CardFile
	for _, f := range m.Files() {
		if m.selected[f.Path] {
			out = append(out, f)
		}
	}
	return out
}

// SelectedCount returns the number of selected files.
func (m *Model) SelectedCount() int {
	n := 0
	for range m.selected {
		n++
	}
	return n
}

// TotalCount returns the total number of card files.
func (m *Model) TotalCount() int {
	return len(m.files)
}

// SelectAllVisible sets selection for all visible (passed) paths to v.
func (m *Model) SelectAllVisible(paths []string, v bool) {
	for _, p := range paths {
		m.selected[p] = v
	}
}

// DisplayName returns a short display string for a card file (filename + character name).
func DisplayName(path string, card *charapng.Card) string {
	name := filepath.Base(path)
	if card != nil && card.Name != "" {
		name = card.Name + " (" + filepath.Base(path) + ")"
	}
	return name
}
