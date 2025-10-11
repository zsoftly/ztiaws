# Fuzzy Finder Features

The `ztictl` fuzzy finder (used in `ztictl auth login` for account/role selection) includes modern text editing and mouse support capabilities.

## Keyboard Shortcuts

### Text Editing
- **Ctrl+V** - Paste text from clipboard into search box
- **Ctrl+X** - Cut selected text to clipboard
- **Ctrl+C** - Copy selected text to clipboard
- **Ctrl+Z** - Undo last text edit
- **Backspace/Delete** - Delete selected text (if any) or single character

### Navigation (Text)
- **Ctrl+A** / **Home** - Move cursor to beginning of search box
- **Ctrl+E** / **End** - Move cursor to end of search box
- **Ctrl+B** / **Left Arrow** - Move cursor left one character
- **Ctrl+F** / **Right Arrow** - Move cursor right one character
- **Ctrl+W** - Delete word before cursor (backward kill word)
- **Ctrl+U** - Delete everything before cursor (kill line backward)

### Navigation (List)
- **Up Arrow** / **Ctrl+K** / **Ctrl+P** - Move selection up
- **Down Arrow** / **Ctrl+J** / **Ctrl+N** - Move selection down
- **Page Up** - Scroll up one page
- **Page Down** - Scroll down one page
- **Tab** - Toggle selection (in multi-select mode)

### Actions
- **Enter** - Confirm selection
- **Ctrl+D** / **Ctrl+Q** / **Esc** - Quit/abort the finder

## Mouse Support

### Search Box Interactions
- **Left-click on prompt** - Position cursor at click location in search box
- **Left-click + drag on prompt** - Select text by dragging in search box
- **Double-click on prompt** - Select word under cursor in search box
- **Right-click on prompt** - Paste text from clipboard into search box

### List Interactions
- **Left-click on item** - Select that item in the list
- **Mouse wheel up/down** - Scroll through the item list

## Visual Feedback

- **Selected text** is highlighted with inverted colors for visibility
- **Preview window** shows detailed information about the currently highlighted item
- **Border** surrounds the finder for clear visual boundaries
- **Responsive height** adapts to terminal size (configurable via `ZTICTL_SELECTOR_HEIGHT`)
- **Dynamic width** automatically adjusts based on content length
  - Minimum width: 80 columns (comfortable readability)
  - Maximum width: 160 columns (uses more terminal space when needed)
  - Only takes as much width as needed for your account/role names
  - Considers prompt, borders, preview window, and content in calculation
  - Prevents text truncation by calculating actual content requirements
- **Left-aligned positioning** anchored at bottom-left corner of terminal
  - Extends rightward as content requires
  - Doesn't take over full terminal width unnecessarily
  - Leaves right side of terminal visible for context

## Tips

1. **Fast searching**: Start typing immediately to filter results - no need to click into the search box
2. **Copy account IDs**: Select text with mouse drag, then Ctrl+C to copy account IDs or role names
3. **Paste searches**: Use Ctrl+V to paste account IDs or search terms from clipboard
4. **Quick navigation**: Use Page Up/Down to quickly browse long lists of accounts or roles
5. **Mouse shortcuts**: Right-click to paste is often faster than Ctrl+V

## Error Handling

The fuzzy finder includes panic recovery to handle unexpected errors gracefully. If an unexpected error occurs:

```
❌ An unexpected error occurred in the account selector.
   Please report this issue at: https://github.com/zsoftly/ztiaws/issues
   Error details: [error message]
```

The application will exit cleanly with code 1 instead of crashing with a stack trace.

## Environment Variables

- **ZTICTL_SELECTOR_HEIGHT** - Number of items to display (default: 5, range: 1-20)
  ```bash
  export ZTICTL_SELECTOR_HEIGHT=10  # Show 10 items instead of 5
  ```

## Platform Support

All features are supported on:
- ✅ Linux
- ✅ macOS
- ✅ Windows (Command Prompt, PowerShell, Windows Terminal)

Clipboard operations require:
- Linux: X11 or Wayland (xclip/xsel/wl-clipboard)
- macOS: Built-in (pbcopy/pbpaste)
- Windows: Built-in
