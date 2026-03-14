# Character Card TUI Manager

CLI/TUI for PNG character cards (spec v1/v2/v3). Browse cards, filter by name or tags, and batch-rename files using templates.

## Purpose

- **Extract metadata** from PNG character cards (chara/ccv3 chunks).
- **List and filter** cards by filename/character name or by tags.
- **Batch rename** files using placeholders like `[name]`, `[tag1]`, `[year]`, etc.

## Build & Run

```bash
go build -o charcard ./cmd/charcard
./charcard
```

With no arguments, the app starts the **interactive menu**. From there you can open Settings, Cards, Help, or About.

## How to Use

### Interactive menu (default)

```bash
charcard
```

- **Cards** — open the card list (uses the folder from Settings if set).
- **Settings** — set the default folder for PNG character cards.
- **Help** / **About** — in-app help and credits.

### Direct commands

```bash
charcard tui [directory]     # Card list; uses -dir or Settings path
charcard tui -dir ./cards   # Card list in ./cards
charcard settings list      # Show current settings
charcard settings set-cards-dir -dir ./cards
charcard help
charcard about
```

### Card list (TUI)

- **/** — start filter; type to narrow the list (only matching rows stay visible).
- **Tab** (while filtering) — switch filter mode: by **filename/name** or by **tags**.
- **Space** — toggle selection of the current item.
- **+** / **-** — select all visible / clear selection.
- **t** — move selected items to the top of the list.
- **Enter** (with selection) — choose action (e.g. Batch Rename).

### Batch rename

- Enter a template, e.g. `[name]-[tag1]-[year].png`.
- **Template fields:** `[name]`, `[creator]`, `[spec]`, `[tag1]`, `[tag2]`, … `[year]`, `[month]`, `[day]`, `[index]`.
- `[spec]` is the card version (v1/v2/v3); `[year]`/`[month]`/`[day]` use the file’s modification date.
- Live preview and a confirmation step before applying renames.

## Requirements

- Go 1.22+
- Terminal that supports the TUI (Bubble Tea).
