module github.com/rwinkhart/libmutton

go 1.24.6

require (
	github.com/pkg/sftp v1.13.9
	github.com/pquerna/otp v1.5.0
	github.com/rwinkhart/go-boilerplate v0.1.0
	github.com/rwinkhart/rcw v0.2.1
	golang.design/x/clipboard v0.7.1 // only for mobile builds
	golang.org/x/crypto v0.39.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/boombuler/barcode v1.0.2 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/rwinkhart/peercred-mini v0.1.0 // indirect
	golang.org/x/exp/shiny v0.0.0-20250606033433-dcc06ee1d476 // indirect; only for mobile builds
	golang.org/x/image v0.28.0 // indirect; only for mobile builds
	golang.org/x/mobile v0.0.0-20250606033058-a2a15c67f36f // indirect; only for mobile builds
	golang.org/x/sys v0.33.0 // indirect
)

replace golang.org/x/sys => github.com/rwinkhart/sys-freebsd-13-xucred v0.33.0

replace github.com/Microsoft/go-winio => github.com/rwinkhart/go-winio-easy-pipe-handles v0.0.0-20250407031321-96994a0e8410
