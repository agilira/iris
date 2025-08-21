module github.com/agilira/iris

go 1.24.5

require (
	github.com/agilira/go-errors v1.0.0
	github.com/agilira/xantos v1.0.0
)

require github.com/agilira/zephyros v1.0.2 // indirect

// Local development
replace github.com/agilira/zephyros => ../zephyros

replace github.com/agilira/go-errors => ../go-errors
