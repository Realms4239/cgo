module github.com/Realms4239/cgo/audit

go 1.25

require (
	github.com/Realms4239/cgo/metrics v0.1.0
	github.com/Realms4239/cgo/model v0.1.1
	github.com/Realms4239/cgo/probe v0.1.0
	github.com/Realms4239/cgo/remediate v0.1.0
)

replace (
	github.com/Realms4239/cgo/metrics => ../../metrics
	github.com/Realms4239/cgo/probe => ../probe
	github.com/Realms4239/cgo/remediate => ../remediate
)
