package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	fmt.Println("hello to gobank api")

	if err := godotenv.Load(); err != nil {
		fmt.Println("Error cargando el archivo .env:", err)
		return
	}

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
