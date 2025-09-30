#!/bin/bash

set -e

goreleaser build --snapshot --clean
scp dist/ebus-climate_linux_arm_7/ebus-climate glowworm:
ssh glowworm 'mv ebus-climate /usr/local/ebus-climate && sudo systemctl restart ebus-climate'
