package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	GetAccountByNumber(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=gobak dbname=postgres password=mysecretpassword port=2023 sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil

}

func (s *PostgresStore) init() error {
	return s.CreateAccountTable()
}

func (s *PostgresStore) CreateAccountTable() error {
	query := `create table if not exists account( 
    id serial primary key, 
    first_name varchar(50), 
    last_name varchar(50), 
    encrypted_password varchar(250),
    number serial, 
    balance serial, 
    created_at timestamp)
	`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(account *Account) error {
	query := `
		insert into account (first_name, last_name, encrypted_password, number, balance, created_at)
		values ($1, $2, $3, $4, $5, $6)`

	_, err := s.db.Query(
		query,
		account.FirstName, account.LastName, account.EncryptedPassword, account.Number, account.Balance, account.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateAccount(account *Account) error {
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	_, err := s.db.Query("delete from account where id = $1", id)

	return err
}

func (s *PostgresStore) GetAccountByNumber(number int) (*Account, error) {
	rows, err := s.db.Query("select * from account where number = $1", number)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account with number [%d] not found ", number)
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id = $1", id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %d not found ", id)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, error := s.db.Query("select * from account")

	if error != nil {
		return nil, error
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.EncryptedPassword,
		&account.Number,
		&account.Balance,
		&account.CreatedAt)
	return account, err
}
