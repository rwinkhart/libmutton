module github.com/rwinkhart/libmutton

go 1.26.0

require (
	github.com/pkg/sftp v1.13.10
	github.com/pquerna/otp v1.5.0
	github.com/rwinkhart/go-boilerplate v0.3.1
	github.com/rwinkhart/rcw v0.3.0
	golang.org/x/crypto v0.48.0
	golang.org/x/sys v0.41.0
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/boombuler/barcode v1.1.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/rwinkhart/peercred-mini v0.1.4 // indirect
)

replace golang.org/x/sys => github.com/rwinkhart/sys v0.41.0

replace github.com/Microsoft/go-winio => github.com/rwinkhart/go-winio v0.1.1
