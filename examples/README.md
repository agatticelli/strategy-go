# Strategy-Go Examples

This directory contains working examples demonstrating how to use strategy-go to build trading strategies.

## Prerequisites

```bash
# The examples use local module replacements
# Make sure both calculator-go and strategy-go are available locally
cd /path/to/strategy-go/examples
go mod tidy
```

## Running the Examples

### 1. Basic Usage
Demonstrates the risk-ratio strategy with both LONG and SHORT positions.

```bash
go run basic_usage.go
```

**What it shows:**
- Creating a risk-ratio strategy (2:1 and 3:1)
- Calculating position plans
- Understanding position size, leverage, and TP/SL levels
- Risk-reward analysis for both LONG and SHORT

### 2. Custom Strategy
Shows how to implement your own strategy with custom risk management rules.

```bash
go run custom_strategy.go
```

**What it shows:**
- Implementing the `Strategy` interface
- Custom risk caps (max 1% risk per trade)
- Custom leverage limits (max 10x)
- Custom risk-reward ratios (1.5:1)
- How strategies can enforce trading discipline

### 3. Validation
Demonstrates input validation and error handling.

```bash
go run validation.go
```

**What it shows:**
- Valid vs invalid stop loss placement
- Risk percentage validation
- Entry price validation
- Different validation rules for LONG vs SHORT
- How validation prevents trading errors

## Run All Examples

```bash
# Basic usage
go run basic_usage.go

# Custom strategy
go run custom_strategy.go

# Validation
go run validation.go
```

## Key Concepts

### Strategy Interface
All strategies must implement:
```go
type Strategy interface {
    Name() string
    Description() string
    ValidateParams(params StrategyParams) error
    CalculatePosition(ctx context.Context, params PositionParams) (*PositionPlan, error)
    OnPositionOpened(ctx context.Context, position *Position) error
    OnPriceUpdate(ctx context.Context, position *Position, currentPrice float64) (*StrategyAction, error)
    ShouldClose(ctx context.Context, position *Position, currentPrice float64) (bool, string)
}
```

### Input: PositionParams
```go
strategy.PositionParams{
    Symbol:         "BTC-USDT",
    Side:           strategy.SideLong,
    EntryPrice:     45000.0,
    StopLoss:       44500.0,
    AccountBalance: 1000.0,
    RiskPercent:    2.0,
    MaxLeverage:    125,
}
```

### Output: PositionPlan
```go
type PositionPlan struct {
    Size          float64  // Calculated position size
    Leverage      int      // Calculated leverage
    StopLoss      *StopLossLevel
    TakeProfits   []*TakeProfitLevel
    RiskAmount    float64
    NotionalValue float64
    // ...
}
```

## Integration Patterns

### 1. CLI Integration
```go
// Get user input
params := getUserInput()

// Calculate position
plan, err := strategy.CalculatePosition(ctx, params)
if err != nil {
    return err
}

// Display to user
fmt.Printf("Position Size: %.4f\n", plan.Size)
fmt.Printf("Take Profit: $%.2f\n", plan.TakeProfits[0].Price)
```

### 2. Automated Trading Bot
```go
// Market signal triggers position calculation
if signal.IsBullish() {
    params := strategy.PositionParams{
        Side: strategy.SideLong,
        // ... populate from signal
    }

    plan, err := strat.CalculatePosition(ctx, params)
    if err != nil {
        log.Error(err)
        return
    }

    // Convert plan to broker orders
    orders := planToOrders(plan)
    broker.PlaceOrders(orders)
}
```

### 3. Backtesting
```go
// Historical data simulation
for _, candle := range historicalData {
    if entrySignal(candle) {
        plan, _ := strat.CalculatePosition(ctx, params)
        simulatePosition(plan, futureData)
    }
}
```

## Type Conversions

strategy-go defines its own types (`strategy.Side`) independent of broker types. The trading-cli handles conversions:

```go
// strategy.Side -> calculator.Side
func calculatorSideFromStrategy(side strategy.Side) calculator.Side {
    if side == strategy.SideLong {
        return calculator.SideLong
    }
    return calculator.SideShort
}
```

This keeps strategy-go **broker-agnostic** and reusable with any broker implementation.

## Further Reading

- [Main README](../README.md) - Full API documentation
- [calculator-go](../../calculator-go/) - Mathematical calculations
- [MIGRATION_STATUS.md](../../trading-cli/MIGRATION_STATUS.md) - Architecture overview
