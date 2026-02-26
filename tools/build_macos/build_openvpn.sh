#!/bin/bash
set -e

export CFLAGS="-mmacosx-version-min=11.0"
export CXXFLAGS="-mmacosx-version-min=11.0"
export CPPFLAGS="-mmacosx-version-min=11.0"
export LINKFLAGS="-mmacosx-version-min=11.0"

OPENSSL_CFLAGS="-I/Users/apple/build/openssl/include" \
  OPENSSL_SSL_CFLAGS="-I/Users/apple/build/openssl/include" \
  OPENSSL_CRYPTO_CFLAGS="-I/Users/apple/build/openssl/include" \
  OPENSSL_LIBS="/Users/apple/build/openssl/lib/libssl.a -lz /Users/apple/build/openssl/lib/libcrypto.a -lz" \
  OPENSSL_SSL_LIBS="/Users/apple/build/openssl/lib/libssl.a" \
  OPENSSL_CRYPTO_LIBS="/Users/apple/build/openssl/lib/libcrypto.a -lz" \
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
    --disable-server \
    --disable-management \
    --disable-plugins \
    --disable-plugin-auth-pam \
    --disable-plugin-down-root \
    --disable-lz4 \
    --with-crypto-library=openssl \
    --build=arm-apple-darwin \
    --enable-pkcs11 \
    --enable-static \
    --disable-shared \
    --prefix=/Users/apple/build/openvpn
make install
