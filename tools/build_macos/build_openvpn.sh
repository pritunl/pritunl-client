#!/bin/bash
set -e

export MACOSX_DEPLOYMENT_TARGET="13.0"
export CFLAGS="-O2 -arch arm64 -mmacosx-version-min=13.0"
export CXXFLAGS="$CFLAGS"
export CPPFLAGS="-arch arm64 -mmacosx-version-min=13.0"
export LDFLAGS="-arch arm64 -mmacosx-version-min=13.0"

OPENSSL_CFLAGS="-I/Users/apple/build/openssl/include" \
  OPENSSL_SSL_CFLAGS="-I/Users/apple/build/openssl/include" \
  OPENSSL_CRYPTO_CFLAGS="-I/Users/apple/build/openssl/include" \
  OPENSSL_LIBS="/Users/apple/build/openssl/lib/libssl.a /Users/apple/build/openssl/lib/libcrypto.a" \
  OPENSSL_SSL_LIBS="/Users/apple/build/openssl/lib/libssl.a" \
  OPENSSL_CRYPTO_LIBS="/Users/apple/build/openssl/lib/libcrypto.a" \
  PKCS11_HELPER_CFLAGS="-I/Users/apple/build/pkcs11-helper/include" \
  PKCS11_HELPER_LIBS="-L/Users/apple/build/pkcs11-helper/lib -lpkcs11-helper" \
  LZO_CFLAGS="-I/Users/apple/build/lzo/include" \
  LZO_LIBS="/Users/apple/build/lzo/lib/liblzo2.a" \
  OPTIONAL_LZO_LIBS="/Users/apple/build/lzo/lib/liblzo2.a" \
  LZ4_CFLAGS="-I/Users/apple/build/lz4/include" \
  LZ4_LIBS="/Users/apple/build/lz4/lib/liblz4.a" \
  ./configure \
    --disable-debug \
    --disable-dependency-tracking \
    --disable-silent-rules \
    --disable-management \
    --disable-plugins \
    --disable-plugin-auth-pam \
    --disable-plugin-down-root \
    --disable-lz4 \
    --with-crypto-library=openssl \
    --enable-pkcs11 \
    --enable-static \
    --disable-shared \
    --prefix=/Users/apple/build/openvpn
make install
