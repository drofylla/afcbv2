package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func InitDB() (*DB, error) {
	db, err := sql.Open("sqlite3", "./afcb.db")
	if err != nil {
		return nil, err
	}

	//Test connect
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %v", err)
	}

	//create tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS companies (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			bank_name TEXT,
			account_number TEXT,
			account_document_path TEXT,
			registration_number TEXT,
			registration_document_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			contact_id TEXT,
			needs_password_change BOOLEAN DEFAULT 1,
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

	// ensure the needs_password_change column exists
	_, err = db.Exec(`ALTER TABLE users ADD COLUMN needs_password_change BOOLEAN DEFAULT 1`)
	if err != nil {
		// Ignore "duplicate column" errors
		if !strings.Contains(err.Error(), "duplicate column") {
			fmt.Printf("Note: Could not alter users table: %v\n", err)
		}
	}

	// Insert default admin user if not exists - mark as NOT needing password change
	result, err := db.Exec(`INSERT OR IGNORE INTO users (username, password, needs_password_change) VALUES (?, ?, ?)`,
		"af", "afcb", 0) // Admin doesn't need password change
	if err != nil {
		return nil, fmt.Errorf("failed to create default admin user: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Default admin user created: af / afcb")
	} else {
		log.Printf("Default admin user already exists")
	}

	return &DB{db}, nil
}

// USERS HANDLERS
func (db *DB) GetUser(username string) (*User, error) {
	var user User
	var needsChange interface{} //to handle different types

	err := db.QueryRow("SELECT username, password, contact_id, needs_password_change FROM users WHERE username = ?",
		username).Scan(&user.Username, &user.Password, &user.ContactID, &needsChange)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	//handle possible types of boolean
	switch v := needsChange.(type) {
	case int64:
		user.NeedPasswordChange = (v == 1)
	case bool:
		user.NeedPasswordChange = v
	case string:
		user.NeedPasswordChange = (v == "1" || v == "true")
	case nil:
		user.NeedPasswordChange = true // Default to true if NULL
	default:
		user.NeedPasswordChange = true // Default to true for unknown types
	}
	return &user, nil
}

func (db *DB) CreateUser(user *User) error {
	needsChange := 0
	if user.NeedPasswordChange {
		needsChange = 1
	}
	_, err := db.Exec("INSERT INTO users (username, password, contact_id, needs_password_change) VALUES (?, ?, ?, ?)",
		user.Username, user.Password, user.ContactID, needsChange)
	return err
}

func (db *DB) UpdateUserPassword(username, newPassword string) error {
	_, err := db.Exec("UPDATE users SET password = ?, needs_password_change = 0 WHERE username = ?", newPassword, username)
	return err
}

func (db *DB) UserNeedsPasswordChange(username string) (bool, error) {
	var needsChange int
	err := db.QueryRow("SELECT needs_password_change FROM users WHERE username = ?", username).Scan(&needsChange)
	if err != nil {
		return false, err
	}
	return needsChange == 1, nil
}

func (db *DB) DeleteUser(username string) error {
	_, err := db.Exec("DELETE FROM users WHERE username = ?", username)
	return err
}

// CONTACTS HANDLERS
func (db *DB) CreateContact(contact *Contact) error {
	_, err := db.Exec(`INSERT INTO contacts
		(id, contact_type, first_name, last_name, email, phone, password)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		contact.ID, contact.ContactType, contact.FirstName, contact.LastName, contact.Email, contact.Phone, contact.Password)
	return err
}

func (db *DB) GetContact(id string) (*Contact, error) {
	var contact Contact
	err := db.QueryRow(`SELECT id, contact_type, first_name, last_name, email, phone, password FROM contacts WHERE id = ?`, id).Scan(
		&contact.ID, &contact.ContactType, &contact.FirstName, &contact.LastName, &contact.Email, &contact.Phone, &contact.Password)
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (db *DB) UpdateContact(contact *Contact) error {
	_, err := db.Exec(`UPDATE contacts SET contact_type = ?, first_name = ?, last_name = ?, email = ?, phone = ?, password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		contact.ContactType, contact.FirstName, contact.LastName, contact.Email, contact.Phone, contact.Password, contact.ID)
	return err
}

func (db *DB) DeleteContact(id string) error {
	_, err := db.Exec("DELETE FROM contacts WHERE id = ?", id)
	return err
}

func (db *DB) GetAllContacts() ([]Contact, error) {
	rows, err := db.Query("SELECT id, contact_type, first_name, last_name, email, phone FROM contacts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		err := rows.Scan(&contact.ID, &contact.ContactType, &contact.FirstName, &contact.LastName, &contact.Email, &contact.Phone)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

func (db *DB) SearchContacts(keyword string) ([]Contact, error) {
	query := `SELECT id, contact_type, first_name, last_name, email, phone FROM contacts WHERE first_name LIKE ? OR last_name LIKE ? OR email LIKE ? or phone LIKE ?`

	likeKeyword := "%" + keyword + "%"
	rows, err := db.Query(query, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		err := rows.Scan(&contact.ID, &contact.ContactType, &contact.FirstName, &contact.LastName, &contact.Email, &contact.Phone)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

// Add this function to debug the database schema
func (db *DB) DebugUserTable() error {
	fmt.Println("=== Debugging users table ===")

	// Check table structure
	rows, err := db.Query("PRAGMA table_info(users)")
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Users table columns:")
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt_value interface{}
		rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk)
		fmt.Printf("  Column: %s, Type: %s, PK: %d\n", name, ctype, pk)
	}

	// Check actual data
	dataRows, err := db.Query("SELECT username, password, contact_id, needs_password_change, typeof(needs_password_change) FROM users")
	if err != nil {
		return err
	}
	defer dataRows.Close()

	fmt.Println("Users data:")
	for dataRows.Next() {
		var username, password string
		var contactID sql.NullString
		var needsChange interface{}
		var needsChangeType string

		dataRows.Scan(&username, &password, &contactID, &needsChange, &needsChangeType)
		fmt.Printf("  User: %s, Password: %s, NeedsChange: %v (%s)\n", username, password, needsChange, needsChangeType)
	}

	fmt.Println("=== End debugging ===")
	return nil
}
