#!/bin/sh

/app/tendermint node --home=/app/init --consensus.create_empty_blocks=false &
/app/tnmd -inflators-file=/app/inflators.json 