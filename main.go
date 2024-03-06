package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const serverAddr string = "127.0.0.1:8081"

type Scope struct {
	Project string
	Area    string
}

type Note struct {
	Title string
	Tags  []string
	Text  string
	Scope Scope
}

var mdbClient *mongo.Client

func main() {
	var err error
	ctxBg := context.Background()
	const connStr string = "mongodb+srv://jamontes:Ladelmongo.13@micluster.sh7o7d1.mongodb.net/?retryWrites=true&w=majority&appName=micluster"
	mdbClient, err = mongo.Connect(ctxBg, options.Client().ApplyURI(connStr))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = mdbClient.Disconnect(ctxBg); err != nil {
			panic(err)
		}
	}()
	router := http.NewServeMux()
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HTTP Caracola"))
	})
	router.HandleFunc("POST /notes", createNote)

	server := http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	server.RegisterOnShutdown(func() {
		fmt.Println("Signal shutdown")
	})

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error %v\n", err)
	}
}

func createNote(w http.ResponseWriter, r *http.Request) {
	var note Note
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&note); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	notesCollection := mdbClient.Database("NoteKeeper").Collection("Notes")
	result, err := notesCollection.InsertOne(r.Context(), note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Id: %v", result.InsertedID)
	fmt.Fprintf(w, "Note: %+v", note)
}
