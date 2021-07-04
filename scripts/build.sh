#!/usr/bin/env bash

set -e

REPO=$(cd $(dirname $0)/.. && pwd)
PWD=$(pwd)
COMMAND=tack

cd "$REPO"
VERSION=$(git describe 2> /dev/null)
[ -n "$1" ] && VERSION="$1"
platforms=("windows/amd64" "darwin/amd64" "darwin/arm64" "linux/amd64" "openbsd/amd64" "freebsd/amd64" "netbsd/amd64")

echo "Building $VERSION ..."

for platform in "${platforms[@]}"
do
    echo "* $platform"
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$COMMAND'_'$GOOS'_'$GOARCH'_'$VERSION
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name \
        -ldflags "-X github.com/roblillack/tack/commands.Version=$VERSION" .
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done

cd "$PWD"