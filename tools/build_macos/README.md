###################################################################
# openvpn
###################################################################
mkdir -p /Users/apple/build
cp ./* /Users/apple/build/
cp ../../openvpn_macos/* /Users/apple/build/
cd /Users/apple/build

tar xf m4-1.4.21.tar.xz
cd ./m4-1.4.21
sh ../build_m4.sh

tar xf autoconf-2.72.tar.gz
cd ./autoconf-2.72
sh ../build_autoconf.sh
cd ../

tar xf automake-1.16.5.tar.xz
cd ./automake-1.16.5
sh ../build_automake.sh
cd ../

tar xf libtool-2.4.7.tar.xz
cd ./libtool-2.4.7
sh ../build_libtool.sh
cd ../

tar xf lz4-1.10.0.tar.gz
cd ./lz4-1.10.0
sh ../build_lz4.sh
cd ../

tar xf lzo-2.10.tar.gz
cd ./lzo-2.10
sh ../build_lzo.sh
cd ../

tar xf openssl-3.3.6.tar.gz
cd ./openssl-3.3.6
sh ../build_openssl.sh
cd ../

tar xf pkcs11-helper-1.31.0.tar.bz2
cd ./pkcs11-helper-1.31.0
sh ../build_pkcs11h.sh
cd ../

tar xf openvpn-2.7.0.tar.gz
cd ./openvpn-2.7.0
sh ../build_openvpn.sh
cd ../


codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" ./openvpn/sbin/openvpn

###################################################################
# wireguard
###################################################################
mkdir -p /Users/apple/build
cp ../../wireguard_macos/* /Users/apple/build/
cd /Users/apple/build

export MACOSX_DEPLOYMENT_TARGET=11.0
export CFLAGS="-mmacosx-version-min=11.0"
export CXXFLAGS="-mmacosx-version-min=11.0"
export CPPFLAGS="-mmacosx-version-min=11.0"
export LINKFLAGS="-mmacosx-version-min=11.0"

rm -rf ./bash-5.2
tar xf bash-5.2.tar.gz
cd ./bash-5.2
./configure
make
cp ./bash ../bash-arm64
cd ../

export MACOSX_DEPLOYMENT_TARGET=11.0
export CFLAGS="-mmacosx-version-min=11.0"
export CXXFLAGS="-mmacosx-version-min=11.0"
export CPPFLAGS="-mmacosx-version-min=11.0"
export LINKFLAGS="-mmacosx-version-min=11.0"

rm -rf ./wireguard-go-0.0.20250522
tar xf wireguard-go-0.0.20250522.tar.gz
cd ./wireguard-go-0.0.20250522
make
cp ./wireguard-go ../wireguard-go-arm64
cd ../

export MACOSX_DEPLOYMENT_TARGET=11.0
export CFLAGS="-mmacosx-version-min=11.0"
export CXXFLAGS="-mmacosx-version-min=11.0"
export CPPFLAGS="-mmacosx-version-min=11.0"
export LINKFLAGS="-mmacosx-version-min=11.0"

rm -rf ./wireguard-tools-1.0.20260223
tar xf wireguard-tools-1.0.20260223.tar.gz
cd ./wireguard-tools-1.0.20260223
make -C src WITH_WGQUICK=yes
cp ./src/wg ../wg-arm64
cp ./src/wg-quick/darwin.bash ../wg-quick
cd ../

export MACOSX_DEPLOYMENT_TARGET=11.0
export CFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export CXXFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export CPPFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export LDFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export LINKFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export GOOS=darwin
export GOARCH=amd64

rm -rf ./bash-5.2
tar xf bash-5.2.tar.gz
cd ./bash-5.2
./configure --host=x86_64-apple-darwin
make
cp ./bash ../bash-amd64
cd ../

export MACOSX_DEPLOYMENT_TARGET=11.0
export CFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export CXXFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export CPPFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export LDFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export LINKFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export GOOS=darwin
export GOARCH=amd64

rm -rf ./wireguard-go-0.0.20250522
tar xf wireguard-go-0.0.20250522.tar.gz
cd ./wireguard-go-0.0.20250522
make
cp ./wireguard-go ../wireguard-go-amd64
cd ../

export MACOSX_DEPLOYMENT_TARGET=11.0
export CFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export CXXFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export CPPFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export LDFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export LINKFLAGS="-arch x86_64 -mmacosx-version-min=11.0"
export GOOS=darwin
export GOARCH=amd64

rm -rf ./wireguard-tools-1.0.20260223
tar xf wireguard-tools-1.0.20260223.tar.gz
cd ./wireguard-tools-1.0.20260223
make -C src WITH_WGQUICK=yes
cp ./src/wg ../wg-amd64
cd ../

lipo -info bash-amd64
lipo -info bash-arm64
lipo -create -output bash bash-amd64 bash-arm64

lipo -info wireguard-go-amd64
lipo -info wireguard-go-arm64
lipo -create -output wireguard-go wireguard-go-amd64 wireguard-go-arm64

lipo -info wg-amd64
lipo -info wg-arm64
lipo -create -output wg wg-amd64 wg-arm64

codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" ./bash
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" ./wireguard-go
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" ./wg
codesign --force --timestamp --options=runtime -s "Developer ID Application: Pritunl, Inc. (U22BLATN63)" ./wg-quick
