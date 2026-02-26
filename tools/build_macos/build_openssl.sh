#!/bin/bash
set -e

export CFLAGS="-mmacosx-version-min=11.0"
export CXXFLAGS="-mmacosx-version-min=11.0"
export CPPFLAGS="-mmacosx-version-min=11.0"
export LINKFLAGS="-mmacosx-version-min=11.0"

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
