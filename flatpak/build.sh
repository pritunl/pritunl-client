#!/bin/bash
set -e

VERSION="$1"
if [ -z "$VERSION" ]; then
    echo "usage: pritunl-flatpak-build <version>" >&2
    exit 1
fi

GIT_URL="https://github.com/pritunl/pritunl-client.git"
BASE_URL="https://repo.pritunl.com"
REPO_DIR="/flatpak"
SIGN_KEY="$REPO_DIR/sign.key"
SRC_DIR="/src"
BUILD_DIR="/build"
APP_ID="com.pritunl.Client"
BRANCH="stable"
GNUPG_DIR="$BUILD_DIR/gnupg"

if [ ! -d "$REPO_DIR" ]; then
    echo "repository directory $REPO_DIR not mounted" >&2
    exit 1
fi
if [ ! -f "$SIGN_KEY" ]; then
    echo "signing key $SIGN_KEY not mounted" >&2
    exit 1
fi

mkdir -p -m 700 "$GNUPG_DIR"
gpg --batch --homedir "$GNUPG_DIR" --import "$SIGN_KEY"
GPG_KEY=$(gpg --batch --homedir "$GNUPG_DIR" --list-secret-keys --with-colons | \
    awk -F: '$1 == "sec" {print $5; exit}')
if [ -z "$GPG_KEY" ]; then
    echo "no secret key found in $SIGN_KEY" >&2
    exit 1
fi
echo "signing key: $GPG_KEY"

write_channel_files() {
    local channel="$1"
    local title="$2"
    local repo="$REPO_DIR/$channel"
    local url="$BASE_URL/$channel/flatpak/"

    cat > "$repo/pritunl.flatpakrepo" << EOF
[Flatpak Repo]
Title=$title
Url=$url
Homepage=https://client.pritunl.com
Comment=Pritunl Client
Description=Pritunl Client Flatpak repository
GPGKey=$KEY_B64
EOF

    cat > "$repo/$APP_ID.flatpakref" << EOF
[Flatpak Ref]
Name=$APP_ID
Branch=$BRANCH
Title=Pritunl Client
Url=$url
RuntimeRepo=https://dl.flathub.org/repo/flathub.flatpakrepo
IsRuntime=false
GPGKey=$KEY_B64
EOF
}

echo "version: $VERSION"
if [[ "$VERSION" =~ ^[0-9a-f]{7,40}$ ]]; then
    echo "building from commit"
    git init --quiet "$SRC_DIR"
    git -C "$SRC_DIR" remote add origin "$GIT_URL"
    git -C "$SRC_DIR" fetch --quiet --depth 1 origin "$VERSION"
    git -C "$SRC_DIR" checkout --quiet FETCH_HEAD
else
    git clone --quiet --depth 1 --branch "$VERSION" "$GIT_URL" "$SRC_DIR"
fi
COMMIT=$(git -C "$SRC_DIR" rev-parse HEAD)
echo "commit: $COMMIT"

mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"
cp "$SRC_DIR/flatpak/com.pritunl.Client.yml" ./
cp "$SRC_DIR/flatpak/flathub.json" ./
cp "$SRC_DIR/flatpak/generated-sources.json" ./
cp "$SRC_DIR/flatpak/cli-go.mod.yml" ./
cp "$SRC_DIR/flatpak/cli-modules.txt" ./
cp "$SRC_DIR/flatpak/service-go.mod.yml" ./
cp "$SRC_DIR/flatpak/service-modules.txt" ./
sed -i "s|commit: [0-9a-f]\{40\}|commit: $COMMIT|" com.pritunl.Client.yml
grep -n "url:\|commit:" com.pritunl.Client.yml

gpgconf --homedir "$GNUPG_DIR" --kill gpg-agent
flatpak-builder \
    --force-clean \
    --disable-rofiles-fuse \
    --default-branch="$BRANCH" \
    --repo="$REPO_DIR/stable" \
    --gpg-sign="$GPG_KEY" \
    --gpg-homedir="$GNUPG_DIR" \
    build-dir com.pritunl.Client.yml

flatpak build-update-repo \
    --generate-static-deltas \
    --prune \
    --title="Pritunl" \
    --gpg-sign="$GPG_KEY" \
    --gpg-homedir="$GNUPG_DIR" \
    "$REPO_DIR/stable"

KEY_B64=$(gpg --batch --homedir "$GNUPG_DIR" --export "$GPG_KEY" | base64 -w0)
write_channel_files stable "Pritunl"

rm -rf "$REPO_DIR/unstable"
rsync -a --delete "$REPO_DIR/stable/" "$REPO_DIR/unstable/"
write_channel_files unstable "Pritunl Unstable"

echo "exported $APP_ID $VERSION to $REPO_DIR/stable and $REPO_DIR/unstable"
