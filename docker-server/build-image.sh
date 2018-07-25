#!/bin/sh

# remove previous
rm tendermint_0.22.6_linux_amd64.zip
rm tendermint
rm tnmd
rm -rf init

# download tendermint
wget https://github.com/tendermint/tendermint/releases/download/v0.22.6/tendermint_0.22.6_linux_amd64.zip .
# unzip the file
unzip tendermint_0.22.6_linux_amd64.zip -d .

# remove the zip
rm tendermint_0.22.6_linux_amd64.zip

CGO_ENABLED=0 GOOS=linux go build -o tnmd -a -ldflags '-extldflags "-static"' ../cmd/tnmd/ 

tendermint init --home=init

docker build -t tendermoney .