module strategy-examples

go 1.25.1

require (
	github.com/agatticelli/calculator-go v0.2.0
	github.com/agatticelli/strategy-go v0.0.0
)

replace (
	github.com/agatticelli/calculator-go => ../../calculator-go
	github.com/agatticelli/strategy-go => ../
)
