#!/bin/bash

set -e

TARFILE=\"$npm_package_name\"_$npm_package_version.orig.tar.gz
DEBFILE=\"$npm_package_name\"_v$npm_package_version.deb

mkdir -p dist/app

dpkg -x dist/"$DEBFILE" ./dist/app/
rm -r dist/app/opt/Lumerin\ Wallet/resources/app.asar.unpacked/node_modules/7zip-bin-linux/arm*
perl -pi -e 's/opt/usr\\/lib/' dist/app/usr/share/applications/morpheus-ui-desktop.desktop
tar -czf ppa/"$TARFILE" -C dist/ app/
rm -r dist/app

cd ppa
tar -xzf "$TARFILE"

cd app
dch -b -v "$npm_package_version"-1xenial --distribution xenial Xenial release "$npm_package_version"
debuild -us -uc
dch -b -v "$npm_package_version"-1artful --distribution artful Artful release "$npm_package_version"
debuild -us -uc
dch -b -v "$npm_package_version"-1bionic --distribution bionic Bionic release "$npm_package_version"
debuild -us -uc
