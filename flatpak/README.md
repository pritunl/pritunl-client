# Pritunl Client Flatpak

Build information for the Pritunl Client Flatpak package.

## Runtime layout

| Purpose | Path |
|---|---|
| Service Runtime Data | `$XDG_RUNTIME_DIR/app/com.pritunl.Client/` |
| Service Data | `~/.var/app/com.pritunl.Client/config/pritunl/service/` |
| Service Log | `~/.var/app/com.pritunl.Client/config/pritunl/service/pritunl-client.log` |
| Client Data | `~/.var/app/com.pritunl.Client/config/pritunl/` |

## Prerequisites

```sh
flatpak remote-add --if-not-exists --user flathub https://dl.flathub.org/repo/flathub.flatpakrepo
flatpak install flathub \
    org.freedesktop.Platform//25.08 \
    org.freedesktop.Sdk//25.08 \
    org.electronjs.Electron2.BaseApp//25.08 \
    org.freedesktop.Sdk.Extension.node24//25.08 \
    org.freedesktop.Sdk.Extension.golang//25.08
```

## Build

```sh
cd flatpak
rm -rf .flatpak-builder/ build-dir/
flatpak-builder --user --install --force-clean --disable-rofiles-fuse build-dir com.pritunl.Client.yml
flatpak run com.pritunl.Client
# terminal interface (not finished)
flatpak run --command=pritunl-client com.pritunl.Client
```
