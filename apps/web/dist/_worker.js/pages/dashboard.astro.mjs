globalThis.process ??= {}; globalThis.process.env ??= {};
import { e as createComponent, k as renderComponent, r as renderTemplate } from '../chunks/astro/server_VtDy7goY.mjs';
import { $ as $$Layout } from '../chunks/Layout_B2YVfyWe.mjs';
import { j as jsxRuntimeExports } from '../chunks/jsx-runtime_DoH26EBh.mjs';
import { a as reactExports } from '../chunks/_@astro-renderers_DMBOvNaZ.mjs';
export { r as renderers } from '../chunks/_@astro-renderers_DMBOvNaZ.mjs';

function DeviceCard({ device }) {
  const getStatusColor = (status) => {
    switch (status) {
      case "online":
        return "#10b981";
      case "syncing":
        return "#3b82f6";
      case "offline":
        return "#6b7280";
      default:
        return "#6b7280";
    }
  };
  const getTypeIcon = (type) => {
    switch (type) {
      case "host":
        return "🖥️";
      case "mobile":
        return "📱";
      case "cli":
        return "⌨️";
      case "web":
        return "🌐";
      default:
        return "❓";
    }
  };
  const formatTime = (ts) => {
    const diff = Date.now() - ts;
    if (diff < 6e4) return "Just now";
    if (diff < 36e5) return `${Math.floor(diff / 6e4)}m ago`;
    return new Date(ts).toLocaleTimeString();
  };
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: {
    border: "1px solid #333",
    borderRadius: 8,
    padding: "1rem",
    backgroundColor: "#111",
    display: "flex",
    flexDirection: "column",
    gap: "0.5rem",
    minWidth: "200px"
  }, children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: { display: "flex", justifyContent: "space-between", alignItems: "center" }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: { fontSize: "1.5rem" }, children: getTypeIcon(device.type) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: {
        fontSize: "0.75rem",
        color: getStatusColor(device.status),
        border: `1px solid ${getStatusColor(device.status)}`,
        padding: "2px 6px",
        borderRadius: "12px"
      }, children: device.status.toUpperCase() })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: { fontWeight: "bold" }, children: device.name }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: { fontSize: "0.75rem", color: "#9ca3af" }, children: [
      "ID: ",
      device.id.slice(0, 8),
      "..."
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: { fontSize: "0.75rem", color: "#6b7280" }, children: [
      "seen ",
      formatTime(device.lastSeen)
    ] })
  ] });
}

function DeviceList() {
  const [devices, setDevices] = reactExports.useState([]);
  reactExports.useEffect(() => {
    const mockDevices = [
      { id: "dev-12345678", name: "MacBook Pro", type: "host", status: "online", lastSeen: Date.now() },
      { id: "dev-87654321", name: "iPhone 15", type: "mobile", status: "syncing", lastSeen: Date.now() - 12e4 },
      { id: "dev-cli-001", name: "Dev Server", type: "cli", status: "offline", lastSeen: Date.now() - 864e5 }
    ];
    setDevices(mockDevices);
  }, []);
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("section", { style: { marginBottom: "2rem" }, children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("h2", { style: { fontSize: "1rem", marginBottom: "1rem", display: "flex", alignItems: "center", gap: "0.5rem" }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("span", { children: "☁️" }),
      " Cloud Devices"
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: {
      display: "grid",
      gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))",
      gap: "1rem"
    }, children: [
      devices.map((dev) => /* @__PURE__ */ jsxRuntimeExports.jsx(DeviceCard, { device: dev }, dev.id)),
      devices.length === 0 && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: { color: "#6b7280", fontStyle: "italic" }, children: "No devices found." })
    ] })
  ] });
}

