globalThis.process ??= {}; globalThis.process.env ??= {};
import { e as createComponent, k as renderComponent, r as renderTemplate } from '../chunks/astro/server_VtDy7goY.mjs';
import { $ as $$Layout } from '../chunks/Layout_B2YVfyWe.mjs';
import { j as jsxRuntimeExports } from '../chunks/jsx-runtime_DoH26EBh.mjs';
import { a as reactExports } from '../chunks/_@astro-renderers_DMBOvNaZ.mjs';
export { r as renderers } from '../chunks/_@astro-renderers_DMBOvNaZ.mjs';

function SkillCard({ skill, onToggle }) {
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: {
    border: "1px solid #333",
    borderRadius: 8,
    padding: "1.25rem",
    backgroundColor: "#111",
    display: "flex",
    flexDirection: "column",
    gap: "0.75rem",
    transition: "border-color 0.2s",
    opacity: skill.enabled ? 1 : 0.6
  }, children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: { display: "flex", justifyContent: "space-between", alignItems: "start" }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: { fontSize: "2rem" }, children: skill.emoji }),
      /* @__PURE__ */ jsxRuntimeExports.jsx(
        "button",
        {
          onClick: () => onToggle(skill.id),
          style: {
            padding: "4px 12px",
            borderRadius: "999px",
            border: "none",
            fontSize: "0.75rem",
            fontWeight: "bold",
            cursor: "pointer",
            backgroundColor: skill.enabled ? "rgba(16, 185, 129, 0.2)" : "#374151",
            color: skill.enabled ? "#10b981" : "#9ca3af"
          },
          children: skill.enabled ? "ENABLED" : "DISABLED"
        }
      )
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("h3", { style: { margin: "0 0 0.5rem 0", fontSize: "1.1rem", fontFamily: "monospace" }, children: skill.name }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { style: { margin: 0, fontSize: "0.875rem", color: "#9ca3af", lineHeight: "1.4" }, children: skill.description })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: { marginTop: "auto", paddingTop: "0.75rem", display: "flex", flexDirection: "column", gap: "0.5rem" }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: { display: "flex", gap: "0.5rem", alignItems: "center" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: {
        fontSize: "0.7rem",
        textTransform: "uppercase",
        color: skill.source === "bundled" ? "#60a5fa" : "#c084fc",
        border: "1px solid #333",
        padding: "2px 6px",
        borderRadius: 4,
        fontWeight: "bold"
      }, children: skill.source }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("code", { style: {
        fontSize: "0.7rem",
        color: "#6b7280",
        background: "#000",
        padding: "4px",
        borderRadius: 4,
        overflow: "hidden",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap"
      }, children: skill.path })
    ] })
  ] });
}

function SkillList() {
  const [skills, setSkills] = reactExports.useState([]);
  const [loading, setLoading] = reactExports.useState(true);
  const [error, setError] = reactExports.useState(null);
  reactExports.useEffect(() => {
    const fetchSkills = async () => {
      try {
        const res = await fetch("http://localhost:8080/skills");
        if (!res.ok) throw new Error("Failed to fetch skills");
        const data = await res.json();
        const mapped = (data.skills || []).map((s) => ({
          id: s.id,
          name: s.name,
          description: s.description,
          emoji: s.metadata?.emoji || "🧩",
          // Fallback emoji
          enabled: s.enabled,
          source: s.source,
          path: s.path
        }));
        setSkills(mapped);
      } catch (err) {
        console.error(err);
        setError("Could not load skills from Runtime API. Is pryx-core running?");
      } finally {
        setLoading(false);
      }
    };
    fetchSkills();
  }, []);
  if (loading) return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: { padding: "2rem", color: "#6b7280" }, children: "Loading skills..." });
  if (error) return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: { padding: "2rem", color: "#ef4444" }, children: [
    "Error: ",
    error
  ] });
  const handleToggle = (id) => {
    setSkills((prev) => prev.map(
      (s) => s.id === id ? { ...s, enabled: !s.enabled } : s
    ));
  };
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: { padding: "1rem", fontFamily: "system-ui", maxWidth: "1200px", margin: "0 auto" }, children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("header", { style: { marginBottom: "2rem" }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("h1", { style: { margin: 0, fontSize: "1.5rem", display: "flex", alignItems: "center", gap: "0.5rem" }, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("span", { children: "🛠️" }),
        " Skills Manager"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { style: { color: "#9ca3af", marginTop: "0.5rem" }, children: "Manage the capabilities and tools available to your Pryx agent." })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: {
      display: "grid",
      gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))",
      gap: "1.5rem"
    }, children: skills.map((skill) => /* @__PURE__ */ jsxRuntimeExports.jsx(SkillCard, { skill, onToggle: handleToggle }, skill.id)) })
  ] });
}

const $$Skills = createComponent(($$result, $$props, $$slots) => {
  return renderTemplate`${renderComponent($$result, "Layout", $$Layout, {}, { "default": ($$result2) => renderTemplate` ${renderComponent($$result2, "SkillList", SkillList, { "client:load": true, "client:component-hydration": "load", "client:component-path": "/Users/irfandi/.local/share/opencode/worktree/d686a6dfff4eca54f64908f2ecb63fded89b9888/silent-river/apps/web/src/components/skills/SkillList", "client:component-export": "default" })} ` })}`;
}, "/Users/irfandi/.local/share/opencode/worktree/d686a6dfff4eca54f64908f2ecb63fded89b9888/silent-river/apps/web/src/pages/skills.astro", void 0);

const $$file = "/Users/irfandi/.local/share/opencode/worktree/d686a6dfff4eca54f64908f2ecb63fded89b9888/silent-river/apps/web/src/pages/skills.astro";
const $$url = "/skills";

const _page = /*#__PURE__*/Object.freeze(/*#__PURE__*/Object.defineProperty({
    __proto__: null,
    default: $$Skills,
    file: $$file,
    url: $$url
}, Symbol.toStringTag, { value: 'Module' }));

const page = () => _page;

export { page };
