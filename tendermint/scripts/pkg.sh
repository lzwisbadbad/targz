#!/bin/bash

cd ..
# Delete the old dir
echo "==> Removing old directory..."
rm -rf build/pkg
mkdir -p build/pkg

# Get the git commit
GIT_COMMIT="$(git rev-parse --short=8 HEAD)"
GIT_IMPORT="github.com/bcbchain/tendermint/version"

# Determine the arch/os combos we're building for
XC_ARCH=${XC_ARCH:-"amd64"}   # 386 arm
XC_OS=${XC_OS:-"solaris darwin freebsd linux windows"}
XC_EXCLUDE=${XC_EXCLUDE:-" darwin/arm solaris/amd64 solaris/386 solaris/arm freebsd/amd64 windows/arm "}

# Make sure build tools are available.
#make get_tools

# Get VENDORED dependencies
#make get_vendor_deps

# copy bundle files
echo "==> Copying..."
for PKG in darwin_amd64 linux_amd64 windows_amd64
do
  echo "copy setup 2 $PKG..."
  cp -rf bundle/setup/ build/pkg/"$PKG"
  for CHAINID in $(find ./bundle/genesis -mindepth 1 -maxdepth 1 -type d); do
    CHAIN=$(basename "$CHAINID")
    echo "$CHAIN"
	  cp -rf "./bundle/genesis/${CHAIN}" "./build/pkg/$PKG/pieces/"

	  if [ -d "./build/.config/${CHAIN}" ];then
	    cp -rf "./bundle/.config/${CHAIN}" "./build/pkg/$PKG/pieces/"
	  fi

	  if [ -d "./build/contract/${CHAIN}" ];then
	    cp -rf "./build/contract/${CHAIN}" "./build/pkg/$PKG/pieces/"
	  fi
  done
done

# Build!
# ldflags: -s Omit the symbol table and debug information.
#	         -w Omit the DWARF symbol table.
echo "==> Building..."
IFS=' ' read -ra arch_list <<< "$XC_ARCH"
IFS=' ' read -ra os_list <<< "$XC_OS"
for arch in "${arch_list[@]}"; do
	for os in "${os_list[@]}"; do
		if [[ "$XC_EXCLUDE" !=  *" $os/$arch "* ]]; then
			echo "--> $os/$arch"
			GOOS=${os} GOARCH=${arch} go build -ldflags "-s -w -X ${GIT_IMPORT}.GitCommit=${GIT_COMMIT}" -tags="${BUILD_TAGS}" -o "build/pkg/${os}_${arch}/pieces/tendermint" ./cmd/tendermint
		  go build -ldflags "-s -w" -tags="${BUILD_TAGS}" -o "build/pkg/${os}_${arch}/pieces/p2p_ping" ./cmd/p2p_ping
		fi
	done
done

# Zip all the files.
echo "==> Packaging..."
for PLATFORM in $(find ./build/pkg -mindepth 1 -maxdepth 1 -type d); do
	OSARCH=$(basename "${PLATFORM}")
	echo "--> ${OSARCH}"

	pushd "$PLATFORM" >/dev/null 2>&1
	tar -zcf "../${OSARCH}.tar.gz" ./*
	popd >/dev/null 2>&1
done

# Add "tendermint" and $VERSION prefix to package name.
rm -rf ./build/dist
mkdir -p ./build/dist
for FILENAME in $(find ./build/pkg -mindepth 1 -maxdepth 1 -type f); do
  FILENAME=$(basename "$FILENAME")
	cp "./build/pkg/${FILENAME}" "./build/dist/tendermint_${VERSION}_${FILENAME}"
done

# Make the checksums.
pushd ./build/dist >/dev/null 2>&1
shasum -a256 ./* > "./tendermint_${VERSION}_SHA256SUMS"
popd >/dev/null 2>&1

# Done
echo
echo "==> Results:"
ls -hl ./build/dist