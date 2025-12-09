# Creating Custom Trading Strategies

This guide walks you through implementing custom trading strategies for strategy-go.

## Overview

Creating a new strategy involves:
1. Understanding the Strategy interface
2. Creating a strategy package
3. Implementing position calculation logic
4. Handling position lifecycle callbacks
5. Testing your strategy
6. Integrating with trading-cli

**Estimated time**: 2-4 hours for a basic strategy

---

## Step 1: Understand the Strategy Interface

Every strategy must implement this interface:

```go
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
    OnPositionOpened(ctx context.Context, position *Position) error

    // OnPriceUpdate callback for price updates (for trailing, etc.)
    OnPriceUpdate(ctx context.Context, position *Position, currentPrice float64) (*StrategyAction, error)

    // ShouldClose determines if position should be closed
    ShouldClose(ctx context.Context, position *Position, currentPrice float64) (bool, string)
}
```

### Key Types

**PositionParams** - Input to CalculatePosition:
```go
type PositionParams struct {
    Symbol         string
    Side           Side
    EntryPrice     float64
    StopLoss       float64
    AccountBalance float64
    RiskPercent    float64
    MaxLeverage    int
    Params         StrategyParams  // Strategy-specific parameters
}
```

**PositionPlan** - Output from CalculatePosition:
```go
type PositionPlan struct {
    Symbol        string
    Side          Side
    Size          float64         // Position size in asset units
    EntryPrice    float64
    Leverage      int             // Calculated leverage
    StopLoss      *StopLossLevel
    TakeProfits   []*TakeProfitLevel
    RiskAmount    float64         // Dollar amount risked
    RiskPercent   float64
    NotionalValue float64         // size * entry price
    StrategyName  string
    Timestamp     time.Time
}
```

---

## Step 2: Create Strategy Package

Create a new directory for your strategy:

```bash
mkdir -p /path/to/strategy-go/strategies/yourstrategy
cd yourstrategy
```

Create the main file:
```bash
touch yourstrategy.go
```

---

## Step 3: Implement Your Strategy

### Example 1: Fixed TP/SL Strategy

Simple strategy with fixed take profit and stop loss distances:

**File: `strategies/fixedtpsl/fixedtpsl.go`**

```go
package fixedtpsl

import (
    "context"
    "fmt"
    "time"

    "github.com/agatticelli/calculator-go"
    "github.com/agatticelli/strategy-go"
    "github.com/agatticelli/trading-common-types"
)

// FixedTPSLStrategy implements a strategy with fixed TP/SL distances
type FixedTPSLStrategy struct {
    calculator    *calculator.Calculator
    tpPoints      float64  // Take profit distance in points
    slPoints      float64  // Stop loss distance in points
}

// New creates a new fixed TP/SL strategy
func New(tpPoints, slPoints float64) *FixedTPSLStrategy {
    return &FixedTPSLStrategy{
        calculator: calculator.New(125),
        tpPoints:   tpPoints,
        slPoints:   slPoints,
    }
}

func (s *FixedTPSLStrategy) Name() string {
    return "fixed-tpsl"
}

func (s *FixedTPSLStrategy) Description() string {
    return fmt.Sprintf("Fixed TP/SL strategy (TP: %.0f points, SL: %.0f points)",
        s.tpPoints, s.slPoints)
}

func (s *FixedTPSLStrategy) ValidateParams(params strategy.StrategyParams) error {
    // This strategy doesn't require additional parameters
    return nil
}

func (s *FixedTPSLStrategy) CalculatePosition(ctx context.Context, params strategy.PositionParams) (*strategy.PositionPlan, error) {
    // 1. Validate inputs
    if err := s.calculator.ValidateInputs(
        params.Side,
        params.EntryPrice,
        params.StopLoss,
        params.RiskPercent,
        params.AccountBalance,
    ); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // 2. Calculate position size based on risk
    size := s.calculator.CalculateSize(
        params.AccountBalance,
        params.RiskPercent,
        params.EntryPrice,
        params.StopLoss,
        params.Side,
    )

    // 3. Calculate required leverage
    leverage := s.calculator.CalculateLeverage(
        size,
        params.EntryPrice,
        params.AccountBalance,
        params.MaxLeverage,
    )

    // 4. Calculate TP based on fixed points
    var tpPrice float64
    if params.Side == types.SideLong {
        tpPrice = params.EntryPrice + s.tpPoints
    } else {
        tpPrice = params.EntryPrice - s.tpPoints
    }

    // 5. Build position plan
    return &strategy.PositionPlan{
        Symbol:     params.Symbol,
        Side:       params.Side,
        Size:       size,
        EntryPrice: params.EntryPrice,
        Leverage:   leverage,
        StopLoss: &strategy.StopLossLevel{
            Price: params.StopLoss,
            Type:  types.StopLossTypeFixed,
        },
        TakeProfits: []*strategy.TakeProfitLevel{
            {
                Price:      tpPrice,
                Percentage: 100,
                Type:       types.TakeProfitTypeLimit,
            },
        },
        RiskAmount:    params.AccountBalance * params.RiskPercent / 100,
        RiskPercent:   params.RiskPercent,
        NotionalValue: size * params.EntryPrice,
        StrategyName:  s.Name(),
        Timestamp:     time.Now(),
    }, nil
}

func (s *FixedTPSLStrategy) OnPositionOpened(ctx context.Context, position *strategy.Position) error {
    // No additional actions after opening
    return nil
}

func (s *FixedTPSLStrategy) OnPriceUpdate(ctx context.Context, position *strategy.Position, currentPrice float64) (*strategy.StrategyAction, error) {
    // No dynamic adjustments
    return &strategy.StrategyAction{Type: types.ActionTypeNone}, nil
}

func (s *FixedTPSLStrategy) ShouldClose(ctx context.Context, position *strategy.Position, currentPrice float64) (bool, string) {
    // Let TP/SL orders handle closing
    return false, ""
}
```

