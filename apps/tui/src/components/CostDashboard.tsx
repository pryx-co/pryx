import { createSignal, For, Show, onMount } from "solid-js";
import { useKeyboard } from "@opentui/solid";
import { useEffectService, AppRuntime } from "../lib/hooks";
import { loadConfig, saveConfig } from "../services/config";
import { palette } from "../theme";

type CostPeriod = "daily" | "weekly" | "monthly";
type SortOrder = "date" | "cost" | "tokens";

interface CostEntry {
  date: string;
  provider: string;
  model: string;
  tokens: number;
  cost: number;
}

interface Budget {
  enabled: boolean;
  limit: number;
  period: "daily" | "monthly";
  current: number;
  notifications: boolean;
}

interface CostDashboardProps {
  onClose: () => void;
}

export default function CostDashboard(props: CostDashboardProps) {
  const keyboard = useKeyboard();
  const [period, setPeriod] = createSignal<CostPeriod>("daily");
  const [sortOrder, setSortOrder] = createSignal<SortOrder>("date");
  const [costs, setCosts] = createSignal<CostEntry[]>([]);
  const [totalCost, setTotalCost] = createSignal(0);
  const [totalTokens, setTotalTokens] = createSignal(0);
  const [budget, setBudget] = createSignal<Budget | null>(null);
  const [selectedIndex, setSelectedIndex] = createSignal(0);
  const [loading, setLoading] = createSignal(true);
  const [error, setError] = createSignal("");

  onMount(() => {
    loadCostData();
    loadBudget();
    setupKeyboard();
  });

  const setupKeyboard = () => {
    keyboard.bind("1", () => setPeriod("daily"));
    keyboard.bind("2", () => setPeriod("weekly"));
    keyboard.bind("3", () => setPeriod("monthly"));
    keyboard.bind("s", () => toggleSort());
    keyboard.bind("b", () => showBudgetSettings());
    keyboard.bind("o", () => showOptimizations());
    keyboard.bind("q", () => {
      props.onClose();
    });
  };

  const loadCostData = async () => {
    setLoading(true);
    try {
      const response = await fetch("http://localhost:3000/api/cost");
      if (!response.ok) {
        throw new Error("Failed to load cost data");
      }
      const data = await response.json();
      setCosts(data.costs || []);
      setTotalCost(data.totalCost || 0);
      setTotalTokens(data.totalTokens || 0);
    } catch (err) {
      setError(`Failed to load cost data: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const loadBudget = () => {
    const config = loadConfig();
    if (config.budget) {
      setBudget(config.budget);
    }
  };

  const toggleSort = () => {
    setSortOrder(prev => {
      if (prev === "date") return "cost";
      if (prev === "cost") return "tokens";
      return "date";
    });
  };

  const showBudgetSettings = () => {
    // TODO: Implement budget settings modal
    console.log("Show budget settings");
  };

  const showOptimizations = () => {
    // TODO: Implement optimization suggestions
    console.log("Show optimizations");
  };

  const sortedCosts = () => {
    const sorted = [...costs()];
    if (sortOrder() === "date") {
      return sorted.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
    } else if (sortOrder() === "cost") {
      return sorted.sort((a, b) => b.cost - a.cost);
    } else {
      return sorted.sort((a, b) => b.tokens - a.tokens);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(amount);
  };

  const formatTokens = (tokens: number) => {
    if (tokens >= 1000000) {
      return `${(tokens / 1000000).toFixed(1)}M`;
    } else if (tokens >= 1000) {
      return `${(tokens / 1000).toFixed(1)}K`;
    }
    return tokens.toString();
  };

  const getBudgetUsage = () => {
    if (!budget()) return 0;
    return ((budget()!.current / budget()!.limit) * 100).toFixed(1);
  };

  const getBudgetStatus = () => {
    const usage = parseFloat(getBudgetUsage());
    if (usage >= 90) return "critical";
    if (usage >= 70) return "warning";
    return "ok";
  };

  return (
    <Box flexDirection="column" width="100%" height="100%">
      <Box
        flexDirection="row"
        padding={1}
        backgroundColor={palette.primary}
        color={palette.background}
      >
        <Text bold>💰 Cost Dashboard</Text>
        <Box flexGrow={1} />
        <Text>
          Sort: <Text bold>[S]</Text> | Period: <Text bold>[1]</Text> Daily <Text bold>[2]</Text>{" "}
          Weekly <Text bold>[3]</Text> Monthly
        </Text>
        <Text>
          Quit: <Text bold>[Q]</Text>
        </Text>
      </Box>

      <Show when={loading()}>
        <Box padding={2}>
          <Text>Loading cost data...</Text>
        </Box>
      </Show>

      <Show when={error()}>
        <Box padding={2} backgroundColor={palette.error}>
          <Text color={palette.background}>{error()}</Text>
        </Box>
      </Show>

      <Show when={!loading() && !error()}>
        <Box flexDirection="column" padding={1}>
          <Box flexDirection="row" padding={1} backgroundColor={palette.secondary}>
            <Box flexGrow={1}>
              <Text bold>Total Cost</Text>
              <Text fontSize={2}>{formatCurrency(totalCost())}</Text>
            </Box>
            <Box flexGrow={1}>
              <Text bold>Total Tokens</Text>
              <Text fontSize={2}>{formatTokens(totalTokens())}</Text>
            </Box>
            <Box flexGrow={1}>
              <Text bold>Period</Text>
              <Text fontSize={2}>
                {period() === "daily" && "Today"}
                {period() === "weekly" && "This Week"}
                {period() === "monthly" && "This Month"}
              </Text>
            </Box>
          </Box>

          <Show when={budget() && budget()!.enabled}>
            <Box padding={1} marginTop={1} backgroundColor={palette.background}>
              <Text bold>Budget Status</Text>
              <Box flexDirection="row" alignItems="center" marginTop={1}>
                <Text width={15}>Usage:</Text>
                <Text
                  color={
                    getBudgetStatus() === "critical"
                      ? palette.error
                      : getBudgetStatus() === "warning"
                        ? palette.warning
                        : palette.success
                  }
                >
                  {getBudgetUsage()}%
                </Text>
                <Box flexGrow={1} />
                <Text>
                  {formatCurrency(budget()!.current)} / {formatCurrency(budget()!.limit)}
                </Text>
              </Box>
              <Box flexDirection="row" alignItems="center" marginTop={1}>
                <Text width={15}>Remaining:</Text>
                <Text bold>{formatCurrency(budget()!.limit - budget()!.current)}</Text>
                <Box flexGrow={1} />
                <Text>
                  Manage Budget: <Text bold>[B]</Text>
                </Text>
              </Box>
            </Box>
          </Show>

          <Box padding={1} marginTop={1} backgroundColor={palette.background}>
            <Text bold>Cost Breakdown</Text>
          </Box>

          <Box flexDirection="column" flexGrow={1} padding={1} backgroundColor={palette.background}>
            <Box flexDirection="row" padding={0.5}>
              <Text width={15}>Date</Text>
              <Text width={20}>Provider</Text>
              <Text width={15}>Model</Text>
              <Text width={10} textAlign="right">
                Tokens
              </Text>
              <Text width={10} textAlign="right">
                Cost
              </Text>
            </Box>

            <Box flexDirection="column">
              <For each={sortedCosts()}>
                {(entry, index) => (
                  <Box
                    flexDirection="row"
                    padding={0.5}
                    backgroundColor={index() === selectedIndex() ? palette.primary : undefined}
                    color={index() === selectedIndex() ? palette.background : undefined}
                  >
                    <Text width={15}>{entry.date}</Text>
                    <Text width={20}>{entry.provider}</Text>
                    <Text width={15}>{entry.model}</Text>
                    <Text width={10} textAlign="right">
                      {formatTokens(entry.tokens)}
                    </Text>
                    <Text width={10} textAlign="right">
                      {formatCurrency(entry.cost)}
                    </Text>
                  </Box>
                )}
              </For>
            </Box>
          </Box>

          <Box flexDirection="row" padding={1} marginTop={1} backgroundColor={palette.secondary}>
            <Text>
              Show Budget: <Text bold>[B]</Text>
            </Text>
            <Box flexGrow={1} />
            <Text>
              Optimizations: <Text bold>[O]</Text>
            </Text>
          </Box>
        </Box>
      </Show>
    </Box>
  );
}

const Box: any = (props: any) => props.children;
const Text: any = (props: any) => {
  const content =
    typeof props.children === "string" ? props.children : props.children?.join?.("") || "";
  return <span style={props}>{content}</span>;
};
