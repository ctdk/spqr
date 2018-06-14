#!/bin/bash

# Adapting goiardi's packaging script.

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# resync tags, if needed
git pull --tags

CURDIR=`pwd`
ARTIFACT_DIR=$CURDIR/artifacts
SPQR_VERSION=`git describe --long --always`
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
go build -ldflags "-X github.com/ctdk/spqr/config.GitHash=$GIT_HASH" -o $BUILD/bin

# Get this one out of the way first, since I somewhat strangely need the CentOS
# 6 / Amazon Linux(v1) package before anything else.

CENTOS_COMMON_DIR="$CURDIR/centos/common"
CENTOS_SCRIPTS="$CURDIR/centos/scripts"

BUILD_ROOT="$BUILD/el6"
FILES_DIR="$CURDIR/centos/6"
mkdir -p $BUILD_ROOT
cd $BUILD_ROOT
mkdir -p usr/sbin
mkdir -p var/lib/spqr
cp $BUILD/bin/spqr usr/sbin/spqr
cp -r $FILES_DIR/fs/etc .
cp -r $COMMON_DIR/* .
cp -r $CENTOS_COMMON_DIR/etc .

fpm -s dir -t rpm -n spqr -v $SPQR_VERSION -C . -p $ARTIFACT_DIR/el6/spqr-VERSION.el6.ARCH.rpm -a amd64 --description "a small consul based user management utility" --license apachev2 -m "<jeremy@goiardi.gl>" .

BUILD_ROOT="$BUILD/el7"
FILES_DIR="$CURDIR/debian/jessie"
mkdir -p $BUILD_ROOT
cd $BUILD_ROOT
mkdir -p usr/sbin
mkdir -p var/lib/spqr
cp $BUILD/bin/spqr usr/sbin/spqr
cp -r $FILES_DIR/fs/lib .
cp -r $COMMON_DIR/* .
cp -r $CENTOS_COMMON_DIR/etc .

fpm -s dir -t rpm -n spqr -v $SPQR_VERSION -C . -p $ARTIFACT_DIR/el6/spqr-VERSION.el7.ARCH.rpm -a amd64 --description "a small consul based user management utility" --license apachev2 -m "<jeremy@goiardi.gl>" .

BUILD_ROOT="$BUILD/jessie"
FILES_DIR="$CURDIR/debian/jessie"
mkdir -p $BUILD_ROOT
cd $BUILD_ROOT
mkdir -p usr/sbin
mkdir -p var/lib/spqr
cp $BUILD/bin/spqr usr/sbin/spqr
cp -r $FILES_DIR/fs/lib .
cp -r $COMMON_DIR/* .

fpm -s dir -t deb -n spqr -v $SPQR_VERSION -C . -p $ARTIFACT_DIR/el6/spqr-VERSION_ARCH.deb -a amd64 --description "a small consul based user management utility" --after-remove $FILES_DIR/scripts/postrm.sh --deb-suggests consul --license apachev2 -m "<jeremy@goiardi.gl>" .
