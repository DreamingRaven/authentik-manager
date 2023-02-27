#!/usr/bin/env bash
# Please ensure this is run from the root directory not from the directoyr it exists
echo "$(set -o pipefail && git describe --long 2>/dev/null | sed 's/\([^-]*-g\)/r\1/;s/-/./g' || printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)")"
