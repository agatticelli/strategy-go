package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/agatticelli/strategy-go"
	"github.com/agatticelli/strategy-go/strategies/riskratio"
)

// This example demonstrates basic usage of the risk-ratio strategy
func main() {
	fmt.Println("=== Basic Risk-Ratio Strategy Usage ===\n")

	// Create a 2:1 risk-reward strategy
	strat := riskratio.New(2.0)

	fmt.Printf("Strategy: %s\n", strat.Name())
	fmt.Printf("Description: %s\n\n", strat.Description())

	// Define position parameters
	params := strategy.PositionParams{
		Symbol:         "BTC-USDT",
		Side:           strategy.SideLong,
		EntryPrice:     45000.0,
		StopLoss:       44500.0,
		AccountBalance: 1000.0,
		RiskPercent:    2.0,
		MaxLeverage:    125,
	}

	fmt.Println("📊 Input Parameters:")
	fmt.Printf("  Symbol: %s\n", params.Symbol)
	fmt.Printf("  Side: %s\n", params.Side)
	fmt.Printf("  Entry Price: $%.2f\n", params.EntryPrice)
	fmt.Printf("  Stop Loss: $%.2f\n", params.StopLoss)
	fmt.Printf("  Account Balance: $%.2f\n", params.AccountBalance)
	fmt.Printf("  Risk Percent: %.1f%%\n", params.RiskPercent)
	fmt.Printf("  Max Leverage: %dx\n\n", params.MaxLeverage)

	// Calculate position plan
	plan, err := strat.CalculatePosition(context.Background(), params)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	// Display results
	fmt.Println("✅ Calculated Position Plan:")
	fmt.Printf("  Position Size: %.4f %s\n", plan.Size, params.Symbol[:3])
	fmt.Printf("  Leverage: %dx\n", plan.Leverage)
	fmt.Printf("  Notional Value: $%.2f\n", plan.NotionalValue)
	fmt.Printf("  Risk Amount: $%.2f (%.1f%% of balance)\n\n", plan.RiskAmount, plan.RiskPercent)

	fmt.Println("📍 Stop Loss:")
	if plan.StopLoss != nil {
		fmt.Printf("  Price: $%.2f\n", plan.StopLoss.Price)
		fmt.Printf("  Type: %s\n\n", plan.StopLoss.Type)
	}

	fmt.Println("🎯 Take Profit Levels:")
	for i, tp := range plan.TakeProfits {
		fmt.Printf("  TP #%d:\n", i+1)
		fmt.Printf("    Price: $%.2f\n", tp.Price)
		fmt.Printf("    Close: %.0f%% of position\n", tp.Percentage)
		fmt.Printf("    Type: %s\n", tp.Type)
	}

	// Calculate expected profits
	slDistance := params.EntryPrice - params.StopLoss
	tpDistance := plan.TakeProfits[0].Price - params.EntryPrice
	fmt.Printf("\n💡 Risk-Reward Analysis:\n")
	fmt.Printf("  Risk Distance: $%.2f per contract\n", slDistance)
	fmt.Printf("  Reward Distance: $%.2f per contract\n", tpDistance)
	fmt.Printf("  Actual R:R Ratio: %.2f:1\n", tpDistance/slDistance)
	fmt.Printf("\n  If SL hit: Loss = $%.2f\n", plan.RiskAmount)
	fmt.Printf("  If TP hit: Profit = $%.2f\n", (tpDistance/slDistance)*plan.RiskAmount)

	fmt.Println("\n" + strings.Repeat("=", 60))

	// Example 2: SHORT position with 3:1 RR
	fmt.Println("\n=== Example 2: SHORT Position (3:1 RR) ===\n")

	strat3R := riskratio.New(3.0)
	fmt.Printf("Strategy: %s\n\n", strat3R.Description())

	shortParams := strategy.PositionParams{
		Symbol:         "ETH-USDT",
		Side:           strategy.SideShort,
		EntryPrice:     3000.0,
		StopLoss:       3100.0,
		AccountBalance: 1000.0,
		RiskPercent:    1.5,
		MaxLeverage:    100,
	}

	fmt.Println("📊 Input Parameters:")
	fmt.Printf("  Symbol: %s\n", shortParams.Symbol)
	fmt.Printf("  Side: %s\n", shortParams.Side)
	fmt.Printf("  Entry Price: $%.2f\n", shortParams.EntryPrice)
	fmt.Printf("  Stop Loss: $%.2f\n", shortParams.StopLoss)
	fmt.Printf("  Risk Percent: %.1f%%\n\n", shortParams.RiskPercent)

	shortPlan, err := strat3R.CalculatePosition(context.Background(), shortParams)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println("✅ Calculated Position Plan:")
	fmt.Printf("  Position Size: %.4f %s\n", shortPlan.Size, shortParams.Symbol[:3])
	fmt.Printf("  Leverage: %dx\n", shortPlan.Leverage)
	fmt.Printf("  Take Profit: $%.2f\n", shortPlan.TakeProfits[0].Price)
	fmt.Printf("  Risk Amount: $%.2f\n", shortPlan.RiskAmount)

	shortSlDistance := shortParams.StopLoss - shortParams.EntryPrice
	shortTpDistance := shortParams.EntryPrice - shortPlan.TakeProfits[0].Price
	fmt.Printf("\n💡 R:R Ratio: %.2f:1\n", shortTpDistance/shortSlDistance)
	fmt.Printf("  If SL hit: Loss = $%.2f\n", shortPlan.RiskAmount)
	fmt.Printf("  If TP hit: Profit = $%.2f\n", (shortTpDistance/shortSlDistance)*shortPlan.RiskAmount)
}
