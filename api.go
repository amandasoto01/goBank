package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

// gorilla mux good support for http and to handle all the requests
// stable package from the beginning
func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountByID)))
	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransferAccount))

	log.Println("JSON API server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

// good practice prefix with handle
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	// switch to switch statement or add .methods
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s ", r.Method)
}

// GET /account
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getID(r)
		if err != nil {
			return err
		}

		account, err := s.store.GetAccountByID(id)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, account)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("Method not allowed")
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	accountRequest := new(CreateAccountRequest)
	//accountRequest := CreateAccountRequest{}

	// decode needs the reference when its decoding plain json bytes
	//if err := json.NewDecoder(r.Body).Decode(&accountRequest); err != nil {
	if err := json.NewDecoder(r.Body).Decode(accountRequest); err != nil {
		return err
	}

	account := NewAccount(accountRequest.FirstName, accountRequest.LastName)

	if err := s.store.CreateAccount(account); err != nil {
		return err

	}

	tokenString, err := createJWT(account)

	if err != nil {
		return err
	}

	fmt.Println("token: ", tokenString)
	return writeJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)

	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

func (s *APIServer) handleTransferAccount(w http.ResponseWriter, r *http.Request) error {
	transferRequest := new(TransferRequest)

	if err := json.NewDecoder(r.Body).Decode(transferRequest); err != nil {
		return err
	}
	defer r.Body.Close()

	return writeJSON(w, http.StatusOK, transferRequest)
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func createJWT(account *Account) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	})

	secret := os.Getenv("JWT_SECRET")
	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString([]byte(secret))
}

func withJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")
		_, err := validateJWT(tokenString)

		if err != nil {
			writeJSON(w, http.StatusForbidden, ApiError{Error: "Invalid token"})
			return
		}

		handlerFunc(w, r)

	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		secret := os.Getenv("JWT_SECRET")
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, ApiError{
				Error: err.Error(),
			})
		}
	}
}

func getID(r *http.Request) (int, error) {
	params := mux.Vars(r)
	idStr := params["id"]
	id, err := strconv.Atoi(idStr)

	if err != nil {
		return 0, fmt.Errorf("invalid id given %s ", idStr)
	}

	return id, nil
}
