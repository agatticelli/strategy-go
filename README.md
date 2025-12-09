# strategy-go

Trading strategy implementations with position sizing and risk management.

## Features

- Multiple independent strategies
- Position sizing based on risk percentage
- Leverage calculation
- Take profit / Stop loss management
- Risk-reward ratio calculations

## Installation

```bash
go get github.com/gattimassimo/strategy-go
```

## Usage

```go
import (
    "context"
    "github.com/gattimassimo/strategy-go"
    "github.com/gattimassimo/strategy-go/strategies/riskratio"
    "github.com/gattimassimo/trading-go/broker"
)

// Create risk-ratio strategy (2:1 RR)
strat := riskratio.New(2.0)

// Calculate position
plan, err := strat.CalculatePosition(context.Background(), strategy.PositionParams{
    Symbol:         "ETH-USDT",
    Side:           broker.SideLong,
    EntryPrice:     3950,
    StopLoss:       3900,
    AccountBalance: 1000,
    RiskPercent:    2.0,
    MaxLeverage:    125,
})

// plan contains:
// - Size: 0.4 ETH
// - Leverage: 2x
// - Take Profit: 4050 (2:1 RR)
```

## Available Strategies

- **risk-ratio**: Fixed risk-reward ratio (current default)
- **simple**: Fixed stop loss and take profit
- **pyramid**: Position scaling strategy
- **breakout**: Breakout-based strategy (planned)

## Creating Custom Strategies

```go
type MyStrategy struct {}

func (s *MyStrategy) CalculatePosition(ctx context.Context, params PositionParams) (*PositionPlan, error) {
    // Your custom logic here
    return plan, nil
}
```

## License

MIT
