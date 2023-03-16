package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

func main() {
	db, err := sql.Open("sqlite3", "./books.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/books", createBook).Methods("POST")
	router.HandleFunc("/books/{id}", updateBook).Methods("PUT")
	router.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	// Retrieve all books from the database and return them as JSON
}

func getBook(w http.ResponseWriter, r *http.Request) {
	// Retrieve a single book by ID from the database and return it as JSON
}

func createBook(w http.ResponseWriter, r *http.Request) {
	// Create a new book in the database using the JSON data from the request
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	// Update an existing book in the database using the JSON data from the request and the book ID from the path
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	// Delete a book from the database using the book ID from the path
}
