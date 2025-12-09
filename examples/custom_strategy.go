package main

import (
	"context"
	"fmt"

	"github.com/agatticelli/calculator-go"
	"github.com/agatticelli/strategy-go"
)

// ConservativeStrategy is a custom strategy with conservative risk management
type ConservativeStrategy struct {
	calculator *calculator.Calculator
	rrRatio    float64
	maxRisk    float64
}

// NewConservative creates a new conservative strategy
func NewConservative() *ConservativeStrategy {
	return &ConservativeStrategy{
		calculator: calculator.New(10), // Conservative: max 10x leverage
		rrRatio:    1.5,                // Conservative: 1.5:1 RR
		maxRisk:    1.0,                // Conservative: max 1% risk per trade
	}
}

func (s *ConservativeStrategy) Name() string {
	return "conservative"
}

func (s *ConservativeStrategy) Description() string {
	return fmt.Sprintf("Conservative strategy (%.1f:1 RR, max %.1f%% risk, max 10x leverage)", s.rrRatio, s.maxRisk)
}

func (s *ConservativeStrategy) ValidateParams(params strategy.StrategyParams) error {
	return nil
}

func (s *ConservativeStrategy) CalculatePosition(ctx context.Context, params strategy.PositionParams) (*strategy.PositionPlan, error) {
	// Conservative: cap risk at maxRisk%
	riskPercent := params.RiskPercent
	if riskPercent > s.maxRisk {
		riskPercent = s.maxRisk
		fmt.Printf("⚠️  Risk capped at %.1f%% (requested %.1f%%)\n", s.maxRisk, params.RiskPercent)
	}

	// Validate inputs
	if err := s.calculator.ValidateInputs(params.Side, params.EntryPrice, params.StopLoss, riskPercent, params.AccountBalance); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Calculate position size
	size := s.calculator.CalculateSize(
		params.AccountBalance,
		riskPercent,
		params.EntryPrice,
		params.StopLoss,
		params.Side,
	)

	// Calculate leverage (capped at 10x)
	leverage := s.calculator.CalculateLeverage(size, params.EntryPrice, params.AccountBalance, 10)

	// Calculate TP based on conservative RR ratio
	tpPrice := s.calculator.CalculateRRTakeProfit(params.EntryPrice, params.StopLoss, s.rrRatio, params.Side)

	return &strategy.PositionPlan{
		Symbol:     params.Symbol,
		Side:       params.Side,
		Size:       size,
		EntryPrice: params.EntryPrice,
		Leverage:   leverage,
		StopLoss: &strategy.StopLossLevel{
			Price: params.StopLoss,
			Type:  strategy.StopLossTypeFixed,
		},
		TakeProfits: []*strategy.TakeProfitLevel{
			{
				Price:      tpPrice,
				Percentage: 100,
				Type:       strategy.TakeProfitTypeLimit,
			},
		},
		RiskAmount:    params.AccountBalance * riskPercent / 100,
		RiskPercent:   riskPercent,
		NotionalValue: size * params.EntryPrice,
		StrategyName:  s.Name(),
	}, nil
}

func (s *ConservativeStrategy) OnPositionOpened(ctx context.Context, position *strategy.Position) error {
	return nil
}

func (s *ConservativeStrategy) OnPriceUpdate(ctx context.Context, position *strategy.Position, currentPrice float64) (*strategy.StrategyAction, error) {
	return &strategy.StrategyAction{Type: strategy.ActionTypeNone}, nil
}

func (s *ConservativeStrategy) ShouldClose(ctx context.Context, position *strategy.Position, currentPrice float64) (bool, string) {
	return false, ""
}

func main() {
	fmt.Println("=== Custom Strategy Example ===\n")

	// Create custom conservative strategy
	strat := NewConservative()

	fmt.Printf("Strategy: %s\n", strat.Name())
	fmt.Printf("Description: %s\n\n", strat.Description())

	// Test with high risk percentage (will be capped)
	params := strategy.PositionParams{
		Symbol:         "BTC-USDT",
		Side:           strategy.SideLong,
		EntryPrice:     45000.0,
		StopLoss:       44500.0,
		AccountBalance: 1000.0,
		RiskPercent:    3.0, // Requesting 3%, but strategy will cap at 1%
		MaxLeverage:    125,
	}

	fmt.Println("📊 Input Parameters:")
	fmt.Printf("  Requested Risk: %.1f%%\n", params.RiskPercent)
	fmt.Printf("  Requested Max Leverage: %dx\n\n", params.MaxLeverage)

	plan, err := strat.CalculatePosition(context.Background(), params)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println("\n✅ Calculated Position Plan:")
	fmt.Printf("  Actual Risk: %.1f%% (capped by strategy)\n", plan.RiskPercent)
	fmt.Printf("  Actual Leverage: %dx (capped at 10x)\n", plan.Leverage)
	fmt.Printf("  Position Size: %.4f\n", plan.Size)
	fmt.Printf("  Take Profit: $%.2f (1.5:1 RR)\n", plan.TakeProfits[0].Price)
	fmt.Printf("  Risk Amount: $%.2f\n", plan.RiskAmount)

	fmt.Println("\n💡 This demonstrates how custom strategies can:")
	fmt.Println("   - Cap risk percentage to enforce discipline")
	fmt.Println("   - Limit maximum leverage to reduce liquidation risk")
	fmt.Println("   - Set custom risk-reward ratios")
	fmt.Println("   - Add any custom logic you need")
}
