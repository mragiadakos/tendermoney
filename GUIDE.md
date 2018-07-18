- First we need to start ipfs daemon on its own console
$ ipfs daemon

- Secondly we need to start tendermint node on its own console, with a fresh blockchain
  The version of tendermint used is 0.21.0
$ rm -rf ~/.tendermint/
$ tendermint init
$ tendermint node

- Now we need to create an inflator's public and private key, using the client
$ ./client gi --filename=inflator.json
The generate was successful
$ cat inflator.json 
{"PublicKey":"98cd0a61bcd7336f9f60ded6b04ec8714b3b860c52e56125fa6c71eb63ec7b32","PrivateKey":"acde5bab96bd2098d41e84b30a5c4e91154aac1699b9a0d9ee138cfa2bbbfe01"}

- We will add the inflator's key in another file. 
  The file will contain the list of inflators in a json format and put it near the validator.
$ echo '["98cd0a61bcd7336f9f60ded6b04ec8714b3b860c52e56125fa6c71eb63ec7b32"]' > ../app/inflators.json

- Before using the inflators.json from the validator, we need to create an IPFS hash
$ ipfs add inflators.json 
added QmRq7ms5zrpoyVNUfANZG6MFnq2N1s7Soe4kM3eAyro6Q4 inflators.json

- We will start the validator and use from there the inflators.json through the IPFS
$ ./server -inflators=QmRq7ms5zrpoyVNUfANZG6MFnq2N1s7Soe4kM3eAyro6Q4
I[06-29|16:55:51.684] Starting ABCIServer                          module=abci-server impl=ABCIServer
I[06-29|16:55:51.685] Waiting for new connection...                module=abci-server

- We will create the first coin with the value of 1 and save it in a local folder called 'vault'
  However, the value of a coin needs to be specific based on this list:
   500, 100, 50, 20, 10, 5, 2, 1, 0.50, 0.20, 0.10, 0.05, 0.02, 0.01

$ ./client i --key inflator.json --vault vault --value 1
The coin created successfully and saved in vault/f448dea6-d09d-4113-ae26-60ae31f0e9b7

$ cat vault/f448dea6-d09d-4113-ae26-60ae31f0e9b7
{"OwnerPrivateKey":"12079eea474c60652629975f853051c588cdfd5da62c05ea8f69773a1949ef02","OwnerPublicKey":"249c19438271f202c5be70462d9ae981c17b6a5a631f5244ff6f15f569143433","UUID":"f448dea6-d09d-4113-ae26-60ae31f0e9b7","Value":1}

- Lets sum two coins into one and create a coin with value of 2. 
  First will create another coin
$ ./client i --key inflator.json --vault vault --value 1
The coin created successfully and saved in vault/0228fb59-9ebe-48ea-86f6-231ecceab4be
  
  We will sum these two coins that we just created, into one
$ ./client s --vault vault --coins="0228fb59-9ebe-48ea-86f6-231ecceab4be,f448dea6-d09d-4113-ae26-60ae31f0e9b7"
The coin created successfully and saved in vault/67c281d7-752e-4b7d-a3af-00cc7b294d42

  Now the new coin has the value of 2 
$ cat vault/67c281d7-752e-4b7d-a3af-00cc7b294d42
{"OwnerPrivateKey":"77d03b94b26b50027bca22abc498cb434f63fc76a4c48e619a7a8b5d205b330e","OwnerPublicKey":"fa2a8a80414e1d87666db964dc70dada3ac54639cb229b56bb491bcb11659d76","UUID":"67c281d7-752e-4b7d-a3af-00cc7b294d42","Value":2}

  The previous coins have been deleted from the system and folder.

- Lets devide this coin into two coins of 1.
$ ./client d --vault vault --coin=67c281d7-752e-4b7d-a3af-00cc7b294d42 --values=1,1
2  new coins have been created:
vault/7790ffa7-6e64-417a-a542-0ceedb11e50e
vault/72b93cf8-ac6a-4d5f-9742-10da3113516c

  Also we can see that the previous file of the coin has been replaced by two other files
