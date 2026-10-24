#!/bin/bash
set -e

export MACOSX_DEPLOYMENT_TARGET="13.0"
export CFLAGS="-O2 -arch arm64 -mmacosx-version-min=13.0"
export CXXFLAGS="$CFLAGS"
export CPPFLAGS="-arch arm64 -mmacosx-version-min=13.0"
export LDFLAGS="-arch arm64 -mmacosx-version-min=13.0"
export OPENSSL_CFLAGS="-I/Users/apple/build/openssl/include"
export OPENSSL_LIBS="/Users/apple/build/openssl/lib/libssl.a /Users/apple/build/openssl/lib/libcrypto.a"

#autoreconf --verbose --install --force
./configure \
  --disable-debug \
  --disable-dependency-tracking \
  --disable-threading \
  --disable-slotevent \
  --disable-crypto-engine-gnutls \
  --disable-crypto-engine-nss \
  --disable-crypto-engine-mbedtls \
  --disable-shared \
  --enable-static \
  --prefix=/Users/apple/build/pkcs11-helper
make install
