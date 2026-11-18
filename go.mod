module github.com/Realms4239/cgo

go 1.25

require (
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/lipgloss v1.1.0
	golang.org/x/sys v0.36.0
)

require github.com/cilium/ebpf v0.19.0 // indirect

require github.com/Realms4239/cgo/exporter v0.1.0 // indirect

require (
	github.com/Realms4239/cgo/api v0.1.0
	github.com/Realms4239/cgo/attest v0.1.0
	github.com/Realms4239/cgo/audit v0.1.2
	github.com/Realms4239/cgo/capabilities v0.1.0
	github.com/Realms4239/cgo/doctor v0.1.0
	github.com/Realms4239/cgo/figures v0.1.0
	github.com/Realms4239/cgo/iface v0.1.0
	github.com/Realms4239/cgo/metrics v0.1.0
	github.com/Realms4239/cgo/model v0.1.1
	github.com/Realms4239/cgo/probe v0.1.0
	github.com/Realms4239/cgo/profile v0.1.1
	github.com/Realms4239/cgo/qdisc v0.1.0
	github.com/Realms4239/cgo/remediate v0.1.0
	github.com/Realms4239/cgo/results v0.1.3
	github.com/Realms4239/cgo/runreq v0.1.0
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/colorprofile v0.2.3-0.20250311203215-f60798e515dc // indirect
	github.com/charmbracelet/x/ansi v0.10.1 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13-0.20250311204145-2c3ea96c31dd // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.2.0 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/text v0.23.0 // indirect
)

replace github.com/Realms4239/cgo/model => ./model

replace github.com/Realms4239/cgo/metrics => ./metrics

replace github.com/Realms4239/cgo/qdisc => ./qdisc

replace github.com/Realms4239/cgo/results => ./results

replace github.com/Realms4239/cgo/runreq => ./runreq

replace github.com/Realms4239/cgo/figures => ./figures

replace github.com/Realms4239/cgo/probe => ./probe

replace github.com/Realms4239/cgo/capabilities => ./capabilities

replace github.com/Realms4239/cgo/profile => ./profile

replace github.com/Realms4239/cgo/api => ./api

replace github.com/Realms4239/cgo/audit => ./audit

replace github.com/Realms4239/cgo/doctor => ./doctor

replace github.com/Realms4239/cgo/iface => ./iface

replace github.com/Realms4239/cgo/remediate => ./remediate

replace github.com/Realms4239/cgo/exporter => ./exporter

replace github.com/Realms4239/cgo/attest => ./attest
