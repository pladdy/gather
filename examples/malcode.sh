#!/usr/bin/env bash
SOURCE_DIR=$(pwd $(dirname $0))
${SOURCE_DIR}/grun download -u http://malc0de.com/bl/IP_Blacklist.txt -s malcode.txt
