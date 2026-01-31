import { render } from "@opentui/solid";
import App from "./src/components/App";

// Enable terminal mouse tracking
const enableMouseTracking = () => {
  // X10 mouse tracking (basic clicks)
  process.stdout.write("\u001b[?9h");
  // VT200 mouse tracking (button press and release)
  process.stdout.write("\u001b[?1000h");
  // Button event tracking (all button events)
  process.stdout.write("\u001b[?1002h");
  // SGR mouse mode (extended coordinates)
  process.stdout.write("\u001b[?1006h");
  // Focus events (for detecting when terminal gains/loses focus)
  process.stdout.write("\u001b[?1004h");
};

// Disable terminal mouse tracking
const disableMouseTracking = () => {
  process.stdout.write("\u001b[?1004l");
  process.stdout.write("\u001b[?1006l");
  process.stdout.write("\u001b[?1002l");
  process.stdout.write("\u001b[?1000l");
  process.stdout.write("\u001b[?9l");
  // Reset cursor style
  process.stdout.write("\u001b[0 q");
  // Clear screen
  process.stdout.write("\u001b[2J\u001b[H");
};

process.on("SIGINT", () => {
  disableMouseTracking();
  process.exit(0);
});

process.on("exit", () => {
  disableMouseTracking();
});

try {
  // Enable mouse tracking before rendering
  enableMouseTracking();

  render(() => <App />, {
    targetFps: 60,
    exitOnCtrlC: false,
    useMouse: true,
    enableMouseMovement: true,
  });
} catch (e) {
  disableMouseTracking();
  console.error("Failed to start TUI:", e);
  const fs = require("fs");
  fs.writeFileSync("tui-crash.log", String(e) + "\n" + (e instanceof Error ? e.stack : ""));
  process.exit(1);
}