### Example 2: Trailing Stop Strategy

Strategy that moves stop loss as price moves favorably:

**File: `strategies/trailing/trailing.go`**

```go
package trailing

import (
    "context"
    "fmt"
    "time"

    "github.com/agatticelli/calculator-go"
    "github.com/agatticelli/strategy-go"
    "github.com/agatticelli/trading-common-types"
)

// TrailingStrategy implements a trailing stop loss strategy
type TrailingStrategy struct {
    calculator      *calculator.Calculator
    rrRatio         float64  // Initial RR ratio
    trailPercent    float64  // Trailing percentage (e.g., 0.01 = 1%)
    activationRR    float64  // RR at which trailing activates (e.g., 1.0 = 1:1)
    highestPrice    map[string]float64  // Track highest price per position
}

func New(rrRatio, trailPercent, activationRR float64) *TrailingStrategy {
    return &TrailingStrategy{
        calculator:   calculator.New(125),
        rrRatio:      rrRatio,
        trailPercent: trailPercent,
        activationRR: activationRR,
        highestPrice: make(map[string]float64),
    }
}

func (s *TrailingStrategy) Name() string {
    return "trailing"
}

func (s *TrailingStrategy) Description() string {
    return fmt.Sprintf("Trailing stop strategy (RR: %.1f:1, Trail: %.1f%%, Activation: %.1f:1)",
        s.rrRatio, s.trailPercent*100, s.activationRR)
}

func (s *TrailingStrategy) ValidateParams(params strategy.StrategyParams) error {
    return nil
}

func (s *TrailingStrategy) CalculatePosition(ctx context.Context, params strategy.PositionParams) (*strategy.PositionPlan, error) {
    // Similar to fixed TP/SL, but mark SL as trailing
    if err := s.calculator.ValidateInputs(
        params.Side, params.EntryPrice, params.StopLoss,
        params.RiskPercent, params.AccountBalance,
    ); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    size := s.calculator.CalculateSize(
        params.AccountBalance,
        params.RiskPercent,
        params.EntryPrice,
        params.StopLoss,
        params.Side,
    )

    leverage := s.calculator.CalculateLeverage(
        size, params.EntryPrice, params.AccountBalance, params.MaxLeverage,
    )

    tpPrice := s.calculator.CalculateRRTakeProfit(
        params.EntryPrice, params.StopLoss, s.rrRatio, params.Side,
    )

    return &strategy.PositionPlan{
        Symbol:     params.Symbol,
        Side:       params.Side,
        Size:       size,
        EntryPrice: params.EntryPrice,
        Leverage:   leverage,
        StopLoss: &strategy.StopLossLevel{
            Price: params.StopLoss,
            Type:  types.StopLossTypeTrailing,  // Mark as trailing
        },
        TakeProfits: []*strategy.TakeProfitLevel{
            {Price: tpPrice, Percentage: 100, Type: types.TakeProfitTypeLimit},
        },
        RiskAmount:    params.AccountBalance * params.RiskPercent / 100,
        RiskPercent:   params.RiskPercent,
        NotionalValue: size * params.EntryPrice,
        StrategyName:  s.Name(),
        Timestamp:     time.Now(),
    }, nil
}

func (s *TrailingStrategy) OnPositionOpened(ctx context.Context, position *strategy.Position) error {
    // Initialize highest price tracking
    s.highestPrice[position.Symbol] = position.EntryPrice
    return nil
}

func (s *TrailingStrategy) OnPriceUpdate(ctx context.Context, position *strategy.Position, currentPrice float64) (*strategy.StrategyAction, error) {
    // Track highest price
    if highest, ok := s.highestPrice[position.Symbol]; !ok || currentPrice > highest {
        s.highestPrice[position.Symbol] = currentPrice
    }

    highest := s.highestPrice[position.Symbol]

    // Calculate current profit in RR terms
    slDistance := abs(position.EntryPrice - position.LiquidationPrice)  // Approximation
    currentProfit := abs(currentPrice - position.EntryPrice)
    currentRR := currentProfit / slDistance

    // Only trail if we've reached activation RR
    if currentRR < s.activationRR {
        return &strategy.StrategyAction{Type: types.ActionTypeNone}, nil
    }

    // Calculate new stop loss
    var newSL float64
    if position.Side == types.SideLong {
        newSL = highest * (1 - s.trailPercent)
        // Only move SL up, never down
        if newSL <= position.LiquidationPrice {
            return &strategy.StrategyAction{Type: types.ActionTypeNone}, nil
        }
    } else {
        newSL = highest * (1 + s.trailPercent)
        if newSL >= position.LiquidationPrice {
            return &strategy.StrategyAction{Type: types.ActionTypeNone}, nil
        }
    }

    // Return action to adjust stop loss
    return &strategy.StrategyAction{
        Type:     types.ActionTypeAdjustSL,
        NewPrice: newSL,
    }, nil
}

func (s *TrailingStrategy) ShouldClose(ctx context.Context, position *strategy.Position, currentPrice float64) (bool, string) {
    // Don't close manually - let trailing SL handle it
    return false, ""
}

func abs(x float64) float64 {
    if x < 0 {
        return -x
    }
    return x
}
```

