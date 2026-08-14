#!/bin/bash
set -e

cd "$( dirname "${BASH_SOURCE[0]}" )"
cd ../

# Version is read from the source tree, building before builder.py set-version
# stamps the package and both Go binaries with the previous release version
if grep -q "<%= version %>" CHANGES
then
    echo "ERROR: unreleased version placeholder in CHANGES, run" \
        "tools/builder.py set-version before building" >&2
    exit 1
fi

APP_VER="$(sed -n 's/.*"version": *"\([^"]*\)".*/\1/p' client/package.json | head -n1)"

if [ -z "$APP_VER" ]
then
    echo "ERROR: unable to read version from client/package.json" >&2
    exit 1
fi

for ver in \
    "$(sed -n 's/^Version \([^ ]*\).*/\1/p' CHANGES | head -n1)" \
    "$(sed -n 's/.*"version": *"\([^"]*\)".*/\1/p' client/package-lock.json | head -n1)" \
    "$(sed -n 's/.*Version = "\([^"]*\)".*/\1/p' service/constants/constants.go | head -n1)" \
    "$(sed -n 's/.*Version = "\([^"]*\)".*/\1/p' cli/constants/constants.go | head -n1)" \
    "$(sed -n 's/.*MyAppVersion "\([^"]*\)".*/\1/p' resources_win/setup.iss | head -n1)"
do
    if [ "$ver" != "$APP_VER" ]
    then
        echo "ERROR: source tree version mismatch, client/package.json has" \
            "$APP_VER but found $ver, run tools/builder.py set-version" >&2
        exit 1
    fi
done

export APP_VER
echo "Building version $APP_VER"

rm -rf build
mkdir -p build/resources
go clean -cache

node -e "fs=require('fs');f='client/package.json';c=fs.readFileSync(f,'utf8');fs.writeFileSync(f,c.replace(/,\s*\"scripts\": \{[^}]*\}/,''))"

# Service
cd service
go get
GOOS=darwin GOARCH=amd64 go build -v -o service_x86_64
GOOS=darwin GOARCH=arm64 go build -v -o service_arm64
lipo -create -output service service_x86_64 service_arm64
rm -rf service_x86_64
rm -rf service_arm64
cd ..
cp service/service build/resources/pritunl-service
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" build/resources/pritunl-service

# CLI
cd cli
go get
GOOS=darwin GOARCH=amd64 go build -v -o cli_x86_64
GOOS=darwin GOARCH=arm64 go build -v -o cli_arm64
lipo -create -output cli cli_x86_64 cli_arm64
rm -rf cli_x86_64
rm -rf cli_arm64
cd ..
mkdir -p build/resources
cp cli/cli build/resources/pritunl-client
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" build/resources/pritunl-client

# Device Auth
cd service_macos
rm -f "Pritunl Device Authentication"
swiftc -sdk $(xcrun --show-sdk-path --sdk macosx) -target arm64-apple-macos11 -framework CryptoKit -framework LocalAuthentication -framework Security -framework Foundation device_auth.swift -o "Pritunl Device Authentication_arm64"
swiftc -sdk $(xcrun --show-sdk-path --sdk macosx) -target x86_64-apple-macos11 -framework CryptoKit -framework LocalAuthentication -framework Security -framework Foundation device_auth.swift -o "Pritunl Device Authentication_x86_64"
lipo -create -output "Pritunl Device Authentication" "Pritunl Device Authentication_arm64" "Pritunl Device Authentication_x86_64"
rm -rf Pritunl\ Device\ Authentication_arm64
rm -rf Pritunl\ Device\ Authentication_x86_64
cp "Pritunl Device Authentication" "../build/resources/Pritunl Device Authentication"
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" "Pritunl Device Authentication"
cd ..

# Service Helper
cd service_macos
rm -f pritunl-service-helper
swiftc -sdk $(xcrun --show-sdk-path --sdk macosx) -target arm64-apple-macos13 -framework ServiceManagement -framework Foundation service_helper.swift -o pritunl-service-helper_arm64
swiftc -sdk $(xcrun --show-sdk-path --sdk macosx) -target x86_64-apple-macos13 -framework ServiceManagement -framework Foundation service_helper.swift -o pritunl-service-helper_x86_64
lipo -create -output pritunl-service-helper pritunl-service-helper_arm64 pritunl-service-helper_x86_64
rm -rf pritunl-service-helper_arm64
rm -rf pritunl-service-helper_x86_64
cp pritunl-service-helper ../build/resources/pritunl-service-helper
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" ../build/resources/pritunl-service-helper
cd ..

# Openvpn
cp openvpn_macos/openvpn10 build/resources/pritunl-openvpn10
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" build/resources/pritunl-openvpn10
lipo -create -output openvpn_macos/openvpn_universal openvpn_macos/openvpn openvpn_macos/openvpn_arm64
cp openvpn_macos/openvpn_universal build/resources/pritunl-openvpn
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" build/resources/pritunl-openvpn
rm -rf openvpn_macos/openvpn_universal

# WireGuard
cp wireguard_macos/bash build/resources/bash
cp wireguard_macos/wg build/resources/wg
cp wireguard_macos/wg-quick build/resources/wg-quick
cp wireguard_macos/wireguard-go build/resources/wireguard-go
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" build/resources/bash
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" build/resources/wg
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" build/resources/wg-quick
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" build/resources/wireguard-go

# Pritunl
mkdir -p build/macos/Applications
cd client
npm install
./node_modules/.bin/electron-rebuild
node package.js

cd ../
mv build/macos/Applications/Pritunl-darwin-universal/Pritunl.app build/macos/Applications/
rm -rf build/macos/Applications/Pritunl-darwin-universal
sleep 3
npx @electron/fuses read --app build/macos/Applications/Pritunl.app

# Service Daemon
mkdir -p build/macos/Library/LaunchDaemons
cp service_macos/com.pritunl.service.plist build/macos/Library/LaunchDaemons

# Package
find build/macos -type f -exec xattr -c {} \;
xattr -c resources_macos/scripts/postinstall
chmod +x resources_macos/scripts/postinstall
xattr -c resources_macos/scripts/preinstall
chmod +x resources_macos/scripts/preinstall
cd build
pkgbuild --root macos --scripts ../resources_macos/scripts --sign "Developer ID Installer: Pritunl, Inc. (U22BLATN63)" --identifier com.pritunl.pkg.Pritunl --version $APP_VER --ownership recommended --install-location / Build.pkg
productbuild --resources ../resources_macos --distribution ../resources_macos/distribution.xml --sign "Developer ID Installer: Pritunl, Inc. (U22BLATN63)" --version $APP_VER Pritunl.pkg
zip Pritunl.pkg.zip Pritunl.pkg
rm -f Build.pkg

# Notarize
xcrun notarytool submit Pritunl.pkg --keychain-profile "Pritunl" --apple-id "contact@pritunl.com" --team-id U22BLATN63 --output-format json
sleep 10
xcrun notarytool history --keychain-profile "Pritunl" --apple-id "contact@pritunl.com" --team-id U22BLATN63
