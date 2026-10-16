#!/usr/bin/env bash

ANSI_RED='\033[0;31m'
ANSI_GREEN='\033[0;32m'
ANSI_YELLOW='\033[0;33m'
ANSI_RESET='\033[0m'

export GOEXPERIMENT="jsonv2"

binary_path="builds/EDCITR"

echo "Starting build script"


if [ -f $binary_path ]; then
  echo -e "${ANSI_YELLOW}Removing previous build${ANSI_RESET}"
  rm $binary_path
fi

echo "Building..."

go build -trimpath -ldflags "-w -s" -o "$binary_path" ./cmd/EDCITR/

if [ $? -eq 0 ]; then
  echo -e "${ANSI_GREEN}Build done${ANSI_RESET}"
else
  echo -e "${ANSI_RED}Build failed${ANSI_RESET}"
fi



