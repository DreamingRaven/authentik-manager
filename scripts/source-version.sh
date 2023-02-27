#!/usr/bin/env bash
# Please ensure this is run from the root directory not from the directoyr it exists
echo "$(cat charts/ak/values.yaml | grep -P -o '(?<=ghcr.io/goauthentik/server:).*(?=\")')"
