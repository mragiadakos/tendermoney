package models

type RetrieveData struct {
	TransactionHash string            // sha256 hex
	NewOwners       map[string]string // map[uuid]public_key_hex
	Inflator        string
}
