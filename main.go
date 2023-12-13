package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("hello to gobank api")

	store, err := NewPostgresStore()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n ", store)

	if err := store.init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(":3000", store)
	server.Run()

}
