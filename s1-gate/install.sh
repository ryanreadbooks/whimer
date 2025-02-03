#!/bin/bash

set -e

ROCKS_DEP=("penlight")
for rock in "${ROCKS_DEP[@]}"; do
  echo "Attempting to install $rock..."
  if luarocks list --porcelain $rock | grep -q "installed"; then
    echo $rock already installed
  else
    echo installing $rock via luarocks...
    luarocks install $rock
  fi
done

echo "All luarocks packages have been installed"
echo ''

OPM_DEP=("SkyLothar/lua-resty-jwt" "ledgetech/lua-resty-http")
for dep in "${OPM_DEP[@]}"; do
  echo "Attempting to install $dep..."
  opm get $dep
done
