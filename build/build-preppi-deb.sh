#!/bin/bash
set -e

WORKDIR="/opt/gopath/src/github.com/cfunkhouser/preppi"
ARCHS="$@"
if [ "${ARCHS}" == "" ] ; then
  ARCHS="amd64 armhf i386"
fi

BUILDDIR="${WORKDIR}/build/out"
BUILDID=$(cat "${BUILDDIR}/VERSION")

debpackagearch () {
  echo $@
  local PREPPI_ARCH="${1}"
  if [ "${PREPPI_ARCH}" == "" ] ; then
    echo "No arch specificied, which is nonsense." 1>&2
    return 1
  fi

  local PREPPI_BIN_PATH="${BUILDDIR}/bin/preppi-linux-${PREPPI_ARCH}"
  if [ ! -f "${PREPPI_BIN_PATH}" ] ; then 
    echo "Couldn't find the prebuilt binary for ${PREPPI_ARCH}" 1>&2
    return 1
  fi  
  local PREPPI_SIZE=$(stat -c %s "${PREPPI_BIN_PATH}")
  local PREPPI_PKGDIR="${BUILDDIR}/package/${PREPPI_ARCH}/preppi"

  # Set up the Debian package metadata and file structure
  mkdir -p "${PREPPI_PKGDIR}"
  cp -rv "${WORKDIR}/build/package/"* "${PREPPI_PKGDIR}"
  sed -i'' "s/%VERSION%/${BUILDID}/g" "${PREPPI_PKGDIR}/DEBIAN/control"
  sed -i'' "s/%ARCH%/${PREPPI_ARCH}/g" "${PREPPI_PKGDIR}/DEBIAN/control"
  sed -i'' "s/%SIZE%/${PREPPI_SIZE}/g" "${PREPPI_PKGDIR}/DEBIAN/control"

  # Create /usr/local/bin in the package directory, and copy the preppi binary
  mkdir -p "${PREPPI_PKGDIR}/usr/local/bin"
  cp "${PREPPI_BIN_PATH}" "${PREPPI_PKGDIR}/usr/local/bin/preppi"

  # Create the package
  pushd "${BUILDDIR}/package/${PREPPI_ARCH}" && dpkg-deb --build preppi ; popd

  # Rename and reparent the assembled package
  mv -v "${PREPPI_PKGDIR}.deb" "${BUILDDIR}/preppi-${BUILDID}-${PREPPI_ARCH}.deb"

  # Clean up to make sure we don't accidentally package the wrong binary for the
  # arch in subsequent builds.
  rm -rf "${PREPPI_PKGDIR}"
}

for ARCH in ${ARCHS} ; do
  debpackagearch "${ARCH}"
done
