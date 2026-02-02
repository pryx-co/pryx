# TUI Input Handling Fix Plan

## Current Problems

1. **Arrow keys not working** - No navigation in CommandPalette or Chat history
2. **Mouse not working** - Can't click on UI elements
3. **Copy-paste not working** - Multi-byte input not handled
4. **Limited keybindings** - Only Enter and Backspace work

## Root Cause Analysis

The current implementation uses raw `process.stdin.on("data", ...)` which:

- Only receives single-byte input
- Doesn't parse ANSI escape sequences (ESC[A, ESC[B for arrows)
- Doesn't enable mouse reporting
- Doesn't handle bracketed paste mode

## Solution

### Phase 1: Use OpenTUI's Built-in Input Component

OpenTUI provides an `<input>` element with proper event handling:

```tsx
<input
  value={inputValue()}
  onChange={v => setInputValue(v)}
  onSubmit={() => handleSubmit()}
  placeholder="Type a message..."
/>
```

**Benefits:**

- Built-in arrow key support (left/right cursor movement)
- Built-in copy-paste support
- Built-in line editing (Home, End, Ctrl+A, Ctrl+E)
- Proper focus management

### Phase 2: Add Keyboard Navigation to CommandPalette

Add arrow key handlers to CommandPalette:

```tsx
// In CommandPalette component
onMount(() => {
  const handleKey = (data: Buffer) => {
    const seq = data.toString();

    // Arrow up: ESC[A or ESCOA
    if (seq === "\u001b[A" || seq === "\u001bOA") {
      setSelectedIndex(i => Math.max(0, i - 1));
    }
    // Arrow down: ESC[B or ESCOB
    else if (seq === "\u001b[B" || seq === "\u001bOB") {
      setSelectedIndex(i => Math.min(props.commands.length - 1, i + 1));
    }
    // Enter
    else if (seq === "\r" || seq === "\n") {
      props.commands[selectedIndex()].action();
    }
    // Escape
    else if (seq === "\u001b") {
      props.onClose();
    }
  };

  process.stdin.on("data", handleKey);
  onCleanup(() => process.stdin.off("data", handleKey));
});
```

### Phase 3: Enable Mouse Support

Enable mouse reporting in terminal:

```typescript
// In index.tsx or App.tsx
onMount(() => {
  // Enable mouse tracking
  process.stdout.write("\u001b[?1000h"); // Mouse click tracking
  process.stdout.write("\u001b[?1002h"); // Mouse motion tracking
  process.stdout.write("\u001b[?1015h"); // Mouse protocol
  process.stdout.write("\u001b[?1006h"); // SGR mouse protocol

  onCleanup(() => {
    // Disable mouse tracking
    process.stdout.write("\u001b[?1000l");
    process.stdout.write("\u001b[?1002l");
    process.stdout.write("\u001b[?1015l");
    process.stdout.write("\u001b[?1006l");
  });
});
```

### Phase 4: Chat History Navigation

Add up/down arrow support in Chat to navigate message history:

```tsx
const [history, setHistory] = createSignal<string[]>([]);
const [historyIndex, setHistoryIndex] = createSignal(-1);

const handleKey = (data: Buffer) => {
  const seq = data.toString();

  // Arrow up - previous message
  if (seq === "\u001b[A") {
    const idx = historyIndex();
    if (idx < history().length - 1) {
      const newIdx = idx + 1;
      setHistoryIndex(newIdx);
      setInputValue(history()[newIdx]);
    }
  }
  // Arrow down - next message
  else if (seq === "\u001b[B") {
    const idx = historyIndex();
    if (idx > 0) {
      const newIdx = idx - 1;
      setHistoryIndex(newIdx);
      setInputValue(history()[newIdx]);
    } else if (idx === 0) {
      setHistoryIndex(-1);
      setInputValue("");
    }
  }
};
```

## Implementation Priority

1. **CRITICAL**: Replace manual stdin handling with OpenTUI `<input>` component
2. **HIGH**: Add keyboard navigation to CommandPalette
3. **MEDIUM**: Add chat history navigation
4. **LOW**: Enable mouse support (nice-to-have)

## Testing Checklist

- [ ] Arrow keys move cursor in input
- [ ] Arrow keys navigate CommandPalette
- [ ] Arrow keys navigate chat history
- [ ] Copy-paste works
- [ ] Home/End keys work
- [ ] Ctrl+A/E work (beginning/end of line)
- [ ] Tab switches views
- [ ] Escape closes CommandPalette

## References

- ANSI Escape Sequences: https://en.wikipedia.org/wiki/ANSI_escape_code
- Xterm Mouse Tracking: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-Mouse-Tracking
- OpenTUI Documentation: (check node_modules/@opentui/core)
