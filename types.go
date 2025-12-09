package strategy

import (
	"github.com/agatticelli/trading-common-types"
)

// Re-export common types for backward compatibility and convenience
// This allows users to use strategy.Side instead of types.Side

type (
	// Core types
	Side           = types.Side
	OrderType      = types.OrderType
	StopLossType   = types.StopLossType
	TakeProfitType = types.TakeProfitType
	ActionType     = types.ActionType

	// Position and Order types
	Position     = types.Position
	OrderRequest = types.OrderRequest

	// Strategy types
	PositionParams    = types.PositionParams
	PositionPlan      = types.PositionPlan
	StopLossLevel     = types.StopLossLevel
	TakeProfitLevel   = types.TakeProfitLevel
	StrategyAction    = types.StrategyAction
)

// Re-export constants
const (
	SideLong  = types.SideLong
	SideShort = types.SideShort

	OrderTypeMarket     = types.OrderTypeMarket
	OrderTypeLimit      = types.OrderTypeLimit
	OrderTypeStop       = types.OrderTypeStop
	OrderTypeTakeProfit = types.OrderTypeTakeProfit
	OrderTypeTrailing   = types.OrderTypeTrailingStop

	StopLossTypeFixed    = types.StopLossTypeFixed
	StopLossTypeTrailing = types.StopLossTypeTrailing

	TakeProfitTypeLimit    = types.TakeProfitTypeLimit
	TakeProfitTypeTrailing = types.TakeProfitTypeTrailing

	ActionTypeNone        = types.ActionTypeNone
	ActionTypeAdjustSL    = types.ActionTypeAdjustSL
	ActionTypeAdjustTP    = types.ActionTypeAdjustTP
	ActionTypeClose       = types.ActionTypeClose
	ActionTypeAddPosition = types.ActionTypeAddPosition
)
