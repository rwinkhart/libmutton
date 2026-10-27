## Building libmuttonserver
This repository is also home to the code for the libmuttonserver binary.

Official binaries are built without CGO (for portability), built position-independent (for security, allows [ASLR](https://en.wikipedia.org/wiki/Address_space_layout_randomization)), and stripped of debug info (for size reduction), as follows:
```
CGO_ENABLED=0 go build -buildmode=pie -ldflags="-s -w" -trimpath ./libmuttonserver.go
```

> [!WARNING]
>Ensure the server binary is named `libmuttonserver`!
>If it is not, clients will fail to find & call it via SSH.
