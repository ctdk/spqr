#!/bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# resync tags, if needed
git pull --tags

CURDIR=`pwd`
ARTIFACT_DIR=$CURDIR/artifacts
GOIARDI_VERSION=`git describe --long --always`
GIT_HASH=`git rev-parse --short HEAD`
COMMON_DIR="$CURDIR/common"
BUILD="$CURDIR/build"

rm -r $BUILD
rm -r $ARTIFACT_DIR

for VAR in jessie el6 el7; do
	mkdir -p $ARTIFACT_DIR/$VAR
done

mkdir -p $BUILD/bin

cd ..
go build


