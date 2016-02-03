#!/bin/bash

# Build Go sources
GOOS=linux GOARCH=amd64 go build

# Copy binary
sudo cp spatch deb/spatch/usr/sbin/spatch

# Update ownership
sudo chown -R root deb/spatch/*

# Create deb file
dpkg-deb --build deb/spatch

# Upload it
curl -X POST https://byt.tl/upload -F file=@deb/spatch.deb
