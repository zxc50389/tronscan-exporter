#!/bin/sh

# Stage build-time metadata for packaging. This exporter has no runtime assets
# beyond the compiled binary, so we only record the version being built.
cd "$(dirname "$0")"
mkdir -p workdir

echo -n "${CI_COMMIT_TAG}" > ./workdir/version

ls -al workdir
