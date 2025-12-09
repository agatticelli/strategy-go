package calculator

import (
	"fmt"
	"math"

	"github.com/agatticelli/trading-go/broker"
)

// Calculator handles position sizing and risk calculations
type Calculator struct {
	maxLeverage int
}

// New creates a new calculator
func New(maxLeverage int) *Calculator {
	return &Calculator{
		maxLeverage: maxLeverage,
	}
}

// CalculateSize calculates position size based on risk
// Formula: size = risk_amount / price_risk
// Where: risk_amount = balance * (risk% / 100)
//        price_risk = |entry - stop_loss|
func (c *Calculator) CalculateSize(balance, riskPercent, entry, stopLoss float64, side broker.Side) float64 {
	riskAmount := balance * (riskPercent / 100)

	var priceRisk float64
	if side == broker.SideLong {
		priceRisk = entry - stopLoss
	} else {
		priceRisk = stopLoss - entry
	}

	return riskAmount / priceRisk
}

// CalculateLeverage calculates required leverage
// Formula: leverage = ceil(notional_value / balance)
// Where: notional_value = size * price
func (c *Calculator) CalculateLeverage(size, price, balance float64, maxLeverage int) int {
	notional := size * price
	requiredLeverage := int(math.Ceil(notional / balance))

	if requiredLeverage > maxLeverage {
		return maxLeverage
	}
	if requiredLeverage < 1 {
		return 1
	}
	return requiredLeverage
}

// CalculateRRTakeProfit calculates TP price based on risk-reward ratio
// Formula:
//   LONG:  tp = entry + (sl_distance * rr_ratio)
//   SHORT: tp = entry - (sl_distance * rr_ratio)
func (c *Calculator) CalculateRRTakeProfit(entry, stopLoss float64, rrRatio float64, side broker.Side) float64 {
	slDistance := math.Abs(entry - stopLoss)

	if side == broker.SideLong {
		return entry + (slDistance * rrRatio)
	}
	return entry - (slDistance * rrRatio)
}

// ValidatePriceLogic validates order price logic (prevent market execution)
func (c *Calculator) ValidatePriceLogic(side broker.Side, entry, current float64) error {
	if side == broker.SideLong && entry >= current {
		return fmt.Errorf("LONG limit order entry (%.4f) must be below current price (%.4f)", entry, current)
	}
	if side == broker.SideShort && entry <= current {
		return fmt.Errorf("SHORT limit order entry (%.4f) must be above current price (%.4f)", entry, current)
	}
	return nil
}

// ValidateStopLoss validates SL price logic
func (c *Calculator) ValidateStopLoss(side broker.Side, entry, stopLoss float64) error {
	if side == broker.SideLong && stopLoss >= entry {
		return fmt.Errorf("LONG stop loss (%.4f) must be below entry (%.4f)", stopLoss, entry)
	}
	if side == broker.SideShort && stopLoss <= entry {
		return fmt.Errorf("SHORT stop loss (%.4f) must be above entry (%.4f)", stopLoss, entry)
	}
	return nil
}

// CalculatePnLPercent calculates PnL percentage for a position
// Formula:
//   LONG:  (mark - entry) / entry * 100
//   SHORT: (entry - mark) / entry * 100
func (c *Calculator) CalculatePnLPercent(side broker.Side, entryPrice, markPrice float64) float64 {
	if entryPrice <= 0 {
		return 0
	}
	if side == broker.SideLong {
		return ((markPrice - entryPrice) / entryPrice) * 100
	}
	return ((entryPrice - markPrice) / entryPrice) * 100
}

// CalculateDistanceToPrice calculates percentage distance from current price to target
// Formula:
//   LONG:  (target - current) / current * 100
//   SHORT: (current - target) / current * 100
func (c *Calculator) CalculateDistanceToPrice(side broker.Side, currentPrice, targetPrice float64) float64 {
	if currentPrice <= 0 {
		return 0
	}
	if side == broker.SideLong {
		return ((targetPrice - currentPrice) / currentPrice) * 100
	}
	return ((currentPrice - targetPrice) / currentPrice) * 100
}

// CalculateExpectedPnL calculates expected PnL for a closing order
// Returns both nominal (dollar) and percentage values
func (c *Calculator) CalculateExpectedPnL(side broker.Side, entryPrice, exitPrice, size float64) (nominal float64, percent float64) {
	if side == broker.SideLong {
		nominal = (exitPrice - entryPrice) * size
	} else {
		nominal = (entryPrice - exitPrice) * size
	}

	percent = c.CalculatePnLPercent(side, entryPrice, exitPrice)
	return nominal, percent
}

// ValidateInputs validates all calculation inputs
func (c *Calculator) ValidateInputs(side broker.Side, entryPrice, stopLoss, riskPercent, accountEquity float64) error {
	if entryPrice <= 0 {
		return fmt.Errorf("entry price must be positive: %.2f", entryPrice)
	}

	if stopLoss <= 0 {
		return fmt.Errorf("stop loss must be positive: %.2f", stopLoss)
	}

	if riskPercent <= 0 || riskPercent > 100 {
		return fmt.Errorf("risk percent must be between 0 and 100: %.2f", riskPercent)
	}

	if accountEquity <= 0 {
		return fmt.Errorf("account equity must be positive: %.2f", accountEquity)
	}

	// Validate stop loss position relative to entry
	if err := c.ValidateStopLoss(side, entryPrice, stopLoss); err != nil {
		return err
	}

	return nil
}
