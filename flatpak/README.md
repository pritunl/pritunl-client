# Pritunl Client Flatpak

**UNDER DEVELOPMENT NOT FOR FLATHUB SUBMISSION**

Build information for the Pritunl Client Flatpak package.

## Runtime layout

| Purpose | Path |
|---|---|
| Service Runtime Data | `$XDG_RUNTIME_DIR/app/com.pritunl.Client/` |
| Service Data | `~/.var/app/com.pritunl.Client/config/pritunl/service/` |
| Service Log | `~/.var/app/com.pritunl.Client/config/pritunl/service/pritunl-client.log` |
| Client Data | `~/.var/app/com.pritunl.Client/config/pritunl/` |

## Prerequisites

```bash
flatpak remote-add --if-not-exists --user flathub https://dl.flathub.org/repo/flathub.flatpakrepo
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
flatpak run org.flatpak.Builder --user --install --force-clean --disable-rofiles-fuse build-dir com.pritunl.Client.local.yml
flatpak run com.pritunl.Client
# terminal interface (not finished)
flatpak run --command=pritunl-client com.pritunl.Client
```

## Git Build

```bash
flatpak install flathub org.flatpak.Builder
COMMIT="865713bbddb479f7abf4b9c7e50f7102311ed88b"
mkdir com.pritunl.Client
cd com.pritunl.Client
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/com.pritunl.Client.yml"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/flathub.json"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/generated-sources.json"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/cli-go.mod.yml"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/cli-modules.txt"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/service-go.mod.yml"
wget "https://raw.githubusercontent.com/pritunl/pritunl-client/$COMMIT/flatpak/service-modules.txt"
sed -i "s|commit: [0-9a-f]\{40\}|commit: $COMMIT|" com.pritunl.Client.yml
grep -n "url:\|commit:" com.pritunl.Client.yml
flatpak run --command=flatpak-builder-lint org.flatpak.Builder manifest com.pritunl.Client.yml
flatpak run --command=flatpak-builder-lint org.flatpak.Builder repo repo
flatpak-builder --user --install --force-clean --disable-rofiles-fuse --repo=repo build-dir com.pritunl.Client.yml
flatpak run com.pritunl.Client
```

## Device Authentication

Device authentication requires TPM access by running the commands below.

```bash
flatpak override --user --device=all com.pritunl.Client
sudo tee /etc/udev/rules.d/70-pritunl-tpm.rules << 'EOF'
KERNEL=="tpmrm[0-9]*", SUBSYSTEM=="tpmrm", TAG+="uaccess"
EOF
sudo udevadm control --reload
sudo udevadm trigger --subsystem-match=tpmrm
```

This can be removed using the commands below.

```bash
flatpak override --user --nodevice=all com.pritunl.Client
sudo rm /etc/udev/rules.d/70-pritunl-tpm.rules
sudo udevadm control --reload
sudo setfacl -b /dev/tpmrm0
```

## Debug

```bash
flatpak ps
flatpak enter <instance> sh
```
