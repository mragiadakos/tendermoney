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
  If the tax is lower than the constant value, then the validator will choose the lowest value from the constant values
- send: this action is to send money to a person.
  The sender will add the uuids of the coins in a list and he will create a hash based on the list.
  Then he will use the dleq algorithm to create a proof.
  Both the hash and the secret from the proof, will send to the receiver.
  Also he will create a list of coins for the fee.
  The validator will get the coins from the list, empty them from public keys and lock them for the inflator to give them new public keys.
  The sender's signature will be based on the public keys of all the coins.
- receive: this action will give the receiver the opportunity to create public keys for the coins.
  He will put the hash, the verifier of the proof and the list of public keys based on the coins. 
  From the new public keys, the receiver will create the signature
- retrieve_fee: this action can only be used by the inflator, to unlock the fees and put new owners to the coins.

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
  Success
  - The coin exists in the DB (d)

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
  - The list of coins is empty (d)
  - A coin added twice (d)
  - NewCoin is empty (d)
  - The NewOwner is empty (d)
  - The sum of coins is not equal to a constant value (d)
  - A coin does not exists. (d)
  - The NewCoin exists already (d)
  - The NewOwner exists already (d)
  - The list of public keys based on the coins, does not validate the signature (d)
  Success:
  - The new coin is the database with the owner and the old one coins and owners are not in (d)

- Divide
Request:
{
    Type: DIVIDE
    Signature: hex
    Data: {
        Coin: uuid 
        NewCoins: map[uuid]{ Value: float64, Owner: public_key_hex }
    }
}
Response:
 The request will fail on these scenarios:
 - The coin is empty (d)
 - The new coins is empty (d)
 - The owner of a new coin is empty (d)
 - A value from the new coins, is not based on the constant values (d)
 - The new coins have the same owner (d)
 - The sum of new coins' values, is not equal to the value of the coin (d)
 - A coin from the new coins, exists already (d)
 - An owner from the new owners, exists already (d) 
 - The owner of the coin with the new coins do not validate the signature (d)
 Success
 - The new coins with the owners exist in the DB and the old one does not (d)

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
  - The percentage is a negative number (d)
  - The percentage is over 100 (d)
  - The inflator is not in the list of inflators (d)
  - The signature does not validate the inflator (d)
  - A method based on the tax that will return the lowest constant value fee (d)
    Examples if tax is 23%
    - the transaction is 1 then the fee will 0.23 (d)
    - The transaction is 0.50 then the fee will be 0.11 (d)
    - The transaction is 0.10 then the fee will be 0.02 (d)
    - The transaction is 0.01 the the fee will be 0.01 (d)
  Success
  - The tax is saved on the DB. Save 3 taxes and expect that the last, is the one that can only be used. (d)


- Send
Request:
{
    Type: SEND
    Signature: hex
    Data: {
        Coins : []uuid
        Proof:  dleq.Proof
        Fee : []uuid 
    }
}
Response:
 The request will fail on these scenarios:
 - The list of coins is empty (d)
 - The list of coins from fee is empty when tax exists (d)
 - A coin from the coins added twice (d)
 - A coin from the fee added twice (d)
 - A coin from the coins and fee added on both (d)
 - A coin from the coins does not exist (d)
 - A coin from the fee does not exist (d)
 - The fee is not based on the tax  (d)
 - The list of public keys, based on the coins and fee, do not validate the signature (d)
 - The proof is not encoded correctly (d)
 Success:
 - The transaction exists in the db based on the hash of coins (d)
 - All the coins are locked and unusable for any action (d)
   Fail to sum (d)
   Fail to divide (d) 
   Fail to another send (d)


- Receive
Request:
{
    Type: RECEIVE
    Signature: hex
    Data: {
        TransactionHash: sha256 hash, in hex, of the list of coins
        NewOwners: map[uuid]public_key_hex  
        ProofVerification:{
            G: kyber.Point
            H: kyber.Point
            XG: kyber.Point
            XH: kyber.Point
        }
    }
}
Response:
  The request will fail on these scenarios:
  - The hash is empty (d)
  - The hash does not exist (d)
  - The coins are not in the transaction. (d)
  - The new owners are already owners (d)
  - The proof is not correct (d)
  - The proof is not valid (d)
  - The signature does not validate based on the new owners (d)
  - Can not receive the coins twice (d)
  Success
  - The coins have been unlocked (d)
  - The coins have new owners (d)
  - The older owners have been removed (d)
  - The transaction's has been received (d)

- Retrieve Fee
Request:
{
  Type: RETRIEVE_FEE
  Signature: hex
  Data:{
    TransactionHash: sha256 hash, in hex, of the list of coins
    NewOwners: map[uuid]public_key_hex  
    Inflator: public_key_hex
  }
}
Response:
  The request will fail on these scenarios:
  - The NewOwners is empty
  - The inflator is empty
  - The coins are not locked
  - The public keys of the owners exist already
  - The signature does from both the inflator and new owners, does not validate the transaction
  Success:
  - The coins have the new owners and they are unlocked.


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


- Get the list of uuid fees.
Request:
{
    Fees: map[uuid]Coin
}