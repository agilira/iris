module github.com/agilira/iris/otel

go 1.24.5

require (
	github.com/agilira/iris v0.0.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
	go.opentelemetry.io/otel/trace v1.28.0
)

replace github.com/agilira/iris => ../

require (
	github.com/agilira/argus v1.0.1 // indirect
	github.com/agilira/flash-flags v1.0.1 // indirect
	github.com/agilira/go-errors v1.1.0 // indirect
	github.com/agilira/go-timecache v1.0.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
)
