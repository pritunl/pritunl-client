#!/bin/bash
set -e

APP_VER="$(curl -s https://api.github.com/repos/pritunl/pritunl-client/releases/latest | python -c 'import json,sys;print(json.load(sys.stdin)["tag_name"])')"

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

# Pritunl
mkdir -p build/macos/Applications
cd client
npm install
./node_modules/.bin/electron-rebuild
node package.js
cd ../
mv build/macos/Applications/Pritunl-darwin-x64/Pritunl.app build/macos/Applications/
rm -rf build/macos/Applications/Pritunl-darwin-x64

# Service
cd service
GOPATH="$(pwd)/go" go get -d
GOPATH="$(pwd)/go" go build -v
cd ..
cp service/service build/macos/Applications/Pritunl.app/Contents/Resources/pritunl-service

# Service Daemon
mkdir -p build/macos/Library/LaunchDaemons
cp service_macos/com.pritunl.service.plist build/macos/Library/LaunchDaemons

# Device Authentication
cd service_macos
rm -f "Pritunl Device Authentication"
swiftc -sdk $(xcrun --show-sdk-path --sdk macosx) -framework CryptoKit -framework LocalAuthentication -framework Security -framework Foundation device_auth.swift -o "Pritunl Device Authentication"
cp "./Pritunl Device Authentication" build/macos/Applications/Pritunl.app/Contents/Resources/
cd ..

# Openvpn
cp openvpn_macos/openvpn build/macos/Applications/Pritunl.app/Contents/Resources/pritunl-openvpn
cp openvpn_macos/openvpn10 build/macos/Applications/Pritunl.app/Contents/Resources/pritunl-openvpn10

# WireGuard
cp wireguard_macos/bash build/macos/Applications/Pritunl.app/Contents/Resources/bash
cp wireguard_macos/wg build/macos/Applications/Pritunl.app/Contents/Resources/wg
cp wireguard_macos/wg-quick build/macos/Applications/Pritunl.app/Contents/Resources/wg-quick
cp wireguard_macos/wireguard-go build/macos/Applications/Pritunl.app/Contents/Resources/wireguard-go

# CLI
cd cli
GOPATH="$(pwd)/go" go get -d
GOPATH="$(pwd)/go" go build -v
cd ..
cp cli/cli build/macos/Applications/Pritunl.app/Contents/Resources/pritunl-client

# Files
sudo touch /var/run/pritunl_auth
sudo touch /var/log/pritunl.log
sudo touch /var/log/pritunl.log.1

# Preinstall
echo "###################################################"
echo "Preinstall: Stopping pritunl service..."
echo "###################################################"
kill -2 $(ps aux | grep Pritunl.app | awk '{print $2}') &> /dev/null || true
sudo launchctl unload /Library/LaunchAgents/com.pritunl.client.plist &> /dev/null || true
sudo launchctl unload /Library/LaunchDaemons/com.pritunl.service.plist &> /dev/null || true

# Install
echo "###################################################"
echo "Installing..."
echo "###################################################"
sudo rm -rf /Applications/Pritunl.app
sudo cp -r build/macos/Applications/Pritunl.app /Applications
sudo cp -f build/macos/Library/LaunchDaemons/com.pritunl.service.plist /Library/LaunchDaemons/com.pritunl.service.plist

# Postinstall
echo "###################################################"
echo "Postinstall: Starting pritunl service..."
echo "###################################################"
sudo launchctl enable system/com.pritunl.service
sudo launchctl load /Library/LaunchDaemons/com.pritunl.service.plist

cd ..
rm -rf pritunl-client-$APP_VER
