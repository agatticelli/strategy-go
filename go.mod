module github.com/agatticelli/strategy-go

go 1.25.1

require github.com/agatticelli/calculator-go v0.0.0-00010101000000-000000000000

require github.com/agatticelli/trading-go v0.1.0 // indirect

replace (
	github.com/agatticelli/calculator-go => ../calculator-go
	github.com/agatticelli/trading-go => ../trading-go
)
