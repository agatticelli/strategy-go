package main

import (
	"context"
	"fmt"

	"github.com/agatticelli/strategy-go"
	"github.com/agatticelli/strategy-go/strategies/riskratio"
)

// This example demonstrates input validation and error handling
func main() {
	fmt.Println("=== Strategy Validation Example ===\n")

	strat := riskratio.New(2.0)

	// Test 1: Valid inputs
	fmt.Println("Test 1: Valid LONG position")
	validParams := strategy.PositionParams{
		Symbol:         "BTC-USDT",
		Side:           strategy.SideLong,
		EntryPrice:     45000.0,
		StopLoss:       44500.0,
		AccountBalance: 1000.0,
		RiskPercent:    2.0,
		MaxLeverage:    125,
	}

	plan, err := strat.CalculatePosition(context.Background(), validParams)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n\n", err)
	} else {
		fmt.Printf("  ✅ Valid - Position size: %.4f, Leverage: %dx\n\n", plan.Size, plan.Leverage)
	}

	// Test 2: Invalid stop loss (above entry for LONG)
	fmt.Println("Test 2: Invalid stop loss placement (LONG SL above entry)")
	invalidSL := strategy.PositionParams{
		Symbol:         "BTC-USDT",
		Side:           strategy.SideLong,
		EntryPrice:     45000.0,
		StopLoss:       45500.0, // Invalid: SL above entry
		AccountBalance: 1000.0,
		RiskPercent:    2.0,
		MaxLeverage:    125,
	}

	_, err = strat.CalculatePosition(context.Background(), invalidSL)
	if err != nil {
		fmt.Printf("  ❌ Error (expected): %v\n\n", err)
	} else {
		fmt.Println("  ✅ Unexpectedly passed\n")
	}

	// Test 3: Invalid risk percentage
	fmt.Println("Test 3: Invalid risk percentage (> 100%)")
	invalidRisk := strategy.PositionParams{
		Symbol:         "BTC-USDT",
		Side:           strategy.SideLong,
		EntryPrice:     45000.0,
		StopLoss:       44500.0,
		AccountBalance: 1000.0,
		RiskPercent:    150.0, // Invalid: > 100%
		MaxLeverage:    125,
	}

	_, err = strat.CalculatePosition(context.Background(), invalidRisk)
	if err != nil {
		fmt.Printf("  ❌ Error (expected): %v\n\n", err)
	} else {
		fmt.Println("  ✅ Unexpectedly passed\n")
	}

	// Test 4: Zero or negative values
	fmt.Println("Test 4: Invalid entry price (negative)")
	invalidEntry := strategy.PositionParams{
		Symbol:         "BTC-USDT",
		Side:           strategy.SideLong,
		EntryPrice:     -45000.0, // Invalid: negative
		StopLoss:       44500.0,
		AccountBalance: 1000.0,
		RiskPercent:    2.0,
		MaxLeverage:    125,
	}

	_, err = strat.CalculatePosition(context.Background(), invalidEntry)
	if err != nil {
		fmt.Printf("  ❌ Error (expected): %v\n\n", err)
	} else {
		fmt.Println("  ✅ Unexpectedly passed\n")
	}

	// Test 5: SHORT with invalid SL (below entry)
	fmt.Println("Test 5: Invalid stop loss for SHORT (SL below entry)")
	invalidShortSL := strategy.PositionParams{
		Symbol:         "BTC-USDT",
		Side:           strategy.SideShort,
		EntryPrice:     45000.0,
		StopLoss:       44500.0, // Invalid: SL below entry for SHORT
		AccountBalance: 1000.0,
		RiskPercent:    2.0,
		MaxLeverage:    125,
	}

	_, err = strat.CalculatePosition(context.Background(), invalidShortSL)
	if err != nil {
		fmt.Printf("  ❌ Error (expected): %v\n\n", err)
	} else {
		fmt.Println("  ✅ Unexpectedly passed\n")
	}

	// Test 6: Valid SHORT position
	fmt.Println("Test 6: Valid SHORT position")
	validShort := strategy.PositionParams{
		Symbol:         "BTC-USDT",
		Side:           strategy.SideShort,
		EntryPrice:     45000.0,
		StopLoss:       45500.0, // Valid: SL above entry for SHORT
		AccountBalance: 1000.0,
		RiskPercent:    2.0,
		MaxLeverage:    125,
	}

	plan, err = strat.CalculatePosition(context.Background(), validShort)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n\n", err)
	} else {
		fmt.Printf("  ✅ Valid - Position size: %.4f, TP: $%.2f\n\n", plan.Size, plan.TakeProfits[0].Price)
	}

	fmt.Println("💡 Always validate inputs before calculating positions!")
	fmt.Println("   The strategy uses calculator-go's ValidateInputs() internally")
}
