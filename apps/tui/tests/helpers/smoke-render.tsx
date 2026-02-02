import { render } from "@opentui/solid";
import App from "../../src/components/App";

// Force TTY for library checks
Object.defineProperty(process.stdout, "isTTY", { value: true });
Object.defineProperty(process.stdout, "columns", { value: 80 });
Object.defineProperty(process.stdout, "rows", { value: 24 });

console.log("Starting Render Script...");

try {
  render(() => <App />);

  // Allow time to render frames
  setTimeout(() => {
    console.log("Render timeout reached. Exiting.");
    process.exit(0);
  }, 3000);
} catch (e) {
  console.error("Render failed:", e);
  process.exit(1);
}
