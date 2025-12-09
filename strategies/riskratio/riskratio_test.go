package riskratio

import (
	"context"
	"math"
	"testing"

	"github.com/agatticelli/strategy-go"
	"github.com/agatticelli/trading-common-types"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		rrRatio float64
	}{
		{
			name:    "Standard 2:1 RR",
			rrRatio: 2.0,
		},
		{
			name:    "Conservative 1.5:1 RR",
			rrRatio: 1.5,
		},
		{
			name:    "Aggressive 3:1 RR",
			rrRatio: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := New(tt.rrRatio)
			if strat == nil {
				t.Fatal("New() returned nil")
			}
			if strat.calculator == nil {
				t.Error("calculator not initialized")
			}
			if strat.rrRatio != tt.rrRatio {
				t.Errorf("rrRatio = %.2f, want %.2f", strat.rrRatio, tt.rrRatio)
			}
		})
	}
}

func TestName(t *testing.T) {
	strat := New(2.0)
	if name := strat.Name(); name != "risk-ratio" {
		t.Errorf("Name() = %q, want %q", name, "risk-ratio")
	}
}

func TestDescription(t *testing.T) {
	tests := []struct {
		name     string
		rrRatio  float64
		wantDesc string
	}{
		{
			name:     "2:1 ratio",
			rrRatio:  2.0,
			wantDesc: "Fixed risk-reward ratio strategy (2.0:1)",
		},
		{
			name:     "1.5:1 ratio",
			rrRatio:  1.5,
			wantDesc: "Fixed risk-reward ratio strategy (1.5:1)",
		},
		{
			name:     "3:1 ratio",
			rrRatio:  3.0,
			wantDesc: "Fixed risk-reward ratio strategy (3.0:1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := New(tt.rrRatio)
			if desc := strat.Description(); desc != tt.wantDesc {
				t.Errorf("Description() = %q, want %q", desc, tt.wantDesc)
			}
		})
	}
}

func TestValidateParams(t *testing.T) {
	strat := New(2.0)

	// Risk-ratio strategy doesn't use additional params
	if err := strat.ValidateParams(strategy.StrategyParams{}); err != nil {
		t.Errorf("ValidateParams() error = %v, want nil", err)
	}

	// Test with some params (should still pass)
	params := strategy.StrategyParams{
		"foo": "bar",
		"baz": 123,
	}
	if err := strat.ValidateParams(params); err != nil {
		t.Errorf("ValidateParams() with params error = %v, want nil", err)
	}
}

