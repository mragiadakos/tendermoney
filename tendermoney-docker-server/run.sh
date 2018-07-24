#!/bin/sh

/app/tendermint node --home=/app/init &
/app/tnmd -inflators-file=/app/inflators.json 