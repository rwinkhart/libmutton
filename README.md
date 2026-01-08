# libmutton
libmutton is a library for building simple, SSH-synchronized password managers in Go.

[MUTN](https://github.com/rwinkhart/MUTN) is its reference implementation (in the CLI).

> [!WARNING]
>It is your responsibility to assess the security and stability of libmutton and to ensure it meets your needs before using it.
>I am not responsible for any data loss or breaches of your information resulting from the use of libmutton.
>libmutton is a new project that is constantly being updated, and though safety and security are priorities, they cannot be guaranteed.

# Developing Third-Party Clients
See the [developer guide](https://github.com/rwinkhart/libmutton/blob/main/wiki/developers.md).

# Roadmap
#### Release v1.0.0
- [ ] Stabilize API (reduce need for future breaking changes)
- [ ] Ensure usability as external C library for use in platform-native languages (e.g. Swift, Kotlin)
- [ ] Create packaging scripts (libmuttonserver)
    - [ ] Stable source PKGBUILD
    - [ ] Stable source APKBUILD
    - [ ] Debian/Ubuntu
    - [ ] Fedora
    - [ ] FreeBSD
    - [ ] Windows installer
- Perform extensive testing
