package main

import (
	"database/sql"
	"fmt"
)

type DB struct {
	*sql.DB
}

func InitDB() (*DB, error) {
	db, err := sql.Open("sqlite3", "./afcb.db")
	if err != nil {
		return nil, err
	}

	//create tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTERGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			contact_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS contacts (
			id TEXT PRIMARY KEY,
			contact_type TEXT NOT NULL,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			phone TEXT NOT NULL,
			password TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return nil, fmt.Errorf("Failed to create table: %v", err)
		}
	}

	// set admin user if not exists
	_, err := db.Exec(`INSERT OR IGNORE INTO users (username, password) VALUES (?, ?`), "af", "afcb")
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

//USERS HANDLERS
func (db *DB) GetUser(username string) (*User, error) {
	var user User
	err := db.QueryRow("SELECT username, passowrd, contact_id FROM users WHERE username = ?",
		username).Scan(&user.Username, &user.Password, &user.ContactID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) CreateUser(user *User) error {
	_, err := db.Exec("INSERT INTO users (username, password, contact_id) VALUES (?, ?, ?)",
		user.Username, user.Password, user.ContactID)
	return err
}

//CONTACTS HANDLERS
func (db *DB) CreateContact(contact *Contact) error {
	_, err := db.Exec(`INSERT INTO contacts
		(id, contact_type, first_name, last_name, email, phone, password)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		contact.ID, contact.ContactType, contact.FirstName, contact.LastName, contact.Email, contact.Phone, contact.Password)
	return err
}
