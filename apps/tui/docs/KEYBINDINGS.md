# Pryx TUI - Comprehensive Keyboard Shortcuts

## Application Shortcuts

| Key         | Action                       |
| ----------- | ---------------------------- |
| `/`         | Open command palette         |
| `?`         | Show keyboard shortcuts help |
| `Esc`       | Close palette/help/cancel    |
| `Tab`       | Switch to next view          |
| `Shift+Tab` | Switch to previous view      |
| `Ctrl+C`    | Quit application             |
| `Ctrl+L`    | Clear screen / Refresh       |

## View Navigation

| Key | View     |
| --- | -------- |
| `1` | Chat     |
| `2` | Sessions |
| `3` | Channels |
| `4` | Skills   |
| `5` | Settings |

## Command Palette (when open)

| Key       | Action                     |
| --------- | -------------------------- |
| `↑` / `↓` | Navigate commands          |
| `Enter`   | Select highlighted command |
| `1-9`     | Quick select by number     |
| `Esc`     | Close palette              |
| `c`       | Chat view                  |
| `s`       | Sessions view              |
| `n`       | Channels view              |
| `,`       | Settings view              |
| `k`       | Skills view                |
| `q`       | Quit                       |
| `?`       | Help                       |

## Chat Input

| Key               | Action                         |
| ----------------- | ------------------------------ |
| `Enter`           | Send message                   |
| `↑`               | Previous message in history    |
| `↓`               | Next message in history        |
| `←` / `→`         | Move cursor                    |
| `Home` / `Ctrl+A` | Go to start of line            |
| `End` / `Ctrl+E`  | Go to end of line              |
| `Backspace`       | Delete character before cursor |
| `Delete`          | Delete character after cursor  |
| `Ctrl+K`          | Clear from cursor to end       |
| `Ctrl+U`          | Clear from cursor to start     |
| `Ctrl+W`          | Delete word before cursor      |
| `Ctrl+Y`          | Paste (yank)                   |
| `Ctrl+D`          | Delete character under cursor  |

## Scrolling (in message views)

| Key         | Action               |
| ----------- | -------------------- |
| `Page Up`   | Scroll up one page   |
| `Page Down` | Scroll down one page |
| `Ctrl+↑`    | Scroll up one line   |
| `Ctrl+↓`    | Scroll down one line |

## Copy & Paste

The TUI supports standard terminal copy/paste:

- **macOS**: `Cmd+C` / `Cmd+V`
- **Linux/Windows**: `Ctrl+Shift+C` / `Ctrl+Shift+V`
- Or use terminal's native paste (often `Ctrl+V` or right-click)

## Help Screen

Press `?` at any time to see the full keyboard shortcuts reference.

## Implementation Details

The TUI now uses:

- OpenTUI's `<input>` component for proper text editing
- ANSI escape sequence parsing for special keys
- Comprehensive keybinding registry in `src/lib/keybindings.ts`
- Keyboard shortcuts help overlay

## Files Changed

1. **src/components/Chat.tsx** - Uses OpenTUI input component
2. **src/components/App.tsx** - Added comprehensive global keybindings
3. **src/components/CommandPalette.tsx** - Added arrow key navigation
4. **src/components/KeyboardShortcuts.tsx** - New help overlay component
5. **src/lib/keybindings.ts** - New keybinding constants and utilities
