#!/bin/sh

cd "$(dirname "$0")"
pwd
mkdir -p workdir

cp -ap ../src/config/default-production-bac.conf ./workdir/

cp -ap ../src/html ./workdir/
cp -ap ../src/lang ./workdir/
cp -ap ../src/GeoLite2-City.mmdb ./workdir/

echo -n "${CI_COMMIT_TAG}" > ./workdir/version


ls -al workdir
