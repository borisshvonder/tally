#!/bin/sh

VERSION=0.5.0
TIMESTAMP=`date -u +%Y%m%d.%H%M`
FULLVERSION="$VERSION-$TIMESTAMP"

build() 
{
	local os="$1"
	local arch="$2"
	local ext="$3"
	mkdir -p "release/$os/$arch"
	(cd tallycli && GOOS="$os" GOARCH="$arch" go build \
		-ldflags "-X main.version=$FULLVERSION" \
		-o "../release/$os/$arch/tally${ext}")
}

rm -rf release/*
build linux amd64
build linux 386
build freebsd amd64
build freebsd 386
build windows amd64 .exe
build windows 386 .exe
