import { createCliRenderer, BoxRenderable, TextRenderable } from "@opentui/core";

async function main() {
  console.log("Creating renderer...");
  const renderer = await createCliRenderer();
  console.log("Renderer created.");

  const box = new BoxRenderable(renderer, {
    id: "main-box",
    borderStyle: "double",
    width: "100%",
    height: "100%",
    title: "Core Test",
    border: true,
  });

  const text = new TextRenderable(renderer, {
    id: "text-1",
    text: "Hello World from Core!",
  });

  // BoxRenderable uses Yoga layout, so we add children differently usually,
  // but BaseRenderable.add() should work.
  box.add(text);
  renderer.root.add(box);

  console.log("Starting renderer...");
  renderer.start();

  // Exit after 5 seconds
  setTimeout(() => {
    renderer.stop();
    console.log("Renderer stopped.");
    process.exit(0);
  }, 5000);
}

try {
  main();
} catch (e) {
  console.error("Manual test failed:", e);
  process.exit(1);
}
