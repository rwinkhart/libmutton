module github.com/rwinkhart/libmutton

go 1.24.3

require (
	github.com/fortis/go-steam-totp v0.0.0-20171114202746-18e928674727
	github.com/pkg/sftp v1.13.9
	github.com/pquerna/otp v1.4.1-0.20231130234153-3357de7c0481
	github.com/rwinkhart/rcw v0.0.0-20250508234041-b49d9f1eea42
	golang.design/x/clipboard v0.7.0 // only for Android builds
	golang.org/x/crypto v0.38.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/boombuler/barcode v1.0.2 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/rwinkhart/peercred-mini v0.0.0-20250407033241-c09add2eceea // indirect
	golang.org/x/exp/shiny v0.0.0-20250506013437-ce4c2cf36ca6 // indirect; only for Android builds
	golang.org/x/image v0.27.0 // indirect; only for Android builds
	golang.org/x/mobile v0.0.0-20250506005352-78cd7a343bde // indirect; only for Android builds
	golang.org/x/sys v0.33.0 // indirect
)

require github.com/rwinkhart/go-boilerplate v0.0.0-20250509154735-0846290a7620

replace golang.org/x/sys => github.com/rwinkhart/sys-freebsd-13-xucred v0.32.0

replace github.com/Microsoft/go-winio => github.com/rwinkhart/go-winio-easy-pipe-handles v0.0.0-20250407031321-96994a0e8410
