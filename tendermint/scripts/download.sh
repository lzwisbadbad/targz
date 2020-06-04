#!/usr/bin/env bash
set -e

# WARN: non hermetic build (people must run this script inside docker to
# produce deterministic binaries).
cd ..
CONTRACT_DIR=./build/contract/
PREFIX="https://"
SUFFIX="/releases/download/"

i=1
for _ in $(cat scripts/bcb.mod)
do
  NUM=$i
  TAG=$(awk 'NR=='$NUM' {print $1}' scripts/bcb.mod)
  VER=$(awk 'NR=='$NUM' {print $2}' scripts/bcb.mod)

  if [[ "$TAG" == "" ]];then
    continue
  fi

  if [[ "$VER" == "go.mod" ]];then
    ii=1
    for _ in $(cat go.mod)
    do
      N=$ii
      GTAG=$(awk 'NR=='$N' {print $1}' go.mod)
      GVER=$(awk 'NR=='$N' {print $2}' go.mod)

      if [[ "$GTAG" == "$TAG" ]];then
        VER="$GVER"
        break
      fi
      : $(( ii++ ))
    done
  fi

  FILENAME="${TAG##*/}"
  DOWNLOAD="$PREFIX$TAG$SUFFIX${VER:1}/$FILENAME""_$VER.tar.gz"

  echo "==> Downloading from ${DOWNLOAD}"
  rm -rf "$CONTRACT_DIR"
  mkdir -p "$CONTRACT_DIR"
  pushd "$CONTRACT_DIR" >/dev/null

  curl -OL "$DOWNLOAD"

  tar -zxf "$FILENAME""_$VER.tar.gz"
  rm -f "$FILENAME""_$VER.tar.gz"
  popd >/dev/null
  : $(( i++ ))
done

cd scripts