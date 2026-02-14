## Clipboard Troubleshooting + Clipboard Manager Rant
### Clipboard managers can break libmutton-based password managers
***
The intended way for libmutton to interact with the system clipboard is for it to clear it 30 seconds after copying a field. **Unfortunately, this does not work by default on all systems due to the prevalence of clipboard managers.**

Clipboard managers save a history of what has been copied to the clipboard, which is already a big enough issue on its own for people who copy sensitive information to their clipboard. Some **clipboard managers simply will not allow the clipboard to be empty** and will replace its contents with the last copied item if you attempt to clear it. One such naughty clipboard manager is **KDE Klipper**, which comes **packaged into KDE Plasma** and is typically **enabled by default** on most distributions. Due to this behavior, **KDE Klipper breaks libmutton's clipboard clearing functionality** and should not be left enabled.

It is likely other popular clipboard managers exhibit this behavior. I noticed it with KDE Klipper, which is what prompted me to create this wiki page. **Clipboard managers should not be enabled by default in any environment** or distribution due to their **potential security implications**.
