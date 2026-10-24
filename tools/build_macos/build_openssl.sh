#!/bin/bash
set -e

export MACOSX_DEPLOYMENT_TARGET="13.0"
export CFLAGS="-O2 -arch arm64 -mmacosx-version-min=13.0"
export CXXFLAGS="$CFLAGS"
export CPPFLAGS="-arch arm64 -mmacosx-version-min=13.0"
export LDFLAGS="-arch arm64 -mmacosx-version-min=13.0"

unset OPENSSL_LOCAL_CONFIG_DIR

perl ./Configure \
  darwin64-arm64-cc \
  enable-ec_nistp_64_gcc_128 \
  no-ssl3 \
  no-ssl3-method \
  no-zlib \
  no-shared \
  --openssldir=etc/"openssl@3" \
  --prefix=/Users/apple/build/openssl
make
make test
make install
