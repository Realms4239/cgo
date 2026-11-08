module github.com/Realms4239/cgo/api

go 1.25

require (
	github.com/Realms4239/cgo/audit v0.1.0
	github.com/Realms4239/cgo/figures v0.1.0
	github.com/Realms4239/cgo/metrics v0.1.0
	github.com/Realms4239/cgo/model v0.1.0
	github.com/Realms4239/cgo/probe v0.1.0
	github.com/Realms4239/cgo/remediate v0.1.0
	github.com/Realms4239/cgo/results v0.1.0
)

replace (
	github.com/Realms4239/cgo/audit => ../audit
	github.com/Realms4239/cgo/figures => ../figures
	github.com/Realms4239/cgo/metrics => ../metrics
	github.com/Realms4239/cgo/model => ../../model
	github.com/Realms4239/cgo/probe => ../probe
	github.com/Realms4239/cgo/remediate => ../remediate
	github.com/Realms4239/cgo/results => ../results
)