### Example 3: Multi-TP Strategy

Strategy with multiple take profit levels:

**File: `strategies/multitp/multitp.go`**

```go
package multitp

import (
    "context"
    "fmt"
    "time"

    "github.com/agatticelli/calculator-go"
    "github.com/agatticelli/strategy-go"
    "github.com/agatticelli/trading-common-types"
)

// MultiTPStrategy implements a strategy with multiple TP levels
type MultiTPStrategy struct {
    calculator *calculator.Calculator
    tpLevels   []TPLevel  // TP levels with percentages
}

type TPLevel struct {
    RRRatio    float64  // Risk-reward ratio for this level
    Percentage float64  // Percentage to close (0-100)
}

func New(tpLevels []TPLevel) *MultiTPStrategy {
    return &MultiTPStrategy{
        calculator: calculator.New(125),
        tpLevels:   tpLevels,
    }
}

func (s *MultiTPStrategy) Name() string {
    return "multi-tp"
}

func (s *MultiTPStrategy) Description() string {
    return fmt.Sprintf("Multi-TP strategy with %d levels", len(s.tpLevels))
}

func (s *MultiTPStrategy) ValidateParams(params strategy.StrategyParams) error {
    // Validate that percentages sum to 100
    total := 0.0
    for _, level := range s.tpLevels {
        total += level.Percentage
    }
    if total != 100.0 {
        return fmt.Errorf("TP percentages must sum to 100, got %.2f", total)
    }
    return nil
}

func (s *MultiTPStrategy) CalculatePosition(ctx context.Context, params strategy.PositionParams) (*strategy.PositionPlan, error) {
    if err := s.calculator.ValidateInputs(
        params.Side, params.EntryPrice, params.StopLoss,
        params.RiskPercent, params.AccountBalance,
    ); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    size := s.calculator.CalculateSize(
        params.AccountBalance,
        params.RiskPercent,
        params.EntryPrice,
        params.StopLoss,
        params.Side,
    )

    leverage := s.calculator.CalculateLeverage(
        size, params.EntryPrice, params.AccountBalance, params.MaxLeverage,
    )

    // Calculate multiple TP levels
    tpLevels := make([]*strategy.TakeProfitLevel, len(s.tpLevels))
    for i, level := range s.tpLevels {
        tpPrice := s.calculator.CalculateRRTakeProfit(
            params.EntryPrice, params.StopLoss, level.RRRatio, params.Side,
        )
        tpLevels[i] = &strategy.TakeProfitLevel{
            Price:      tpPrice,
            Percentage: level.Percentage,
            Type:       types.TakeProfitTypeLimit,
        }
    }

    return &strategy.PositionPlan{
        Symbol:        params.Symbol,
        Side:          params.Side,
        Size:          size,
        EntryPrice:    params.EntryPrice,
        Leverage:      leverage,
        StopLoss:      &strategy.StopLossLevel{
            Price: params.StopLoss,
            Type:  types.StopLossTypeFixed,
        },
        TakeProfits:   tpLevels,  // Multiple TP levels
        RiskAmount:    params.AccountBalance * params.RiskPercent / 100,
        RiskPercent:   params.RiskPercent,
        NotionalValue: size * params.EntryPrice,
        StrategyName:  s.Name(),
        Timestamp:     time.Now(),
    }, nil
}

func (s *MultiTPStrategy) OnPositionOpened(ctx context.Context, position *strategy.Position) error {
    return nil
}

func (s *MultiTPStrategy) OnPriceUpdate(ctx context.Context, position *strategy.Position, currentPrice float64) (*strategy.StrategyAction, error) {
    return &strategy.StrategyAction{Type: types.ActionTypeNone}, nil
}

func (s *MultiTPStrategy) ShouldClose(ctx context.Context, position *strategy.Position, currentPrice float64) (bool, string) {
    return false, ""
}
```

