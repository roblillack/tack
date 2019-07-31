#!/bin/sh

DIRNAME=`dirname $0`
PREFIX=`cd $DIRNAME && pwd`

$PREFIX/../lib/tack/tack $@