function Dashboard() {
  const [events, setEvents] = reactExports.useState([]);
  const [stats, setStats] = reactExports.useState(null);
  const [selectedEvent, setSelectedEvent] = reactExports.useState(null);
  const [connected, setConnected] = reactExports.useState(false);
  reactExports.useEffect(() => {
    const ws = new WebSocket("ws://localhost:3000/ws");
    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);
    ws.onmessage = (msg) => {
      try {
        const evt = JSON.parse(msg.data);
        if (evt.event === "trace.event") {
          setEvents((prev) => [...prev.slice(-99), evt.payload]);
        } else if (evt.event === "session.stats") {
          setStats(evt.payload);
        }
      } catch {
      }
    };
    return () => ws.close();
  }, []);
  const formatDuration = (ms) => {
    if (!ms) return "-";
    if (ms < 1e3) return `${ms}ms`;
    return `${(ms / 1e3).toFixed(2)}s`;
  };
  const formatCost = (cost) => `$${cost.toFixed(4)}`;
  const getEventColor = (event) => {
    if (event.status === "error") return "#ef4444";
    if (event.status === "running") return "#f59e0b";
    switch (event.type) {
      case "tool_call":
        return "#3b82f6";
      case "approval":
        return "#8b5cf6";
      case "message":
        return "#10b981";
      default:
        return "#6b7280";
    }
  };
  const timelineWidth = 600;
  const now = Date.now();
  const timeWindow = 6e4;
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: { padding: "1rem", fontFamily: "system-ui", maxWidth: "1200px", margin: "0 auto" }, children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("header", { style: { display: "flex", justifyContent: "space-between", marginBottom: "2rem" }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: { display: "flex", alignItems: "center", gap: "1.5rem" }, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("h1", { style: { margin: 0, fontSize: "1.5rem", display: "flex", alignItems: "center", gap: "0.5rem" }, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("span", { children: "⚡" }),
          " Pryx Cloud"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("nav", { style: { display: "flex", gap: "1rem", fontSize: "0.9rem" }, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("a", { href: "/dashboard", style: { color: "#fff", textDecoration: "none", fontWeight: "bold" }, children: "Dashboard" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("a", { href: "/skills", style: { color: "#9ca3af", textDecoration: "none" }, children: "Skills" })
        ] })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { style: {
        color: connected ? "#10b981" : "#ef4444",
        display: "flex",
        alignItems: "center",
        gap: "0.5rem"
      }, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: {
          width: 8,
          height: 8,
          borderRadius: "50%",
          backgroundColor: connected ? "#10b981" : "#ef4444"
        } }),
        connected ? "Live" : "Offline"
      ] })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx(DeviceList, {}),
    stats && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: {
      display: "grid",
      gridTemplateColumns: "repeat(4, 1fr)",
      gap: "1rem",
      marginBottom: "1.5rem"
    }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx(StatCard, { label: "Cost", value: formatCost(stats.cost), color: "#f59e0b" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx(StatCard, { label: "Tokens", value: stats.tokens.toLocaleString(), color: "#3b82f6" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx(StatCard, { label: "Duration", value: formatDuration(stats.duration), color: "#10b981" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx(StatCard, { label: "Events", value: stats.eventCount.toString(), color: "#8b5cf6" })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("section", { children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("h2", { style: { fontSize: "1rem", marginBottom: "0.5rem" }, children: "Trace Timeline" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: {
        border: "1px solid #333",
        borderRadius: 8,
        padding: "1rem",
        backgroundColor: "#111"
      }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("svg", { width: timelineWidth, height: Math.max(events.length * 24 + 20, 100), children: events.map((event, i) => {
        const start = Math.max(0, (event.startTime - (now - timeWindow)) / timeWindow);
        const end = event.endTime ? Math.min(1, (event.endTime - (now - timeWindow)) / timeWindow) : 1;
        const x = start * timelineWidth;
        const width = Math.max(4, (end - start) * timelineWidth);
        return /* @__PURE__ */ jsxRuntimeExports.jsxs("g", { onClick: () => setSelectedEvent(event), children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "rect",
            {
              x,
              y: i * 24 + 4,
              width,
              height: 18,
              rx: 4,
              fill: getEventColor(event),
              style: { cursor: "pointer" }
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "text",
            {
              x: x + 4,
              y: i * 24 + 16,
              fontSize: 10,
              fill: "#fff",
              children: event.name.slice(0, 20)
            }
          )
        ] }, event.id);
      }) }) })
    ] }),
    selectedEvent && /* @__PURE__ */ jsxRuntimeExports.jsxs("section", { style: { marginTop: "1rem" }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("h2", { style: { fontSize: "1rem", marginBottom: "0.5rem" }, children: "Event Details" }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: {
        border: "1px solid #333",
        borderRadius: 8,
        padding: "1rem",
        backgroundColor: "#111",
        fontSize: "0.875rem"
      }, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("p", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "ID:" }),
          " ",
          selectedEvent.id
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("p", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Type:" }),
          " ",
          selectedEvent.type
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("p", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Name:" }),
          " ",
          selectedEvent.name
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("p", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Status:" }),
          " ",
          selectedEvent.status
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("p", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Duration:" }),
          " ",
          formatDuration(selectedEvent.duration)
        ] }),
        selectedEvent.correlationId && /* @__PURE__ */ jsxRuntimeExports.jsxs("p", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Correlation ID:" }),
          " ",
          selectedEvent.correlationId
        ] }),
        selectedEvent.error && /* @__PURE__ */ jsxRuntimeExports.jsxs("p", { style: { color: "#ef4444" }, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Error:" }),
          " ",
          selectedEvent.error
        ] })
      ] })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("section", { style: { marginTop: "1rem" }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("h2", { style: { fontSize: "1rem", marginBottom: "0.5rem" }, children: "Recent Events" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: {
        border: "1px solid #333",
        borderRadius: 8,
        overflow: "hidden",
        backgroundColor: "#111"
      }, children: events.slice(-10).reverse().map((event) => /* @__PURE__ */ jsxRuntimeExports.jsxs(
        "div",
        {
          style: {
            padding: "0.5rem 1rem",
            borderBottom: "1px solid #222",
            display: "flex",
            alignItems: "center",
            gap: "0.5rem",
            fontSize: "0.875rem"
          },
          children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: {
              width: 8,
              height: 8,
              borderRadius: "50%",
              backgroundColor: getEventColor(event)
            } }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: { color: "#9ca3af" }, children: event.type }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("span", { children: event.name }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: { marginLeft: "auto", color: "#6b7280" }, children: formatDuration(event.duration) })
          ]
        },
        event.id
      )) })
    ] })
  ] });
}
function StatCard({ label, value, color }) {
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: {
    border: "1px solid #333",
    borderRadius: 8,
    padding: "1rem",
    backgroundColor: "#111"
  }, children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: { fontSize: "0.75rem", color: "#9ca3af", marginBottom: "0.25rem" }, children: label }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: { fontSize: "1.5rem", fontWeight: "bold", color }, children: value })
  ] });
}

const $$Dashboard = createComponent(($$result, $$props, $$slots) => {
  return renderTemplate`${renderComponent($$result, "Layout", $$Layout, {}, { "default": ($$result2) => renderTemplate` ${renderComponent($$result2, "DashboardComponent", Dashboard, { "client:load": true, "client:component-hydration": "load", "client:component-path": "/Users/irfandi/.local/share/opencode/worktree/d686a6dfff4eca54f64908f2ecb63fded89b9888/silent-river/apps/web/src/components/Dashboard", "client:component-export": "default" })} ` })}`;
}, "/Users/irfandi/.local/share/opencode/worktree/d686a6dfff4eca54f64908f2ecb63fded89b9888/silent-river/apps/web/src/pages/dashboard.astro", void 0);

const $$file = "/Users/irfandi/.local/share/opencode/worktree/d686a6dfff4eca54f64908f2ecb63fded89b9888/silent-river/apps/web/src/pages/dashboard.astro";
const $$url = "/dashboard";

const _page = /*#__PURE__*/Object.freeze(/*#__PURE__*/Object.defineProperty({
    __proto__: null,
    default: $$Dashboard,
    file: $$file,
    url: $$url
}, Symbol.toStringTag, { value: 'Module' }));

const page = () => _page;

export { page };
