package strategy

import (
	"context"

	"github.com/agatticelli/trading-go/broker"
)

// Strategy defines the interface all trading strategies must implement
type Strategy interface {
	// Name returns the strategy name
	Name() string

	// Description returns a human-readable description
	Description() string

	// ValidateParams validates strategy parameters before execution
	ValidateParams(params StrategyParams) error

	// CalculatePosition calculates position size, leverage, TP/SL levels
	CalculatePosition(ctx context.Context, params PositionParams) (*PositionPlan, error)

	// OnPositionOpened callback after position is opened
	OnPositionOpened(ctx context.Context, position *broker.Position) error

	// OnPriceUpdate callback for price updates (for trailing, etc.)
	OnPriceUpdate(ctx context.Context, position *broker.Position, currentPrice float64) (*StrategyAction, error)

	// ShouldClose determines if position should be closed
	ShouldClose(ctx context.Context, position *broker.Position, currentPrice float64) (bool, string)
}

// StrategyParams contains strategy-specific parameters
type StrategyParams map[string]interface{}
