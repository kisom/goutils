#!/bin/bash

set -euxo pipefail

BAZEL="bazel"
if [ -z "$(command -v ${BAZEL})" ]
then
	BAZEL="bazelisk"
fi

$BAZEL run //:gazelle
$BAZEL run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%go_dependencies
$BAZEL run //:gazelle

