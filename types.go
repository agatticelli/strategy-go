package strategy

import (
	"time"

	"github.com/gattimassimo/trading-go/broker"
)

// PositionParams contains input for position calculation
type PositionParams struct {
	Symbol         string
	Side           broker.Side
	EntryPrice     float64
	StopLoss       float64
	AccountBalance float64
	RiskPercent    float64
	MaxLeverage    int

	// Optional strategy-specific params
	Params StrategyParams
}

// PositionPlan is the output of position calculation
type PositionPlan struct {
	Symbol    string
	Side      broker.Side
	Size      float64 // Calculated position size
	EntryPrice float64
	Leverage  int // Calculated leverage

	// Stop loss configuration
	StopLoss *StopLossLevel

	// Take profit levels (can be multiple)
	TakeProfits []*TakeProfitLevel

	// Risk metrics
	RiskAmount    float64 // Dollar risk
	RiskPercent   float64
	NotionalValue float64

	// Metadata
	StrategyName string
	Timestamp    time.Time
}

// StopLossLevel represents a stop loss configuration
type StopLossLevel struct {
	Price float64
	Type  StopLossType // FIXED, TRAILING

	// For trailing stops
	ActivationPrice float64
	CallbackRate    float64
}

// StopLossType represents the type of stop loss
type StopLossType string

const (
	StopLossTypeFixed    StopLossType = "FIXED"
	StopLossTypeTrailing StopLossType = "TRAILING"
)

// TakeProfitLevel represents a take profit level
type TakeProfitLevel struct {
	Price      float64
	Percentage float64        // Percentage of position to close (0-100)
	Type       TakeProfitType // LIMIT, TRAILING

	// For trailing TP
	ActivationPrice float64
	CallbackRate    float64
}

// TakeProfitType represents the type of take profit
type TakeProfitType string

const (
	TakeProfitTypeLimit    TakeProfitType = "LIMIT"
	TakeProfitTypeTrailing TakeProfitType = "TRAILING"
)

// StrategyAction represents an action to take
type StrategyAction struct {
	Type   ActionType
	Reason string
	Orders []*broker.OrderRequest
}

// ActionType represents the type of action
type ActionType string

const (
	ActionTypeNone        ActionType = "NONE"
	ActionTypeAdjustSL    ActionType = "ADJUST_SL"
	ActionTypeAdjustTP    ActionType = "ADJUST_TP"
	ActionTypeClose       ActionType = "CLOSE"
	ActionTypeAddPosition ActionType = "ADD_POSITION"
)
