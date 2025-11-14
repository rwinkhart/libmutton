module github.com/rwinkhart/libmutton

go 1.25.4

require (
	github.com/pkg/sftp v1.13.10
	github.com/pquerna/otp v1.5.0
	github.com/rwinkhart/go-boilerplate v0.1.1-0.20251110055016-10ee4f91fcb6
	github.com/rwinkhart/rcw v0.2.3
	golang.org/x/crypto v0.43.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/boombuler/barcode v1.1.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/rwinkhart/peercred-mini v0.1.2 // indirect
	golang.org/x/sys v0.38.0 // indirect
)

replace golang.org/x/sys => github.com/rwinkhart/sys v0.38.0

replace github.com/Microsoft/go-winio => github.com/rwinkhart/go-winio v0.1.0
