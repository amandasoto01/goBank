package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
)

func seedAccount(store Storage, fname, lname, pw string) *Account {
	acc, err := NewAccount(fname, lname, pw)

	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("new account => ", acc.Number)
	return acc
}

func seedAccounts(s Storage) {
	seedAccount(s, "anthony", "GG", "password")
}

func main() {
	fmt.Println("hello to gobank api")

	seed := flag.Bool("seed", false, "seed of db")
	flag.Parse()

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

	if *seed {
		fmt.Println("seeding the database")

		// SEED
		seedAccounts(store)
	} else {
		fmt.Println("No seed flag provided")
	}

	server := NewAPIServer(":3000", store)
	server.Run()

}