func TestCalculatePosition(t *testing.T) {
	tests := []struct {
		name         string
		rrRatio      float64
		params       strategy.PositionParams
		wantSize     float64
		wantLeverage int
		wantTPPrice  float64
		wantSLPrice  float64
		wantRisk     float64
		wantErr      bool
	}{
		{
			name:    "LONG position 2% risk, 2:1 RR",
			rrRatio: 2.0,
			params: strategy.PositionParams{
				Symbol:         "BTC-USDT",
				Side:           types.SideLong,
				EntryPrice:     45000.0,
				StopLoss:       44500.0,
				AccountBalance: 1000.0,
				RiskPercent:    2.0,
				MaxLeverage:    125,
			},
			wantSize:     0.04,
			wantLeverage: 2,
			wantTPPrice:  46000.0, // entry + (500 * 2)
			wantSLPrice:  44500.0,
			wantRisk:     20.0, // 1000 * 2%
			wantErr:      false,
		},
		{
			name:    "SHORT position 2% risk, 2:1 RR",
			rrRatio: 2.0,
			params: strategy.PositionParams{
				Symbol:         "ETH-USDT",
				Side:           types.SideShort,
				EntryPrice:     3000.0,
				StopLoss:       3100.0,
				AccountBalance: 1000.0,
				RiskPercent:    2.0,
				MaxLeverage:    125,
			},
			wantSize:     0.2,
			wantLeverage: 1,
			wantTPPrice:  2800.0, // entry - (100 * 2)
			wantSLPrice:  3100.0,
			wantRisk:     20.0,
			wantErr:      false,
		},
		{
			name:    "LONG position 1% risk, 3:1 RR",
			rrRatio: 3.0,
			params: strategy.PositionParams{
				Symbol:         "BTC-USDT",
				Side:           types.SideLong,
				EntryPrice:     50000.0,
				StopLoss:       49500.0,
				AccountBalance: 2000.0,
				RiskPercent:    1.0,
				MaxLeverage:    125,
			},
			wantSize:     0.04,
			wantLeverage: 1,
			wantTPPrice:  51500.0, // entry + (500 * 3)
			wantSLPrice:  49500.0,
			wantRisk:     20.0, // 2000 * 1%
			wantErr:      false,
		},
		{
			name:    "HIGH leverage position",
			rrRatio: 2.0,
			params: strategy.PositionParams{
				Symbol:         "BTC-USDT",
				Side:           types.SideLong,
				EntryPrice:     45000.0,
				StopLoss:       44900.0,
				AccountBalance: 1000.0,
				RiskPercent:    2.0,
				MaxLeverage:    125,
			},
			wantSize:     0.2,
			wantLeverage: 9,
			wantTPPrice:  45200.0, // entry + (100 * 2)
			wantSLPrice:  44900.0,
			wantRisk:     20.0,
			wantErr:      false,
		},
		{
			name:    "Invalid: LONG with SL above entry",
			rrRatio: 2.0,
			params: strategy.PositionParams{
				Symbol:         "BTC-USDT",
				Side:           types.SideLong,
				EntryPrice:     45000.0,
				StopLoss:       46000.0, // Invalid: SL > entry for LONG
				AccountBalance: 1000.0,
				RiskPercent:    2.0,
				MaxLeverage:    125,
			},
			wantErr: true,
		},
		{
			name:    "Invalid: SHORT with SL below entry",
			rrRatio: 2.0,
			params: strategy.PositionParams{
				Symbol:         "ETH-USDT",
				Side:           types.SideShort,
				EntryPrice:     3000.0,
				StopLoss:       2900.0, // Invalid: SL < entry for SHORT
				AccountBalance: 1000.0,
				RiskPercent:    2.0,
				MaxLeverage:    125,
			},
			wantErr: true,
		},
		{
			name:    "Invalid: negative entry price",
			rrRatio: 2.0,
			params: strategy.PositionParams{
				Symbol:         "BTC-USDT",
				Side:           types.SideLong,
				EntryPrice:     -45000.0,
				StopLoss:       44500.0,
				AccountBalance: 1000.0,
				RiskPercent:    2.0,
				MaxLeverage:    125,
			},
			wantErr: true,
		},
		{
			name:    "Invalid: risk percent too high",
			rrRatio: 2.0,
			params: strategy.PositionParams{
				Symbol:         "BTC-USDT",
				Side:           types.SideLong,
				EntryPrice:     45000.0,
				StopLoss:       44500.0,
				AccountBalance: 1000.0,
				RiskPercent:    150.0, // Invalid: > 100%
				MaxLeverage:    125,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := New(tt.rrRatio)
			ctx := context.Background()

			plan, err := strat.CalculatePosition(ctx, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Error("CalculatePosition() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("CalculatePosition() error = %v, want nil", err)
			}

			// Validate plan fields
			if plan.Symbol != tt.params.Symbol {
				t.Errorf("Symbol = %q, want %q", plan.Symbol, tt.params.Symbol)
			}
			if plan.Side != tt.params.Side {
				t.Errorf("Side = %v, want %v", plan.Side, tt.params.Side)
			}

			// Size (with tolerance for floating point)
			if math.Abs(plan.Size-tt.wantSize) > 0.0001 {
				t.Errorf("Size = %.4f, want %.4f", plan.Size, tt.wantSize)
			}

			// Leverage
			if plan.Leverage != tt.wantLeverage {
				t.Errorf("Leverage = %d, want %d", plan.Leverage, tt.wantLeverage)
			}

			// Stop Loss
			if plan.StopLoss == nil {
				t.Fatal("StopLoss is nil")
			}
			if math.Abs(plan.StopLoss.Price-tt.wantSLPrice) > 0.01 {
				t.Errorf("StopLoss.Price = %.2f, want %.2f", plan.StopLoss.Price, tt.wantSLPrice)
			}
			if plan.StopLoss.Type != types.StopLossTypeFixed {
				t.Errorf("StopLoss.Type = %v, want %v", plan.StopLoss.Type, types.StopLossTypeFixed)
			}

			// Take Profit
			if len(plan.TakeProfits) != 1 {
				t.Fatalf("len(TakeProfits) = %d, want 1", len(plan.TakeProfits))
			}
			tp := plan.TakeProfits[0]
			if math.Abs(tp.Price-tt.wantTPPrice) > 0.01 {
				t.Errorf("TakeProfit.Price = %.2f, want %.2f", tp.Price, tt.wantTPPrice)
			}
			if tp.Percentage != 100 {
				t.Errorf("TakeProfit.Percentage = %.0f, want 100", tp.Percentage)
			}
			if tp.Type != types.TakeProfitTypeLimit {
				t.Errorf("TakeProfit.Type = %v, want %v", tp.Type, types.TakeProfitTypeLimit)
			}

			// Risk
			if math.Abs(plan.RiskAmount-tt.wantRisk) > 0.01 {
				t.Errorf("RiskAmount = %.2f, want %.2f", plan.RiskAmount, tt.wantRisk)
			}
			if plan.RiskPercent != tt.params.RiskPercent {
				t.Errorf("RiskPercent = %.2f, want %.2f", plan.RiskPercent, tt.params.RiskPercent)
			}

			// Notional value
			expectedNotional := plan.Size * tt.params.EntryPrice
			if math.Abs(plan.NotionalValue-expectedNotional) > 0.01 {
				t.Errorf("NotionalValue = %.2f, want %.2f", plan.NotionalValue, expectedNotional)
			}

			// Strategy name
			if plan.StrategyName != "risk-ratio" {
				t.Errorf("StrategyName = %q, want %q", plan.StrategyName, "risk-ratio")
			}

			// Timestamp should be set
			if plan.Timestamp.IsZero() {
				t.Error("Timestamp is zero")
			}
		})
	}
}

func TestOnPositionOpened(t *testing.T) {
	strat := New(2.0)
	ctx := context.Background()

	position := &strategy.Position{
		Symbol:     "BTC-USDT",
		Side:       types.SideLong,
		Size:       0.1,
		EntryPrice: 45000.0,
	}

	// Should always return nil for simple RR strategy
	if err := strat.OnPositionOpened(ctx, position); err != nil {
		t.Errorf("OnPositionOpened() error = %v, want nil", err)
	}
}

func TestOnPriceUpdate(t *testing.T) {
	strat := New(2.0)
	ctx := context.Background()

	position := &strategy.Position{
		Symbol:     "BTC-USDT",
		Side:       types.SideLong,
		Size:       0.1,
		EntryPrice: 45000.0,
		MarkPrice:  46000.0,
	}

	tests := []struct {
		name         string
		currentPrice float64
	}{
		{
			name:         "Price above entry",
			currentPrice: 46000.0,
		},
		{
			name:         "Price below entry",
			currentPrice: 44000.0,
		},
		{
			name:         "Price at entry",
			currentPrice: 45000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := strat.OnPriceUpdate(ctx, position, tt.currentPrice)
			if err != nil {
				t.Errorf("OnPriceUpdate() error = %v, want nil", err)
			}
			if action == nil {
				t.Fatal("OnPriceUpdate() returned nil action")
			}
			// Simple RR strategy should never take action
			if action.Type != types.ActionTypeNone {
				t.Errorf("Action.Type = %v, want %v", action.Type, types.ActionTypeNone)
			}
		})
	}
}

