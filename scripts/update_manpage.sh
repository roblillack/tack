#!/usr/bin/env bash

set -e

REPO=$(cd $(dirname $0)/.. && pwd)

pandoc "$REPO"/tack.1.md -s -t man -o "$REPO"/tack.1
