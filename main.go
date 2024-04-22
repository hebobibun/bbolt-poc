package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	// Open the BoltDB database
	var err error
	db, err = bolt.Open("items.db", 0600, nil)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()
	log.Println("Database opened successfully")

	// Create buckets if not exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("items"))
		return err
	})
	if err != nil {
		log.Fatal("Error creating bucket:", err)
	}
	log.Println("Bucket 'items' created successfully")

	// Initialize router
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/items", getAllItems).Methods("GET")
	router.HandleFunc("/items/{id}", getItem).Methods("GET")
	router.HandleFunc("/items", createItem).Methods("POST")
	router.HandleFunc("/items/{id}", updateItem).Methods("PUT")
	router.HandleFunc("/items/{id}", deleteItem).Methods("DELETE")

	// Start server
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getAllItems(w http.ResponseWriter, r *http.Request) {
	var items []Item

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var item Item
			if err := json.Unmarshal(v, &item); err != nil {
				return err
			}
			items = append(items, item)
			return nil
		})
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error retrieving items:", err)
		return
	}

	json.NewEncoder(w).Encode(items)
	log.Println("Get all items successfuly.")
}

func getItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var item Item
	var v []byte

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		if b == nil {
			return nil
		}

		v = b.Get([]byte(id))
		if v == nil {
			http.NotFound(w, r)
			log.Println("Item not found for ID:", id)
			return nil
		}

		return json.Unmarshal(v, &item)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error retrieving item:", err)
		return
	}

	json.NewEncoder(w).Encode(item)
	log.Printf("Get item with id %v: %v\n", id, string(v))
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Error decoding JSON:", err)
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		if b == nil {
			return nil
		}

		encoded, err := json.Marshal(item)
		if err != nil {
			return err
		}

		return b.Put([]byte(item.ID), encoded)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error creating item:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Println("Item with ID", item.ID, "created successfully")
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Error decoding JSON:", err)
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		if b == nil {
			return nil
		}

		encoded, err := json.Marshal(item)
		if err != nil {
			return err
		}

		return b.Put([]byte(id), encoded)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error updating item:", err)
		return
	}

	log.Println("Item with ID", id, "updated successfully")
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		if b == nil {
			return nil
		}

		return b.Delete([]byte(id))
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error deleting item:", err)
		return
	}

	log.Println("Item with ID", id, "deleted successfully")
}
