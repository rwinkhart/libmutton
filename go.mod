module github.com/rwinkhart/libmutton

go 1.24.2

require (
	github.com/fortis/go-steam-totp v0.0.0-20171114202746-18e928674727
	github.com/pkg/sftp v1.13.9
	github.com/pquerna/otp v1.4.1-0.20231130234153-3357de7c0481
	github.com/rwinkhart/rcw v0.0.0-20250504183741-f176d57b6ba5
	golang.design/x/clipboard v0.7.0 // only for Android builds
	golang.org/x/crypto v0.37.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/boombuler/barcode v1.0.2 // indirect
	github.com/kr/fs v0.1.0 // indirect
	golang.org/x/exp/shiny v0.0.0-20250408133849-7e4ce0ab07d0 // indirect; only for Android builds
	golang.org/x/image v0.26.0 // indirect; only for Android builds
	golang.org/x/mobile v0.0.0-20250408133729-978277e7eaf7 // indirect; only for Android builds
	golang.org/x/sys v0.32.0 // indirect
)
