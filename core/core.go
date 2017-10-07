package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// Blockchain ...
type Blockchain struct {
	UUID                string              `json:"uuid"`
	Chain               []Block             `json:"chain"`
	CurrentTransactions []Transaction       `json:"current_transactions"`
	Nodes               map[string]struct{} `json:"nodes"`
}

// Block ...
type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	Proof        int64         `json:"proof"`
	PreviousHash string        `json:"previous_hash"`
}

// Transaction ...
type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int64  `json:"amount"`
}

type RemoteChainResponse struct {
	Length int     `json:"length"`
	Chain  []Block `json:"chain"`
}

// NewBlockchain creates the blockchain and initializes a genesis block
func NewBlockchain() *Blockchain {

	blockchain := Blockchain{
		UUID:                uuid.New().String(),
		Chain:               make([]Block, 0),
		CurrentTransactions: make([]Transaction, 0),
		Nodes:               make(map[string]struct{}),
	}

	blockchain.NewBlock(100, "1")

	return &blockchain
}

// NewBlock creates a new block in the blockchain and return it, given
// the proof from the proof of work algorithm and the hash of the previous block.
func (blockchain *Blockchain) NewBlock(proof int64, previousHash string) Block {
	prevHash := previousHash
	if previousHash == "" {
		prevHash = hash(blockchain.Chain[len(blockchain.Chain)-1])
	}

	block := Block{
		Index:        int64(len(blockchain.Chain) + 1),
		Timestamp:    time.Now().Unix(),
		Transactions: blockchain.CurrentTransactions,
		Proof:        proof,
		PreviousHash: prevHash,
	}
	blockchain.CurrentTransactions = nil
	blockchain.Chain = append(blockchain.Chain, block)

	return block
}

// NewTransaction to go into the next mined Block
// given the sender (string), recipient (string) and amount (int64)
// and returns the index of the block that will hold this transaction (int64)
func (blockchain *Blockchain) NewTransaction(transaction Transaction) int64 {
	blockchain.CurrentTransactions = append(blockchain.CurrentTransactions, transaction)

	return blockchain.LastBlock().Index + 1
}

// RegisterNode adds a new node to the list of nodes, using a set of strings to avoid duplicates
func (blockchain *Blockchain) RegisterNode(address string) {
	parsedURL, err := url.Parse(address)
	if err != nil {
		log.Println(err)
	}
	blockchain.Nodes[parsedURL.Host] = struct{}{}
}

// ProofOfWork Algorithm, given lastProof (int64) return proof (int64):
// - Find a number p' such that hash(pp') contains leading 4 zeroes, where
// - p is the previous proof, and p' is the new proof
func (blockchain *Blockchain) ProofOfWork(lastProof int64) int64 {
	var proof int64 = 0
	for !blockchain.ValidProof(lastProof, proof) {
		proof++
	}

	return int64(proof)
}

// Validates the Proof: Does hash(last_proof, proof) contain 4 leading zeroes?
func (blockchain *Blockchain) ValidProof(lastProof int64, proof int64) bool {
	guess := fmt.Sprintf("%d%d", lastProof, proof)
	guessHash := fmt.Sprintf("%x", sha256.Sum256([]byte(guess)))
	log.Println(guessHash)
	return guessHash[:4] == "0000"
}

// ValidChain determines if a given blockchain is valid
func (blockchain *Blockchain) ValidChain(chain *[]Block) bool {
	lastBlock := (*chain)[0]
	currentIndex := 1

	for currentIndex < len(*chain) {
		block := (*chain)[currentIndex]
		log.Printf("LastBlock: %+v\n", lastBlock)
		log.Printf("Block: %+v\n", block)
		log.Println("-----------")

		// Check that the hash of the block is correct
		if block.PreviousHash != hash(lastBlock) {
			return false
		}

		// Check that the Proof of Work is correct
		if !blockchain.ValidProof(lastBlock.Proof, block.Proof) {
			return false
		}

		lastBlock = block
		currentIndex++
	}
	return true
}

// ResolveConflicts is our Consensus Algorithm, it resolves conflicts
// by replacing our chain with the longest one in the network.
func (blockchain *Blockchain) ResolveConflicts() bool {
	neighbours := blockchain.Nodes
	newChain := make([]Block, 0)

	// We're only looking for chains longer than ours
	maxLength := len(blockchain.Chain)
	// Grab and verify the chains from all the nodes in our network
	for node := range neighbours {
		response, err := http.Get(fmt.Sprintf("http://%s/chain", node))
		if err != nil {
			continue
		}
		if response.StatusCode == http.StatusOK {
			var remoteChain RemoteChainResponse
			if err := json.NewDecoder(response.Body).Decode(&remoteChain); err != nil {
				continue
			}

			// Check if the length is longer and the chain is valid
			if remoteChain.Length > maxLength && blockchain.ValidChain(&remoteChain.Chain) {
				maxLength = remoteChain.Length
				newChain = remoteChain.Chain
			}
		}
	}

	// Replace our chain if we discovered a new, valid chain longer than ours
	if len(newChain) > 0 {
		blockchain.Chain = newChain
		return true
	}

	return false
}

// LastBlock returns the last block in the chain
func (blockchain *Blockchain) LastBlock() Block {
	return blockchain.Chain[len(blockchain.Chain)-1]
}

func hash(block Block) string {
	var buffer bytes.Buffer
	jsonblock, err := json.Marshal(block)
	if err != nil {
		log.Printf("Error: %s", err)
	}
	err = binary.Write(&buffer, binary.BigEndian, jsonblock)
	if err != nil {
		log.Printf("Could not compute hash: %s", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(buffer.Bytes()))
}
