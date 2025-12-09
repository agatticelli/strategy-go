package riskratio

import (
	"context"
	"fmt"
	"time"

	"github.com/agatticelli/calculator-go"
	"github.com/agatticelli/strategy-go"
)

// RiskRatioStrategy implements fixed risk-reward ratio strategy
// This is the current default strategy from the CLI
type RiskRatioStrategy struct {
	calculator *calculator.Calculator
	rrRatio    float64 // Default RR ratio (e.g., 2.0 for 2:1)
}

// New creates a new risk-ratio strategy
func New(rrRatio float64) *RiskRatioStrategy {
	return &RiskRatioStrategy{
		calculator: calculator.New(125), // Max leverage 125x
		rrRatio:    rrRatio,
	}
}

// Name returns the strategy name
func (s *RiskRatioStrategy) Name() string {
	return "risk-ratio"
}

// Description returns a human-readable description
func (s *RiskRatioStrategy) Description() string {
	return fmt.Sprintf("Fixed risk-reward ratio strategy (%.1f:1)", s.rrRatio)
}

// ValidateParams validates strategy parameters
func (s *RiskRatioStrategy) ValidateParams(params strategy.StrategyParams) error {
	// No additional params needed for risk-ratio strategy
	return nil
}

// CalculatePosition calculates position size, leverage, and TP/SL
func (s *RiskRatioStrategy) CalculatePosition(ctx context.Context, params strategy.PositionParams) (*strategy.PositionPlan, error) {
	// Convert strategy.Side to calculator.Side
	calcSide := calculatorSideFromStrategy(params.Side)

	// Validate inputs
	if err := s.calculator.ValidateInputs(calcSide, params.EntryPrice, params.StopLoss, params.RiskPercent, params.AccountBalance); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 1. Calculate position size based on risk
	// Formula: size = (balance * risk%) / (entry - sl)
	size := s.calculator.CalculateSize(
		params.AccountBalance,
		params.RiskPercent,
		params.EntryPrice,
		params.StopLoss,
		calcSide,
	)

	// 2. Calculate required leverage
	// Formula: leverage = ceil(notional / balance)
	leverage := s.calculator.CalculateLeverage(
		size,
		params.EntryPrice,
		params.AccountBalance,
		params.MaxLeverage,
	)

	// 3. Calculate TP based on RR ratio
	// Formula: tp = entry + (sl_distance * rr_ratio)
	tpPrice := s.calculator.CalculateRRTakeProfit(
		params.EntryPrice,
		params.StopLoss,
		s.rrRatio,
		calcSide,
	)

	// Build position plan
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
		RiskAmount:    params.AccountBalance * params.RiskPercent / 100,
		RiskPercent:   params.RiskPercent,
		NotionalValue: size * params.EntryPrice,
		StrategyName:  s.Name(),
		Timestamp:     time.Now(),
	}, nil
}

// OnPositionOpened callback after position is opened
func (s *RiskRatioStrategy) OnPositionOpened(ctx context.Context, position *strategy.Position) error {
	// No additional actions after opening for simple RR strategy
	return nil
}

// OnPriceUpdate callback for price updates
func (s *RiskRatioStrategy) OnPriceUpdate(ctx context.Context, position *strategy.Position, currentPrice float64) (*strategy.StrategyAction, error) {
	// No dynamic adjustments in simple RR strategy
	return &strategy.StrategyAction{Type: strategy.ActionTypeNone}, nil
}

// ShouldClose determines if position should be closed
func (s *RiskRatioStrategy) ShouldClose(ctx context.Context, position *strategy.Position, currentPrice float64) (bool, string) {
	// Let TP/SL orders handle closing
	return false, ""
}

// calculatorSideFromStrategy converts strategy.Side to calculator.Side
func calculatorSideFromStrategy(side strategy.Side) calculator.Side {
	if side == strategy.SideLong {
		return calculator.SideLong
	}
	return calculator.SideShort
}
