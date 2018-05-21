#!/bin/bash

# Borrowing goiardi's package deploy script
# NB: not useful until circleci and packagecloud is set up for spqr

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

CURDIR=`pwd`
GOIARDI_VERSION=`git describe --long --always`

# Requires the package_cloud gem to be installed. Do so with:
# gem install package_cloud -v "0.2.43"

if [ -z ${PACKAGECLOUD_REPO} ] ; then
  echo "The environment variable PACKAGECLOUD_REPO must be set."
  exit 1
fi

# debian jessie and stretch (note: add buster later)
for DEB_REL in jessie stretch; do
	package_cloud push ${PACKAGECLOUD_REPO}/debian/${DEB_REL} ${DIR}/artifacts/jessie/*.deb
done

# modern systemd based ubuntus. Passing on trusty for now, but if there ends up
# being a need that can be revisited.
for UBUNTU_REL in xenial artful bionic cosmic; do
	package_cloud push ${PACKAGECLOUD_REPO}/ubuntu/${UBUNTU_REL} ${DIR}/artifacts/jessie/*.deb
done

# centos

package_cloud push ${PACKAGECLOUD_REPO}/el/6 ${DIR}/artifacts/el6/*.rpm
package_cloud push ${PACKAGECLOUD_REPO}/el/7 ${DIR}/artifacts/el7/*.rpm