$ ls vault
72b93cf8-ac6a-4d5f-9742-10da3113516c  7790ffa7-6e64-417a-a542-0ceedb11e50e

- To continue with transactions, we will create a tax. But it is not necessary.
  This tax will be 10%. For example, if we send two coins of 1 value then the fee will be 0.02.
$ ./client tax --key inflator.json --percent 10
The tax has been submitted.

- We can get the latest tax
$ ./client get_latest_tax
Percent:  10
Inflator:  98cd0a61bcd7336f9f60ded6b04ec8714b3b860c52e56125fa6c71eb63ec7b32

- Now we will create two coins for the fee
$ ./client i --key inflator.json --vault vault --value 0.10
The coin created successfully and saved in vault/3ab712fd-85eb-4137-92af-13eee571a2eb
$ ./client i --key inflator.json --vault vault --value 0.10
The coin created successfully and saved in vault/9bfa8bdf-33bf-4339-ae2a-b28f5878dd55

- We will start the transaction by sending the coins.
  It will return the hash of the transaction and the secret that we need to give to the receiver.
  We expect that the coins will be removed from the vault because they will be locked anyway.
$ ./client send --vault vault --coins=72b93cf8-ac6a-4d5f-9742-10da3113516c,7790ffa7-6e64-417a-a542-0ceedb11e50e --fee=9bfa8bdf-33bf-4339-ae2a-b28f5878dd55,3ab712fd-85eb-4137-92af-13eee571a2eb
Hash:  2712d9b5cc7a0f326e6263a7aed15ffaaa4cc904cfdcf533d670fcba3e9e3dd4
Secret:  eyJHSGV4IjoiZWNhZWQwNGNmOTFjMDk3NjQ4ZTZlNDZiNDU5YjhmYzgwNjgxZWI1MTUxM2FmMjg5NzM3Zjc0YzMxMjFlZGM5OCIsIkhIZXgiOiI1ZWYwNTRiYjg4OGVhOTlhNjkxNThiZDkxOWFiNjk1MjM0NTBmMzU1ODMwMWQ5NTkzOWY5MGU0MGMxYzI5NDZjIiwiWEdIZXgiOiJiODExNTEwN2ZiZTQwOGI0ZWY3NWY3ODQ0MjA2ZTU5MzQwYmI2MzMwOGRjNjBkZjYxZGZkN2Y2MTZjNjdjYWRhIiwiWEhIZXgiOiI5ZDU5M2I2YjUxZTdmN2M5MGRlNjdjYjE1OWUxYTdiNzM1NDZlMGNmZjRlYzc2Y2NhMWRmMWVhYThiMWIzNTAwIn0=

- We can check the transaction if it exists in the system
$ ./client get_transaction --hash 2712d9b5cc7a0f326e6263a7aed15ffaaa4cc904cfdcf533d670fcba3e9e3dd4
Hash:  2712d9b5cc7a0f326e6263a7aed15ffaaa4cc904cfdcf533d670fcba3e9e3dd4
Coins:  [72b93cf8-ac6a-4d5f-9742-10da3113516c 7790ffa7-6e64-417a-a542-0ceedb11e50e]
Fee:  [9bfa8bdf-33bf-4339-ae2a-b28f5878dd55 3ab712fd-85eb-4137-92af-13eee571a2eb]
The coins have been received:  false
The fee have been received:  false

- Also we can check the state of a coin, if it is locked or not.
$ ./client get_coin --coin 72b93cf8-ac6a-4d5f-9742-10da3113516c
Coin: 72b93cf8-ac6a-4d5f-9742-10da3113516c
Value: 1
Owner: ed9e73dc6f80cea3e1d84145b3b2c4289c9b2f61d2e83ef1698c5b5e69f99d14
Locked: true

