## Building libmuttonserver
This repository is also home to the code for the libmuttonserver binary.

Official binaries are stripped of debug info for size and built without CGO (except for distribution packages) for portability, as follows:
```
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath ./libmuttonserver.go
```

**Ensure the server binary is named `libmuttonserver`!**
