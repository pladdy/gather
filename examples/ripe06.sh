#!/usr/bin/env bash
SOURCE_DIR=$(pwd $(dirname $0))
${SOURCE_DIR}/grun scrape -u http://data.ris.ripe.net/rrc06/2016.12/ -p bview.20161219.*?.gz -w last -s ripe-6.gz