func TestShouldClose(t *testing.T) {
	strat := New(2.0)
	ctx := context.Background()

	position := &strategy.Position{
		Symbol:     "BTC-USDT",
		Side:       types.SideLong,
		Size:       0.1,
		EntryPrice: 45000.0,
		MarkPrice:  46000.0,
	}

	tests := []struct {
		name         string
		currentPrice float64
	}{
		{
			name:         "Price in profit",
			currentPrice: 47000.0,
		},
		{
			name:         "Price in loss",
			currentPrice: 43000.0,
		},
		{
			name:         "Price at break even",
			currentPrice: 45000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldClose, reason := strat.ShouldClose(ctx, position, tt.currentPrice)
			// Simple RR strategy lets TP/SL orders handle closing
			if shouldClose {
				t.Errorf("ShouldClose() = true, want false (reason: %q)", reason)
			}
			if reason != "" {
				t.Errorf("ShouldClose() reason = %q, want empty", reason)
			}
		})
	}
}

// TestCalculatePosition_EdgeCases tests edge cases and boundary conditions
func TestCalculatePosition_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		rrRatio float64
		params  strategy.PositionParams
		wantErr bool
	}{
		{
			name:    "Very small position",
			rrRatio: 2.0,
			params: strategy.PositionParams{
				Symbol:         "BTC-USDT",
				Side:           types.SideLong,
				EntryPrice:     45000.0,
				StopLoss:       44999.0,
				AccountBalance: 1000.0,
				RiskPercent:    0.1,
				MaxLeverage:    125,
			},
			wantErr: false,
		},
		{
			name:    "Zero account balance",
			rrRatio: 2.0,
			params: strategy.PositionParams{
				Symbol:         "BTC-USDT",
				Side:           types.SideLong,
				EntryPrice:     45000.0,
				StopLoss:       44500.0,
				AccountBalance: 0.0,
				RiskPercent:    2.0,
				MaxLeverage:    125,
			},
			wantErr: true,
		},
		{
			name:    "Entry equals stop loss",
			rrRatio: 2.0,
			params: strategy.PositionParams{
				Symbol:         "BTC-USDT",
				Side:           types.SideLong,
				EntryPrice:     45000.0,
				StopLoss:       45000.0,
				AccountBalance: 1000.0,
				RiskPercent:    2.0,
				MaxLeverage:    125,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := New(tt.rrRatio)
			ctx := context.Background()

			_, err := strat.CalculatePosition(ctx, tt.params)

			if tt.wantErr && err == nil {
				t.Error("CalculatePosition() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("CalculatePosition() error = %v, want nil", err)
			}
		})
	}
}
