import { render } from "@opentui/solid";
import App from "./src/components/App";
import { appendFileSync } from "fs";
import { homedir } from "os";

// Handle exit cleanly
process.on("SIGINT", () => {
    process.exit(0);
});

try {
    render(() => <App />);
} catch (e) {
    console.error("Failed to start TUI:", e);
    const fs = require('fs');
    fs.writeFileSync('tui-crash.log', String(e) + '\n' + (e instanceof Error ? e.stack : ''));
    process.exit(1);
}