- Lets receive the coins of the transactions and put them in the vault
$ ./client  receive  --vault vault --hash 2712d9b5cc7a0f326e6263a7aed15ffaaa4cc904cfdcf533d670fcba3e9e3dd4 --secret eyJHSGV4IjoiZWNhZWQwNGNmOTFjMDk3NjQ4ZTZlNDZiNDU5YjhmYzgwNjgxZWI1MTUxM2FmMjg5NzM3Zjc0YzMxMjFlZGM5OCIsIkhIZXgiOiI1ZWYwNTRiYjg4OGVhOTlhNjkxNThiZDkxOWFiNjk1MjM0NTBmMzU1ODMwMWQ5NTkzOWY5MGU0MGMxYzI5NDZjIiwiWEdIZXgiOiJiODExNTEwN2ZiZTQwOGI0ZWY3NWY3ODQ0MjA2ZTU5MzQwYmI2MzMwOGRjNjBkZjYxZGZkN2Y2MTZjNjdjYWRhIiwiWEhIZXgiOiI5ZDU5M2I2YjUxZTdmN2M5MGRlNjdjYjE1OWUxYTdiNzM1NDZlMGNmZjRlYzc2Y2NhMWRmMWVhYThiMWIzNTAwIn0
2  new coins have been created:
vault/72b93cf8-ac6a-4d5f-9742-10da3113516c
vault/7790ffa7-6e64-417a-a542-0ceedb11e50e

- If we check the state of the coin again, we will see it is not locked
$ ./client get_coin --coin 72b93cf8-ac6a-4d5f-9742-10da3113516c
Coin: 72b93cf8-ac6a-4d5f-9742-10da3113516c
Value: 1
Owner: 04be911344a18f0e662136ffadf291325cd12049cd051a7fe503c7fbb63d2661
Locked: false

- Also the transaction now says that the coins have been received.
$ ./client get_transaction --hash=2712d9b5cc7a0f326e6263a7aed15ffaaa4cc904cfdcf533d670fcba3e9e3dd4
Hash:  2712d9b5cc7a0f326e6263a7aed15ffaaa4cc904cfdcf533d670fcba3e9e3dd4
Coins:  [72b93cf8-ac6a-4d5f-9742-10da3113516c 7790ffa7-6e64-417a-a542-0ceedb11e50e]
Fee:  [9bfa8bdf-33bf-4339-ae2a-b28f5878dd55 3ab712fd-85eb-4137-92af-13eee571a2eb]
The coins have been received:  true
The fee have been received:  false

- However the fee have not been received. As an inflator, I will receive the taxes also
  But also as an inflator I need to see all the transactions that their fee have not been received to automate it.
$ ./client get_transactions_with_unreceived_fee
Hash:  2712d9b5cc7a0f326e6263a7aed15ffaaa4cc904cfdcf533d670fcba3e9e3dd4
Coins:  [72b93cf8-ac6a-4d5f-9742-10da3113516c 7790ffa7-6e64-417a-a542-0ceedb11e50e]
Fee:  [9bfa8bdf-33bf-4339-ae2a-b28f5878dd55 3ab712fd-85eb-4137-92af-13eee571a2eb]
The coins have been received:  true
The fee have been received:  false

- Here we receiving the fee from the transaction in another vault for the taxes
$ ./client receive_fee --key inflator.json --vault vaultTax --hash 2712d9b5cc7a0f326e6263a7aed15ffaaa4cc904cfdcf533d670fcba3e9e3dd4
2  new coins have been created:
vaultTax/9bfa8bdf-33bf-4339-ae2a-b28f5878dd55
vaultTax/3ab712fd-85eb-4137-92af-13eee571a2eb

- We see that the transaction is fully received. 
  That means the coins for the fee is also unlocked.
$ ./client get_transaction --hash=2712d9b5cc7a0f326e6263a7aed15ffaaa4cc904cfdcf533d670fcba3e9e3dd4
Hash:  2712d9b5cc7a0f326e6263a7aed15ffaaa4cc904cfdcf533d670fcba3e9e3dd4
Coins:  [72b93cf8-ac6a-4d5f-9742-10da3113516c 7790ffa7-6e64-417a-a542-0ceedb11e50e]
Fee:  [9bfa8bdf-33bf-4339-ae2a-b28f5878dd55 3ab712fd-85eb-4137-92af-13eee571a2eb]
The coins have been received:  true
The fee have been received:  true