---

## Step 4: Testing Your Strategy

Create comprehensive tests:

**File: `strategies/yourstrategy/yourstrategy_test.go`**

```go
package yourstrategy

import (
    "context"
    "testing"

    "github.com/agatticelli/strategy-go"
    "github.com/agatticelli/trading-common-types"
)

func TestCalculatePosition(t *testing.T) {
    strat := New(/* params */)

    tests := []struct {
        name       string
        params     strategy.PositionParams
        wantSize   float64
        wantLeverage int
        wantErr    bool
    }{
        {
            name: "LONG position 2% risk",
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
            wantErr:      false,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            plan, err := strat.CalculatePosition(ctx, tt.params)

            if tt.wantErr {
                if err == nil {
                    t.Error("Expected error, got nil")
                }
                return
            }

            if err != nil {
                t.Fatalf("Unexpected error: %v", err)
            }

            if plan.Size != tt.wantSize {
                t.Errorf("Size = %.4f, want %.4f", plan.Size, tt.wantSize)
            }

            if plan.Leverage != tt.wantLeverage {
                t.Errorf("Leverage = %d, want %d", plan.Leverage, tt.wantLeverage)
            }
        })
    }
}

func TestOnPriceUpdate(t *testing.T) {
    strat := New(/* params */)
    ctx := context.Background()

    position := &strategy.Position{
        Symbol:     "BTC-USDT",
        Side:       types.SideLong,
        Size:       0.1,
        EntryPrice: 45000.0,
        MarkPrice:  46000.0,
    }

    action, err := strat.OnPriceUpdate(ctx, position, 47000.0)
    if err != nil {
        t.Fatalf("OnPriceUpdate error: %v", err)
    }

    // Assert expected action
    if action.Type != types.ActionTypeAdjustSL {
        t.Errorf("Action type = %v, want %v", action.Type, types.ActionTypeAdjustSL)
    }
}
```

Run tests:
```bash
go test ./strategies/yourstrategy -v
```

---

## Step 5: Integration with trading-cli

To use your strategy in trading-cli:

**File: `trading-cli/internal/executor/executor.go`**

```go
import (
    "github.com/agatticelli/strategy-go/strategies/yourstrategy"
)

// In NewExecutor or similar
func GetStrategy(name string) strategy.Strategy {
    switch name {
    case "risk-ratio":
        return riskratio.New(2.0)
    case "fixed-tpsl":
        return fixedtpsl.New(500, 250)  // Your strategy
    case "trailing":
        return trailing.New(2.0, 0.01, 1.0)
    case "multi-tp":
        levels := []multitp.TPLevel{
            {RRRatio: 1.0, Percentage: 30},
            {RRRatio: 2.0, Percentage: 40},
            {RRRatio: 3.0, Percentage: 30},
        }
        return multitp.New(levels)
    default:
        return riskratio.New(2.0)  // default
    }
}
```

