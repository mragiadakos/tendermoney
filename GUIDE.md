- First we need to start ipfs daemon on its own console
$ ipfs daemon

- Secondly we need to start tendermint node on its own console, with a fresh blockchain
$ rm -rf ~/.tendermint/
$ tendermint init
$ tendermint node

- Now we need to create an inflator's public and private key, using the client


