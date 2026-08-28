#!/bin/bash
set -e

APP_VER="$(curl -s https://api.github.com/repos/pritunl/pritunl-client/releases/latest | python3 -c 'import json,sys;print(json.load(sys.stdin)["tag_name"])')"

read -r -p "Install Pritunl Client v$APP_VER? [y/N] " response
if ! [[ "$response" =~ ^([yY][eE][sS]|[yY])+$ ]]
then
    exit
fi

ROOT_PATH="$(pwd)/pritunl-client-$APP_VER"
function clean {
  rm -rf "$ROOT_PATH"
}

trap clean EXIT

curl -L https://github.com/pritunl/pritunl-client/archive/$APP_VER.tar.gz | tar x
cd pritunl-client-$APP_VER

rm -rf build
mkdir -p build/resources
GOPATH="$(pwd)/go" go clean -cache

node -e "fs=require('fs');f='client/package.json';c=fs.readFileSync(f,'utf8');fs.writeFileSync(f,c.replace(/,\s*\"scripts\": \{[^}]*\}/,''))"

# Service
cd service
GOPATH="$(pwd)/go" go get -d
GOPATH="$(pwd)/go" go build -v
cd ..
cp service/service build/resources/pritunl-service

# CLI
cd cli
GOPATH="$(pwd)/go" go get -d
GOPATH="$(pwd)/go" go build -v
cd ..
cp cli/cli build/resources/pritunl-client

# Device Authentication
cd service_macos
rm -f "Pritunl Device Authentication"
swiftc -sdk $(xcrun --show-sdk-path --sdk macosx) -target arm64-apple-macos11 -framework CryptoKit -framework LocalAuthentication -framework Security -framework Foundation device_auth.swift -o "Pritunl Device Authentication"
cp "Pritunl Device Authentication" "../build/resources/Pritunl Device Authentication"
cd ..

# Service Helper
cd service_macos
rm -f pritunl-service-helper
swiftc -sdk $(xcrun --show-sdk-path --sdk macosx) -target arm64-apple-macos13 -framework ServiceManagement -framework Foundation service_helper.swift -o pritunl-service-helper
cp pritunl-service-helper ../build/resources/pritunl-service-helper
cd ..

# Openvpn
cp openvpn_macos/openvpn_arm64 build/resources/pritunl-openvpn

# WireGuard
cp wireguard_macos/bash build/resources/bash
cp wireguard_macos/wg build/resources/wg
cp wireguard_macos/wg-quick build/resources/wg-quick
cp wireguard_macos/wireguard-go build/resources/wireguard-go

# Pritunl
mkdir -p build/macos/Applications
cd client
npm install
./node_modules/.bin/electron-rebuild
node package.js
cd ../
mv build/macos/Applications/Pritunl-darwin-universal/Pritunl.app build/macos/Applications/
rm -rf build/macos/Applications/Pritunl-darwin-universal

# Service Daemon
mkdir -p build/macos/Library/LaunchDaemons
cp service_macos/com.pritunl.service.plist build/macos/Library/LaunchDaemons

# Privileged Helper Tools
HELPER_DIR=build/macos/Library/PrivilegedHelperTools/pritunl-client
mkdir -p "$HELPER_DIR"
for f in pritunl-service pritunl-openvpn bash wg wg-quick wireguard-go; do
  cp "build/resources/$f" "$HELPER_DIR/$f"
  chmod 0755 "$HELPER_DIR/$f"
done

# Preinstall
echo "###################################################"
echo "Preinstall: Stopping pritunl service..."
echo "###################################################"
kill -2 $(ps aux | grep Pritunl.app | awk '{print $2}') &> /dev/null || true
sudo launchctl unload /Library/LaunchAgents/com.pritunl.client.plist &> /dev/null || true
sudo launchctl bootout system/com.pritunl.service &> /dev/null || true
sudo rm -f /Library/LaunchDaemons/com.pritunl.service.plist

# Install
echo "###################################################"
echo "Installing..."
echo "###################################################"
sudo rm -rf /Applications/Pritunl.app
sudo cp -r build/macos/Applications/Pritunl.app /Applications
sudo rm -rf /Library/PrivilegedHelperTools/pritunl-client
sudo mkdir -p /Library/PrivilegedHelperTools
sudo cp -r build/macos/Library/PrivilegedHelperTools/pritunl-client /Library/PrivilegedHelperTools/pritunl-client
sudo chown -R root:wheel /Library/PrivilegedHelperTools/pritunl-client
sudo chmod 0755 /Library/PrivilegedHelperTools /Library/PrivilegedHelperTools/pritunl-client
sudo cp -f build/macos/Library/LaunchDaemons/com.pritunl.service.plist /Library/LaunchDaemons/com.pritunl.service.plist

# Postinstall
echo "###################################################"
echo "Postinstall: Starting pritunl service..."
echo "###################################################"
sudo xattr -d com.apple.quarantine /Library/LaunchDaemons/com.pritunl.service.plist &> /dev/null || true
sudo launchctl enable system/com.pritunl.service
sudo launchctl bootstrap system /Library/LaunchDaemons/com.pritunl.service.plist

cd ..
rm -rf pritunl-client-$APP_VER
