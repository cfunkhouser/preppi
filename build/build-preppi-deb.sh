#!/bin/bash
set -e

WORKDIR="/opt/gopath/src/github.com/cfunkhouser/preppi"
BUILDDIR="${WORKDIR}/build/out"

PREPPI_VERSION=$(cat "${BUILDDIR}/VERSION")
PREPPI_ARCH="armhf"
PREPPI_SIZE=$(stat -c "%s" "${BUILDDIR}/bin/preppi-linux-armv7")

cd "${WORKDIR}"
mkdir -p "${BUILDDIR}/package/preppi"

# Set up the Debian package metadata and file structure
cp -r "${WORKDIR}/build/package/"* "${BUILDDIR}/package/preppi"
sed -i'' "s/%VERSION%/${PREPPI_VERSION}/g" ${BUILDDIR}/package/preppi/DEBIAN/control
sed -i'' "s/%ARCH%/${PREPPI_ARCH}/g" ${BUILDDIR}/package/preppi/DEBIAN/control
sed -i'' "s/%SIZE%/${PREPPI_SIZE}/g" ${BUILDDIR}/package/preppi/DEBIAN/control

# Create /usr/local/bin in the package directory, and copy the preppi binary
mkdir -p "${BUILDDIR}/package/preppi/usr/local/bin"
cp "${BUILDDIR}/bin/preppi-linux-armv7" "${BUILDDIR}/package/preppi/usr/local/bin/preppi"

# Create the package
pushd "${BUILDDIR}/package" && dpkg-deb --build preppi ; popd

# Rename and reparent the assembled package
mv -v "${BUILDDIR}/package/preppi.deb" "${BUILDDIR}/preppi-${PREPPI_VERSION}-${PREPPI_ARCH}.deb"

# Clean up to make sure we don't accidentally package the wrong binary for the
# arch in subsequent builds.
rm -rf "${BUILDDIR}/package/preppi"