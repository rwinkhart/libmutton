module github.com/rwinkhart/libmutton

go 1.24.3

require (
	github.com/pkg/sftp v1.13.9
	github.com/pquerna/otp v1.5.0
	github.com/rwinkhart/go-boilerplate v0.0.0-20250509173525-20670ec7bb9c
	github.com/rwinkhart/rcw v0.2.0
	golang.design/x/clipboard v0.7.0 // only for Android builds
	golang.org/x/crypto v0.38.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/boombuler/barcode v1.0.2 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/rwinkhart/peercred-mini v0.1.0 // indirect
	golang.org/x/exp/shiny v0.0.0-20250506013437-ce4c2cf36ca6 // indirect; only for Android builds
	golang.org/x/image v0.27.0 // indirect; only for Android builds
	golang.org/x/mobile v0.0.0-20250506005352-78cd7a343bde // indirect; only for Android builds
	golang.org/x/sys v0.33.0 // indirect
)

replace golang.org/x/sys => github.com/rwinkhart/sys-freebsd-13-xucred v0.33.0

replace github.com/Microsoft/go-winio => github.com/rwinkhart/go-winio-easy-pipe-handles v0.0.0-20250407031321-96994a0e8410
