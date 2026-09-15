# Pritunl Client Flatpak

Build information for the Pritunl Client Flatpak package.

## Runtime layout

| Purpose | Path |
|---|---|
| Service Runtime Data | `$XDG_RUNTIME_DIR/app/com.pritunl.client/` |
| Service Data | `~/.var/app/com.pritunl.client/config/pritunl/service/` |
| Service Log | `~/.var/app/com.pritunl.client/config/pritunl/service/pritunl-client.log` |
| Client Data | `~/.var/app/com.pritunl.client/config/pritunl/` |

## Pritunl Flatpak Repository

Stable Repository

```bash
flatpak remote-add --user --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo
flatpak remote-add --user --if-not-exists pritunl https://repo.pritunl.com/stable/flatpak/pritunl.flatpakrepo
flatpak install --user pritunl com.pritunl.client
```

Unstable Repository

```bash
flatpak remote-add --user --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo
flatpak remote-add --user --if-not-exists pritunl-unstable https://repo.pritunl.com/unstable/flatpak/pritunl.flatpakrepo
flatpak install --user pritunl com.pritunl.client
```

Run Client

```bash
# Open GUI Interface
flatpak run com.pritunl.client
# Open TUI Interface
flatpak run --command=pritunl-client com.pritunl.client
```

## Device Authentication

Add Device Authentication

```bash
flatpak override --user --device=all com.pritunl.client
sudo tee /etc/udev/rules.d/70-pritunl-tpm.rules << 'EOF'
KERNEL=="tpmrm[0-9]*", SUBSYSTEM=="tpmrm", TAG+="uaccess"
EOF
sudo udevadm control --reload
sudo udevadm trigger --subsystem-match=tpmrm
flatpak kill com.pritunl.client
```

Remove Device Authention

```bash
flatpak override --user --nodevice=all com.pritunl.client
sudo rm /etc/udev/rules.d/70-pritunl-tpm.rules
sudo udevadm control --reload
sudo setfacl -b /dev/tpmrm0
```

## Debug

```bash
flatpak ps
flatpak enter <instance> sh
```

## Prerequisites

```bash
flatpak remote-add --user --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo
flatpak install --user flathub \
    org.flatpak.Builder \
    org.freedesktop.Platform//25.08 \
    org.freedesktop.Sdk//25.08 \
    org.electronjs.Electron2.BaseApp//25.08 \
    org.freedesktop.Sdk.Extension.node24//25.08 \
    org.freedesktop.Sdk.Extension.golang//25.08
```

## Local Build

```bash
git clone https://github.com/pritunl/pritunl-client.git
cd pritunl-client/flatpak
rm -rf .flatpak-builder/ build-dir/
flatpak run org.flatpak.Builder --user --install --force-clean --disable-rofiles-fuse build-dir com.pritunl.client.local.yml
flatpak run com.pritunl.client
flatpak run --command=pritunl-client com.pritunl.client
```

## Git Publish

```bash
flatpak install flathub org.flatpak.Builder
COMMIT="83538d33f7375bc0381fc3584003e45ffc145e22"
mkdir com.pritunl.client
cd com.pritunl.client
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/com.pritunl.client.yml"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/flathub.json"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/generated-sources.json"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/cli-go.mod.yml"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/cli-modules.txt"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/service-go.mod.yml"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/service-modules.txt"
sed -i "s|commit: [0-9a-f]\{40\}|commit: $COMMIT|" com.pritunl.client.yml
grep -n "url:\|commit:" com.pritunl.client.yml
flatpak run --command=flatpak-builder-lint org.flatpak.Builder manifest com.pritunl.client.yml
flatpak run --command=flatpak-builder-lint org.flatpak.Builder repo repo
```
