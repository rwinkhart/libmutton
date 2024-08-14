## Known Bugs - libmutton
- On Windows, GPG is sometimes (seems unpredictable) incredibly slow to start (often after a reboot), leading to many operations seemingly hanging
    - **This will be addressed** in the migration off of GPG that will take place before v1.0.0
- Password-protected SSH identity files currently only prompt for password entry in the CLI, and thus they are not yet supported in GUI/TUI implementations
