package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"pryx-core/internal/audit"
	"pryx-core/internal/config"
	"pryx-core/internal/cost"
	"pryx-core/internal/store"
)

var (
	costService *cost.CostService
	costTracker *cost.CostTracker
)

func initCostService() {
	cfg := config.Load()
	s, err := store.New(cfg.DatabasePath)
	if err != nil {
		fmt.Printf("Failed to initialize store: %v\n", err)
		os.Exit(1)
	}

	auditRepo := audit.NewAuditRepository(s.DB)
	pricingMgr := cost.NewPricingManager()

	costTracker = cost.NewCostTracker(auditRepo, pricingMgr)
	calculator := cost.NewCostCalculator(pricingMgr)

	costService = cost.NewCostService(costTracker, calculator, pricingMgr, s)
}

func runCost(args []string) int {
	if len(args) == 0 {
		printCostHelp()
		return 2
	}

	cmd := args[0]
	cmdArgs := args[1:]

	initCostService()

	switch cmd {
	case "summary":
		return runCostSummary()
	case "daily":
		return runCostDaily(cmdArgs)
	case "monthly":
		return runCostMonthly(cmdArgs)
	case "budget":
		return runCostBudget(cmdArgs)
	case "pricing":
		return runCostPricing()
	case "optimize":
		return runCostOptimize()
	default:
		fmt.Printf("Unknown cost command: %s\n", cmd)
		printCostHelp()
		return 2
	}
}

func printCostHelp() {
	fmt.Println("pryx-core cost <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  summary           Show total cost summary")
	fmt.Println("  daily            Show daily cost breakdown")
	fmt.Println("  monthly          Show monthly cost breakdown")
	fmt.Println("  budget           Manage cost budget")
	fmt.Println("  pricing          Show model pricing")
	fmt.Println("  optimize         Show cost optimization suggestions")
	fmt.Println("")
	fmt.Println("Budget subcommands:")
	fmt.Println("  set --daily <amount> --monthly <amount>   Set budget limits")
	fmt.Println("  status                                 Show current budget status")
}

func runCostSummary() int {
	summary, err := costService.GetCurrentSessionCost()
	if err != nil {
		fmt.Printf("Failed to get cost summary: %v\n", err)
		return 1
	}

	fmt.Println("Cost Summary")
	fmt.Println("============")
	fmt.Printf("Total Cost:         $%.4f\n", summary.TotalCost)
	fmt.Printf("Total Input Tokens:  %d\n", summary.TotalInputTokens)
	fmt.Printf("Total Output Tokens: %d\n", summary.TotalOutputTokens)
	fmt.Printf("Total Tokens:        %d\n", summary.TotalTokens)
	fmt.Printf("Total Requests:      %d\n", summary.RequestCount)
	fmt.Printf("Average Cost/Req:   $%.4f\n", summary.AverageCostPerReq)
	return 0
}

func runCostDaily(args []string) int {
	days := 7
	if len(args) > 0 {
		_, err := fmt.Sscanf(args[0], "%d", &days)
		if err != nil {
			fmt.Printf("Invalid days parameter: %v\n", err)
			return 1
		}
	}

	startDate := time.Now().AddDate(0, 0, -days)
	endDate := time.Now()

	breakdown, err := costTracker.GetDailyCostsByDateRange(startDate, endDate)
	if err != nil {
		fmt.Printf("Failed to get daily breakdown: %v\n", err)
		return 1
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Date\tTotal Cost\tRequests\tTokens\n")
	fmt.Fprintf(w, "----\t----------\t--------\t------\n")

	for _, day := range breakdown {
		fmt.Fprintf(w, "%s\t$%.4f\t%d\t%d\n",
			day.PeriodStart.Format("2006-01-02"),
			day.TotalCost,
			day.RequestCount,
			day.TotalTokens,
		)
	}

	w.Flush()
	return 0
}

func runCostMonthly(args []string) int {
	months := 1
	if len(args) > 0 {
		_, err := fmt.Sscanf(args[0], "%d", &months)
		if err != nil {
			fmt.Printf("Invalid months parameter: %v\n", err)
			return 1
		}
	}

	// Use current month for now
	monthCost, err := costTracker.GetMonthlyCosts(time.Now().Year(), time.Now().Month())
	if err != nil {
		fmt.Printf("Failed to get monthly breakdown: %v\n", err)
		return 1
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Month\tTotal Cost\tRequests\tTokens\n")
	fmt.Fprintf(w, "-----\t----------\t--------\t------\n")

	fmt.Fprintf(w, "%s\t$%.4f\t%d\t%d\n",
		monthCost.PeriodStart.Format("2006-01"),
		monthCost.TotalCost,
		monthCost.RequestCount,
		monthCost.TotalTokens,
	)

	w.Flush()
	return 0
}

func runCostBudget(args []string) int {
	if len(args) == 0 {
		// Show budget status
		status := costService.GetBudgetStatus("default")

		fmt.Println("Budget Status")
		fmt.Println("============")
		fmt.Printf("Daily Spent:   $%.2f / $%.2f (%.1f%%)\n",
			status.DailySpent, status.DailyRemaining+status.DailySpent, status.DailyPercent)
		fmt.Printf("Monthly Spent: $%.2f / $%.2f (%.1f%%)\n",
			status.MonthlySpent, status.MonthlyRemaining+status.MonthlySpent, status.MonthlyPercent)

		if status.IsOverBudget {
			fmt.Println("\n⚠️  Over budget!")
		}

		if len(status.Warnings) > 0 {
			fmt.Println("\nWarnings:")
			for _, warn := range status.Warnings {
				fmt.Printf("  - %s\n", warn)
			}
		}

		return 0
	}

	subCmd := args[0]
	if subCmd == "set" {
		return runCostBudgetSet(args[1:])
	}

	fmt.Printf("Unknown budget command: %s\n", subCmd)
	return 2
}

func runCostBudgetSet(args []string) int {
	var daily, monthly float64

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--daily":
			if i+1 >= len(args) {
				fmt.Println("Missing value for --daily")
				return 1
			}
			_, err := fmt.Sscanf(args[i+1], "%f", &daily)
			if err != nil {
				fmt.Printf("Invalid daily amount: %v\n", err)
				return 1
			}
			i++
		case "--monthly":
			if i+1 >= len(args) {
				fmt.Println("Missing value for --monthly")
				return 1
			}
			_, err := fmt.Sscanf(args[i+1], "%f", &monthly)
			if err != nil {
				fmt.Printf("Invalid monthly amount: %v\n", err)
				return 1
			}
			i++
		}
	}

	budget := cost.BudgetConfig{
		DailyBudget:   daily,
		MonthlyBudget: monthly,
	}

	costService.SetBudget("default", budget)

	fmt.Printf("Budget set: Daily $%.2f, Monthly $%.2f\n", daily, monthly)
	return 0
}

func runCostPricing() int {
	allPricing := costService.GetAllModelPricing()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Model\tProvider\tInput ($/1K)\tOutput ($/1K)\n")
	fmt.Fprintf(w, "-----\t--------\t------------\t-------------\n")

	for _, p := range allPricing {
		fmt.Fprintf(w, "%s\t%s\t$%.2f\t$%.2f\n",
			p.ModelID, p.Provider, p.InputPricePer1K, p.OutputPricePer1K)
	}

	w.Flush()
	return 0
}

func runCostOptimize() int {
	// For now, return no optimizations
	fmt.Println("Cost Optimization Suggestions")
	fmt.Println("============================")
	fmt.Println("No optimization suggestions available.")
	return 0
}
