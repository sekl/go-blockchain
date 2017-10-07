package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sekl/go-blockchain/core"
)

// Response for the API
type Response struct {
	Message string `json:"message"`
}

// ChainResponse is used to format our /chain API response
type ChainResponse struct {
	Chain  []core.Block `json:"chain"`
	Length int64        `json:"length"`
}

// MineResponse for the API
type MineResponse struct {
	Message      string             `json:"message"`
	Index        int64              `json:"index"`
	Transactions []core.Transaction `json:"transactions"`
	Proof        int64              `json:"proof"`
	PreviousHash string             `json:"previous_hash"`
}

func Mine(w http.ResponseWriter, r *http.Request, blockchain *core.Blockchain, nodeID string) {
	// We run the proof of work algorithm to get the next proof...
	lastBlock := blockchain.LastBlock()
	lastProof := lastBlock.Proof
	proof := blockchain.ProofOfWork(lastProof)

	// We must receive a reward for finding the proof.
	// The sender is "0" to signify that this node has mined a new coin.
	tx := core.Transaction{
		Sender:    "0",
		Recipient: nodeID,
		Amount:    1,
	}
	blockchain.NewTransaction(tx)

	// Forge the new Block by adding it to the chain
	block := blockchain.NewBlock(proof, "")

	response := MineResponse{
		Message:      "New Block Forged",
		Index:        block.Index,
		Transactions: block.Transactions,
		Proof:        block.Proof,
		PreviousHash: block.PreviousHash,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
	log.Printf("%+v\n", response)
}

func NewTransaction(w http.ResponseWriter, r *http.Request, blockchain *core.Blockchain) {
	decoder := json.NewDecoder(r.Body)
	var tx core.Transaction
	err := decoder.Decode(&tx)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	log.Printf("%+v\n", tx)
	// Check that the required fields are in the POST'ed data
	if tx.Sender == "" || tx.Recipient == "" || tx.Amount == 0 {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	// Create a new Transaction
	index := blockchain.NewTransaction(tx)

	response := Response{
		Message: fmt.Sprintf("Transaction will be added to Block %d", index),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(response)
	log.Printf("%+v\n", response)
}

func GetChain(w http.ResponseWriter, r *http.Request, blockchain *core.Blockchain) {
	response := ChainResponse{
		Chain:  blockchain.Chain,
		Length: int64(len(blockchain.Chain)),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
	log.Printf("%+v\n", response)
}

func RegisterNode(w http.ResponseWriter, r *http.Request, blockchain *core.Blockchain) {
	var request map[string][]string
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	for _, node := range request["nodes"] {
		blockchain.RegisterNode(node)
	}

	response := Response{
		Message: "New nodes have been added",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
	log.Printf("%+v\n", response)
}

func ResolveConflict(w http.ResponseWriter, r *http.Request, blockchain *core.Blockchain) {
	replaced := blockchain.ResolveConflicts()

	response := Response{}
	if replaced {
		response.Message = "Our chain was replaced"
	} else {
		response.Message = "Our chain is authorative"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
	log.Printf("%+v\n", response)
}
