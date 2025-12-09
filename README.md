# strategy-go

Trading strategy implementations with position sizing and risk management. This module provides broker-agnostic strategy interfaces and implementations that can be used with any trading system.

## Features

- **Strategy Interface**: Pluggable strategy system for easy customization
- **Position Sizing**: Calculate position size based on risk percentage
- **Leverage Calculation**: Automatic leverage determination based on account balance
- **Risk-Reward Ratios**: Calculate take profit levels based on RR ratios
- **Multiple TP/SL Levels**: Support for partial closes and trailing stops
- **Broker-Agnostic**: Works with any broker through standardized types

## Dependencies

- **[trading-common-types](https://github.com/agatticelli/trading-common-types)**: Shared type definitions (Side, Position, OrderRequest, etc.)
- **[calculator-go](https://github.com/agatticelli/calculator-go)**: Pure math calculations

All types are re-exported for convenience, so you can use `strategy.Side` or `types.Side` interchangeably.

## Installation

```bash
go get github.com/agatticelli/strategy-go
```

## Quick Start

```go
import (
    "context"
    "fmt"

    "github.com/agatticelli/strategy-go"
    "github.com/agatticelli/strategy-go/strategies/riskratio"
)

// Create risk-ratio strategy (2:1 RR)
strat := riskratio.New(2.0)

// Calculate position plan
plan, err := strat.CalculatePosition(context.Background(), strategy.PositionParams{
    Symbol:         "BTC-USDT",
    Side:           strategy.SideLong,
    EntryPrice:     45000.0,
    StopLoss:       44500.0,
    AccountBalance: 1000.0,
    RiskPercent:    2.0,
    MaxLeverage:    125,
})

if err != nil {
    panic(err)
}

fmt.Printf("Position Size: %.4f\n", plan.Size)
fmt.Printf("Leverage: %dx\n", plan.Leverage)
fmt.Printf("Take Profit: $%.2f\n", plan.TakeProfits[0].Price)
fmt.Printf("Risk Amount: $%.2f\n", plan.RiskAmount)
```

**Output:**
```
Position Size: 0.4000
Leverage: 18x
Take Profit: $46000.00
Risk Amount: $20.00
```

For complete working examples, see the [examples/](examples/) directory.

## Available Strategies

### Risk-Ratio Strategy
Fixed risk-reward ratio strategy. Calculates position size based on account risk percentage, and sets take profit at a multiple of the stop loss distance.

```go
// Create a 2:1 risk-reward strategy
strat := riskratio.New(2.0)

// For 3:1 RR
strat := riskratio.New(3.0)
```

**Features:**
- Fixed RR ratio
- Single TP level (100% close)
- Fixed stop loss
- Uses calculator-go for all math

## Architecture

strategy-go is part of a 5-module trading system:

```
calculator-go (v0.2.0)  → Pure math calculations
    ↓
strategy-go             → Trading strategies (this module)
    ↓
trading-cli             → CLI orchestrator
```

**Key Design Decisions:**

1. **Own Type Definitions**: strategy-go defines its own `Side`, `Position`, `OrderRequest` types instead of depending on broker types
2. **Calculator Integration**: Uses calculator-go for all mathematical calculations
3. **Broker-Agnostic**: Strategies work with any broker implementation
4. **Type Conversion**: trading-cli handles conversion between strategy types and broker types

## Strategy Interface

Implement the `Strategy` interface to create custom strategies:

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

### Creating a Custom Strategy

```go
package mystrategies

import (
    "context"
    "github.com/agatticelli/strategy-go"
    "github.com/agatticelli/calculator-go"
)

type ConservativeStrategy struct {
    calculator *calculator.Calculator
}

func NewConservative() *ConservativeStrategy {
    return &ConservativeStrategy{
        calculator: calculator.New(10), // Max 10x leverage
    }
}

func (s *ConservativeStrategy) Name() string {
    return "conservative"
}

func (s *ConservativeStrategy) Description() string {
    return "Conservative strategy with tight SL and 1.5:1 RR"
}

func (s *ConservativeStrategy) CalculatePosition(ctx context.Context, params strategy.PositionParams) (*strategy.PositionPlan, error) {
    // Your custom position calculation logic
    // Use calculator-go for math operations
    size := s.calculator.CalculateSize(
        params.AccountBalance,
        params.RiskPercent,
        params.EntryPrice,
        params.StopLoss,
        calculator.SideLong,
    )

    // Return position plan with your custom logic
    return &strategy.PositionPlan{
        Size: size,
        // ... rest of the plan
    }, nil
}

// Implement remaining interface methods...
```

## Core Types

### Side
```go
type Side string

const (
    SideLong  Side = "LONG"
    SideShort Side = "SHORT"
)
```

### PositionParams
Input parameters for position calculation:
```go
type PositionParams struct {
    Symbol         string
    Side           Side
    EntryPrice     float64
    StopLoss       float64
    AccountBalance float64
    RiskPercent    float64
    MaxLeverage    int
    Params         StrategyParams  // Optional strategy-specific params
}
```

### PositionPlan
Output of position calculation:
```go
type PositionPlan struct {
    Symbol        string
    Side          Side
    Size          float64  // Calculated position size
    EntryPrice    float64
    Leverage      int      // Calculated leverage
    StopLoss      *StopLossLevel
    TakeProfits   []*TakeProfitLevel  // Can have multiple TP levels
    RiskAmount    float64
    RiskPercent   float64
    NotionalValue float64
    StrategyName  string
    Timestamp     time.Time
}
```

### StopLossLevel
```go
type StopLossLevel struct {
    Price           float64
    Type            StopLossType  // FIXED or TRAILING
    ActivationPrice float64       // For trailing stops
    CallbackRate    float64       // For trailing stops
}
```

### TakeProfitLevel
```go
type TakeProfitLevel struct {
    Price           float64
    Percentage      float64          // % of position to close (0-100)
    Type            TakeProfitType   // LIMIT or TRAILING
    ActivationPrice float64          // For trailing TP
    CallbackRate    float64          // For trailing TP
}
```

## Dependencies

- **calculator-go** (v0.2.0+): Pure mathematical calculations
  - Position sizing
  - Leverage calculation
  - Risk-reward ratios
  - PnL calculations

No other dependencies - uses only Go standard library.

## Examples

See the [examples/](examples/) directory for working code:

- **basic_usage.go**: Using the risk-ratio strategy
- **custom_strategy.go**: Implementing a custom strategy
- **multiple_tp.go**: Multiple take profit levels
- **validation.go**: Input validation and error handling

## Testing

```bash
go test ./...
```

## License

MIT
