# libmutton
libmutton is a library for building simple, SSH-synchronized password managers in Go.

[MUTN](https://github.com/rwinkhart/MUTN) is its reference implementation (in the CLI).

> [!WARNING]
>It is your responsibility to assess the security and stability of libmutton and to ensure it meets your needs before using it.
>I am not responsible for any data loss or breaches of your information resulting from the use of libmutton.
>libmutton is a new project that is constantly being updated, and though safety and security are priorities, they cannot be guaranteed.

# Developing Thid-Party Clients
See the [developer guide](https://github.com/rwinkhart/libmutton/blob/main/wiki/developers.md).

# Roadmap
#### Release v0.2.1
- [ ] Only run getSSHClient once to prevent being asked for keyfile password multiple times
    - [ ] After this, handle all errors in sync/client.go
- [ ] Ensure all config files and entry files are created with 0600 permissions
- [ ] Add fail-specific error codes
- [x] Split into separate repos
    1. libmutton: backend package (rename to core), sync package, libmuttonserver
    3. MUTN: cli package
#### Release v0.3.0
- [ ] Swap to native (cascade) encryption (custom)
- [ ] Implement "netpin" (quick-unlock) with new encryption
#### Release v0.4.0
- [ ] Password aging support
    - [ ] Append UNIX timestamp to entry names
        - [ ] Add yellow/red dot indicators to entry list readout for when passwords should be changed
#### Release v0.5.0
- [ ] Add refresh/re-encrypt functionality
#### Release v1.0.0
- [ ] Create packaging scripts (libmuttonserver)
    - [ ] Stable source PKGBUILD
    - [ ] Stable source APKBUILD
    - [ ] Debian/Ubuntu
    - [ ] Fedora
    - [ ] FreeBSD
    - [ ] Windows installer
- Hunt for polishing opportunities and bugs
