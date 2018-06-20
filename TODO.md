Tendermoney

The tendermoney will be a blockchain to exchange money, where each money is represented by only one public key.
On the transaction the money will change ownership by changing the public key.
The money can have only these values 500, 100, 50, 20, 10, 5, 2, 1, 50c, 20c,10c, 5c, 2c and 1c, called constant values.
There will be 5 types of transactions: inflate, sum, devide, tax, send, receive.
- inflate: this action will inflate the number of money that will be in circulation.
  Only the inflators can use this action.
- sum: this action will sum the constant value of coins, to get a new coin based on the constant values.
  For example, two coins of value 10 will get 20. However, adding 10 and 5, will not create a coin with the value 15.
  The uuids of the previous coins will be destroyed and no one can use them.
- divide: this action will devide the constant value of a coin, to get new coins based on the constant values.
  For example, one coin of value 10 will be two coins of 5.
  The uuids of the previous coin will be destroyed and no one can use it.
- tax: this action will tell what is the fee of the transactions in percentage.
  Only the inflators can add it. 
  Only the latest tax will be used for the transactions after it.
- send: this action will put the public keys of the coins into one and sign the list coins' UUIDs, a sha256 hash of the list
  and a shared signature (that signes the hash). 
  The validator will wait for the receiver to give the other shared signature to validate the transaction.
  The send will also contain a fee, based on the public keys the validator gave to the user.
- receiver: this action will put the hash, the shared secret and the list of public keys based on the list of UUIDs' order and a signature 
  from the public keys.

The user can query public keys and get the UUID and the value that represents and vice versa.
Also he can query the hash of the sender.



Actions API:

- Inflate
Request:
{
    Type: INFLATE
    Signature: hex
    Data: {
        Coin: uuid
        Value: Float
        Owner: public key in hex
        Inflator: public key
    }
}
Response:
  The request will fail on these scenarios:
  - Value is not in the list of contant values (d)
  - The coin is empty (d)
  - The owner is empty (d)
  - The inflator is empty (d)
  - The owner is not in the list of the inflators (d)
  - The coin's public key with the inflator's public key does not validate the signature (d)
  - The coin exists already (d)
  - The public key exists already (d)

- Tax
Request:
{
    Type: TAX
    Signature: hex
    Data: {
        Percentage: int
        Inflator: public_key_hex
    }
}
Response:
  The request will fail on these scenarios:
  - The percentage is a negative number
  - The percentage is over 100
  - The signature does not validate the inflator

- Sum
Request:
{
    Type: SUM
    Signature: hex
    Data: {
        Coins: []uuid 
        NewCoin: uuid
        NewOwner: punlic key in hex
    }
}
Response:
  The request will fail on these scenarios:
  - The list of coins is empty.
  - A coin does not have an owner.
  - The list of public keys based on the coins, does not validate the signature
  - NewCoin is empty
  - NewCoin exists already
  - The NewOwner is empty
  - The NewOwner exists already

- Divide
Request:
{
    Type: DIVIDE
    Signature: hex
    Data: {
        Coin: uuid 
        NewCoins: map[uuid]{ Value: int, Owner: public_key_hex }
    }
}
Response:
 The request will fail on these scenarios:
 - The coin is empty
 - The owner of the coin does not validate the signature
 - The value of a coin is not based on the constant values
 - The sum of values is not equal the value of the coin
 - A coin from the new coins, exists already
 - An owner from the new owners, exists already

- Send
Request:
{
    Type: SEND
    Signature: hex
    Data: {
        Coins : []uuid
        Hash: sha256 hash of the list of uuids
        SharedSignature: hex
        PubPoly: public key in hex
        Fee : map[uuid]public_key_hex  
    }
}
Response:
 The request will fail on these scenarios:
 - The list of coins is empty
 - A coin from the coins and fee exists
 - The hash does not validate the coins
 - The SharedSignature is empty
 - The list of public keys, based on the coins and fee, do not validate the signature
 - The SharedSignature is empty
 - The PubPoly is empty
 - The public keys of the fee haven't submitted by the validator

- Receive
Request:
{
    Type: RECEIVE
    Signature: hex
    Data: {
        Hash: sha256 hash of the list of uuids
        NewOwners: map[uuid]public_key_hex  
        SharedSignature: hex
    }
}
Response:
  The request will fail on these scenarios:
  - The hash is empty
  - The hash does not exist from a sender
  - The hash has already been collected
  - The NewOwners uuids does not exist
  - The public keys are already owners
  - The shared signature does not validate the hash with the PubPoly of the sender



Query API

- Get the coin of a public key
Request:
{
    Coin: uuid
}
Response:
  Successful:
{
    Owner: public key hex
}
The request fail if the uuid does not exists

- Get the public key of a coin
{
    Owner: public key hex
}
Response:
  Successful:
{
    Coin: uuid
}

- Get latest tax
Request:
Path: /tax
Response:
{
    Percentage: int
}

- Get the public keys from the validator for the fee
Request:
{
    Signature: hex
    Data: {
        Coins: []uuid
        Date: time
    }
}
Responce:
Successful scenarios
{
  Owners: map[uuid]public_key_hex
}
Failed scenarios
 - The coins do not validate the signature
 - The date passed one minute

