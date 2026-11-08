module github.com/Realms4239/cgo/results

go 1.25

require (
	github.com/Realms4239/cgo/metrics v0.1.0
	github.com/Realms4239/cgo/model v0.1.0
)

replace (
	github.com/Realms4239/cgo/metrics => ../metrics
	github.com/Realms4239/cgo/model => ../model
)
