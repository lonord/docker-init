#!/bin/bash

cd "$(dirname "$0")"

if [ -z $PLATFORM ]; then
	export PLATFORM=$(uname)
fi
if [ -z $ARCH ]; then
	export ARCH=$(arch)
fi

make
