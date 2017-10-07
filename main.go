package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/sekl/go-blockchain/api"
	"github.com/sekl/go-blockchain/core"
)

func main() {
	blockchain := core.NewBlockchain()

	nodeID := uuid.New().String()

	r := chi.NewRouter()
	r.Get("/mine", func(w http.ResponseWriter, r *http.Request) {
		api.Mine(w, r, blockchain, nodeID)
	})
	r.Post("/transactions/new", func(w http.ResponseWriter, r *http.Request) {
		api.NewTransaction(w, r, blockchain)
	})
	r.Get("/chain", func(w http.ResponseWriter, r *http.Request) {
		api.GetChain(w, r, blockchain)
	})
	r.Post("/nodes/register", func(w http.ResponseWriter, r *http.Request) {
		api.RegisterNode(w, r, blockchain)
	})
	r.Get("/nodes/resolve", func(w http.ResponseWriter, r *http.Request) {
		api.ResolveConflict(w, r, blockchain)
	})

	// Get the port from the command line
	var port int
	flag.IntVar(&port, "p", 5000, "Port to use (default: 5000)")
	flag.Parse()

	log.Printf("Node running on localhost:%d...", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
