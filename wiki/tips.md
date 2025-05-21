## Miscellaneous Tips
### TOTP Support
#### Standard 6-digit TOTP
For standard TOTP, all the user needs to do is add their TOTP secret (NOT the full URL) string as the value for the TOTP field in an entry.
#### Steam TOTP
For Steam TOTP, the user must first extract their TOTP secret from the Steam app for Android.

There are several methods for doing this, but the most consistent I have found is [detailed here](https://github.com/JustArchiNET/ArchiSteamFarm/discussions/2786).

This method will result in a Base64-encoded Steam TOTP secret. libmutton requires a base32-encoded secret, so this secret must be converted as follows (on Linux/FreeBSD/Mac): `printf '<shared_secret>' | base64 -d | base32`

To signal to libmutton that this TOTP secret is for Steam, prepend it with "steam@" when adding it to the TOTP field in an entry, e.g. "steam@bAsE32sEcReTkEy". This will tell libmutton to use the Steam-specific TOTP encoder.