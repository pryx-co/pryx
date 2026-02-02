# Pryx TUI Theme Update

## Overview

Updated the TUI command palette and theming to follow Moltbot's sophisticated palette system, replacing harsh cyan colors with warm, accessible accent colors.

## Changes Made

### 1. New Theme System (`src/theme.ts`)

Created a comprehensive theme system with:

**Color Palette:**

- `text`: #E8E3D5 (off-white for main text)
- `dim`: #7B7F87 (gray for secondary text)
- `accent`: #F6C453 (warm gold for highlights)
- `accentSoft`: #F2A65A (softer orange for secondary accents)
- `bgPrimary`: #1a1a1a (main background)
- `bgSelected`: #2B2F36 (selected item background - subtle)
- `border`: #3C414B (border color)

**Key Improvements:**

- Selected items now use subtle background (#2B2F36) instead of harsh cyan
- Selected text uses warm gold accent color (#F6C453) for visibility
- Better contrast ratios throughout
- Consistent color usage across components

### 2. Updated CommandPalette (`src/components/CommandPalette.tsx`)

- Removed bright cyan background for selected items
- Selected items now show:
  - Subtle dark background (#2B2F36)
  - Warm gold text color (#F6C453)
- Unselected items use off-white text (#E8E3D5)
- Shortcuts remain dimmed (#7B7F87)
- Border uses muted border color (#3C414B)

### 3. Updated AppHeader (`src/components/AppHeader.tsx`)

- Changed ASCII art color from cyan (#00ffff) to warm gold (#F6C453)
- Updated background to use theme palette
- Subtitle now uses dimmed color

## Visual Comparison

### Before:

- Selected item: Bright cyan background (#00FFFF) with black text
- Harsh contrast, doesn't match dark theme

### After:

- Selected item: Subtle dark background (#2B2F36) with warm gold text (#F6C453)
- Smooth integration with dark theme
- Better readability and aesthetics

## Testing

To test the changes:

1. Build the TUI:

   ```bash
   cd apps/tui
   bun install
   bun run build
   ```

2. Run the TUI:

   ```bash
   bun start
   ```

3. Open the command palette (usually `Cmd/Ctrl+Shift+P` or `?`)

4. Navigate through items to see the new selection styling

## Color Reference (Moltbot Pattern)

The theme follows Moltbot's established patterns:

- **Accent**: Warm gold (#F6C453) - for selected items, highlights
- **Text**: Off-white (#E8E3D5) - for primary text
- **Dim**: Gray (#7B7F87) - for secondary text, shortcuts
- **Background**: Dark (#1a1a1a) - main background
- **Selected BG**: Slightly lighter dark (#2B2F36) - selected item background
- **Border**: Muted gray (#3C414B) - borders and separators

This creates a cohesive, professional appearance that's easy on the eyes.
