## Developer Guide (for third-party libmutton-based clients)
**Important Notice**: libmutton is in early development and is currently a moving target to develop off of. Feel free to jump in early, but greater change stability will be met with release v1.0.0. Check [here](https://github.com/rwinkhart/libmutton/blob/main/wiki/breaking.md) for planned breaking changes.

libmutton was designed to be usable as a library for building other compatible password managers off of. [MUTN](https://github.com/rwinkhart/MUTN) is the official reference CLI password manager, however libmutton can be implemented in many other unique ways.

## Build Tags
Custom build tags can (and sometimes must) be used to achieve desired results.

These are as follows:
- `interactive`: If making an interactive interface (GUI/TUI/interactive CLI), you probably need to use this build tag. Without it, your entire program will exit after any given operation is completed. This behavior is only desired for non-interactive CLI implementations, such as MUTN.
- `wsl`: Allows creating a Linux binary that can interact with the Windows clipboard (for WSL)
- `termux`: Allows creating an Android binary that can interact with the Termux clipboard (for Android)

## Required Global Variable Manipulation
- `global.GetPassword` must be set to allow for different types of clients (CLI, GUI, TUI) to prompt for the password in the most appropriate way.
- `crypt.Daemonize`, true by default, determines whether to make use of the RCW daemon for password caching. This may be best to disable for interactive clients.
- `crypt.RetryPassword`, true by default, determines whether the crypt package should verify user-typed passwords and re-prompt if needed. Turn this off to handle this uniquely in the client.

## Required Arguments
- `clipclear`: Should be accepted by all non-interactive CLI libmutton implementations (not required for interactive GUI/TUI implementations). In order to clear the clipboard on a timer, non-interactive libmutton-based password managers call another instance of their executable with the `clipclear` argument (e.g. `mutn clipclear`) with the intended clipboard contents provided via STDIN. If after 30 seconds the clipboard contents have not changed, they are cleared. Please accept a `clipclear` argument that calls `clip.ClearArgument()`.
- `startrcwd`: Should be accepted by all libmutton implementations making use of the RCW daemon to cache passwords. Please accept a `startrcwd` argument that calls `crypt.RCWDArgument()`.

## Mobile Clipboard Management
In an effort to reduce dependencies not needed in most environments, libmutton no longer provides clipboard management for mobile platforms (except for Termux). This should be handled by your GUI toolkit/framework.

## Configuration
libmutton-based password manager clients should all share the same JSON configuration file.

On UNIX-like systems, this is located at `~/.config/libmutton/libmuttoncfg.json`. On Windows, it is located at `~\AppData\Local\libmutton\config\libmuttoncfg.json`.

If creating a third-party client that requires extra configuration to be stored, please use the `ClientSpecific` map in the `config.CfgT` type (as used by `config.Write()`) to save your application-specific configuration.

This ensures that a user can use multiple client applications with the same configuration while avoiding conflicts.

## Entry Format
A decrypted libmutton entry is a plaintext file where each line indicates a new field in the entry.

These fields are as follows:
```
0/first line: password
1/second line: username
2/third line: TOTP secret
3/fourth line: URL
4+/fifth line+: notes
```