Use from CLI:
```bash
./trading-cli --strategy fixed-tpsl open \
  --symbol BTC-USDT \
  --side LONG \
  --entry 45000 \
  --sl 44500 \
  --risk 2.0
```

---

## Best Practices

### 1. Use calculator-go for All Math

```go
// Bad
size := (balance * risk) / (entry - sl)

// Good
size := s.calculator.CalculateSize(balance, risk, entry, sl, side)
```

### 2. Validate All Inputs

```go
func (s *YourStrategy) CalculatePosition(ctx context.Context, params strategy.PositionParams) (*strategy.PositionPlan, error) {
    // Always validate first
    if err := s.calculator.ValidateInputs(...); err != nil {
        return nil, err
    }

    // Your logic...
}
```

### 3. Handle Both LONG and SHORT

```go
var tpPrice float64
if params.Side == types.SideLong {
    tpPrice = entry + distance
} else {
    tpPrice = entry - distance
}
```

### 4. Use Context for Cancellation

```go
func (s *YourStrategy) CalculatePosition(ctx context.Context, params strategy.PositionParams) (*strategy.PositionPlan, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Proceed...
}
```

### 5. Document Strategy Parameters

```go
// YourStrategy implements a custom trading strategy.
//
// Parameters:
// - param1: description
// - param2: description
//
// Example:
//   strat := New(param1, param2)
//   plan, err := strat.CalculatePosition(ctx, params)
type YourStrategy struct {
    // ...
}
```

---

## Common Strategy Patterns

### Pattern 1: Break Even Move

Move SL to entry after price moves favorably:

```go
func (s *Strategy) OnPriceUpdate(ctx context.Context, position *strategy.Position, currentPrice float64) (*strategy.StrategyAction, error) {
    // Calculate profit
    slDistance := abs(position.EntryPrice - position.LiquidationPrice)
    profit := abs(currentPrice - position.EntryPrice)

    // If profit >= 1:1, move to break even
    if profit >= slDistance {
        return &strategy.StrategyAction{
            Type:     types.ActionTypeAdjustSL,
            NewPrice: position.EntryPrice,
        }, nil
    }

    return &strategy.StrategyAction{Type: types.ActionTypeNone}, nil
}
```

### Pattern 2: Partial Profit Taking

Close partial position at milestones:

```go
func (s *Strategy) OnPriceUpdate(ctx context.Context, position *strategy.Position, currentPrice float64) (*strategy.StrategyAction, error) {
    calc := calculator.New(125)
    pnlPercent := calc.CalculatePnLPercent(position.Side, position.EntryPrice, currentPrice)

    // Close 50% at 1% profit
    if pnlPercent >= 1.0 && !s.hasPartialClose[position.Symbol] {
        s.hasPartialClose[position.Symbol] = true
        return &strategy.StrategyAction{
            Type:       types.ActionTypeClose,
            Percentage: 50,
        }, nil
    }

    return &strategy.StrategyAction{Type: types.ActionTypeNone}, nil
}
```

### Pattern 3: Time-Based Exits

Exit after certain time period:

```go
func (s *Strategy) ShouldClose(ctx context.Context, position *strategy.Position, currentPrice float64) (bool, string) {
    // Close if position open > 24 hours
    if time.Since(position.Timestamp) > 24*time.Hour {
        return true, "time limit exceeded"
    }
    return false, ""
}
```

---

## Checklist

Before considering your strategy complete:

- [ ] All interface methods implemented
- [ ] Uses calculator-go for calculations
- [ ] Handles both LONG and SHORT
- [ ] Input validation working
- [ ] Unit tests written and passing
- [ ] Tested with demo account
- [ ] Documentation complete
- [ ] Examples provided
- [ ] Edge cases handled

---

## Example Strategies to Study

1. **RiskRatio** (`strategies/riskratio/`) - Simple fixed RR strategy
2. **Trailing** (this guide) - Dynamic stop loss
3. **MultiTP** (this guide) - Multiple take profit levels

---

## Need Help?

- Check existing strategies for reference
- Read calculator-go documentation
- Test thoroughly with demo accounts
- Ask questions in discussions

---

**Happy strategy building!** 📈
