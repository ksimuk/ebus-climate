#!/bin/bash

set -e

goreleaser build --snapshot --clean
scp dist/ebus-climate_linux_arm_7/ebus-climate glowworm:ebus-climate-new
ssh glowworm 'mv ebus-climate-new ebus-climate'
