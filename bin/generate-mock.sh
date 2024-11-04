#!/usr/bin/env bash

set -euo pipefail

# Loop through all .go files except those in proto and vendor directories
for file in $(find . -name '*.go' | grep -v proto | grep -v /vendor/); do
    # Check if the file contains an interface definition
    if grep -q '^type.*interface {$' "${file}"; then
        # Set the destination path and create the directory if it doesn't exist
        dest="test/mock/${file//internal\//}"
        mkdir -p "$(dirname "${dest}")"
        
        # Generate the mock
        mockgen -source="${file}" -destination="${dest}"
    fi
done
