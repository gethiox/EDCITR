#!/usr/bin/env bash

declare -a all_required_commands=(
  "ffmpeg"
  "find"
  "xargs"
  "rename"
)

declare -i fail=0

for required_command in "${all_required_commands[@]}"; do
  if ! command -v "$required_command" &>/dev/null; then
    fail=1
    echo "required command '$required_command' is not available."
  fi
done

if [[ "$fail" -ne 0 ]]; then
  echo "aborting."
  exit 1
fi

if [[ "$#" -ne 1 ]]; then
  echo "usage:"
  echo "  directory path argument with sound files inside required."
  echo
  echo "  WARNING: action is destructive, source files will be removed if provided in output (.ogg) format,"
  echo "  make sure you have backup copy of your sound/video collection before proceeding!"

  exit 1
fi

echo "converting files..."
find "$1" -type f -print0 | xargs -0 -P 32 -I"{}" ffmpeg -i "{}" -loglevel warning -y -vn -filter:a "loudnorm=I=-12:LRA=1.0:TP=-1.0" -ac 1 -ar 22050 -c:a libvorbis -q:a 0 "{}-normalized.ogg"
echo "conversion finished"

echo "removing '-normalized' mark from converted filenames..."
find "$1" -type f -name "*-normalized.ogg" -print0 | xargs -0 -P 8 -I"{}" rename "s/-normalized.ogg/.ogg/" "{}"
echo "finished"
