#!/bin/bash

echo "Building latest LightBeatGateway..."

# Stash any changes
git stash

# Get latest from github
git pull origin master

# Build Project
/usr/local/go/bin/go build 

# Run!!
./LightBeatGateway


