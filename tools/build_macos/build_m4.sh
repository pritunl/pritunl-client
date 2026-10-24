#!/bin/bash
set -e

export MACOSX_DEPLOYMENT_TARGET="13.0"
export CFLAGS="-O2 -arch arm64 -mmacosx-version-min=13.0"
export CXXFLAGS="$CFLAGS"
export CPPFLAGS="-arch arm64 -mmacosx-version-min=13.0"
export LDFLAGS="-arch arm64 -mmacosx-version-min=13.0"

./configure
make
sudo make install
