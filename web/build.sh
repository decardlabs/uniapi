#!/bin/sh

set -eu

version=$(date +%s)
echo "Building theme: modern"

mkdir -p build
rm -rf build/modern

cd modern
yarn install --frozen-lockfile
REACT_APP_VERSION=$version yarn build
