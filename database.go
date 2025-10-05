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

	// ensure company_id column exists in contacts table
	_, err = db.Exec(`ALTER TABLE contacts ADD COLUMN company_id TEXT`)
	if err != nil {
		// Ignore "duplicate column" errors
		if !strings.Contains(err.Error(), "duplicate column") {
			fmt.Printf("Note: Could not alter contacts table: %v\n", err)
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

// COMPANY HANDLERS
func (db *DB) CreateCompany(company *Company) error {
	_, err := db.Exec(`INSERT INTO companies
		(id, name, bank_name, account_number, account_document_path, registration_number, registration_document_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		company.ID, company.Name, company.BankName, company.AccountNumber, company.AccountDocumentPath, company.RegistrationNumber, company.RegistrationDocumentPath)
	return err
}

func (db *DB) GetCompany(id string) (*Company, error) {
	var company Company
	err := db.QueryRow(`SELECT id, name, bank_name, account_number, account_document_path,
		registration_number, registration_document_path, created_at FROM companies WHERE id = ?`, id).Scan(&company.ID, &company.Name, &company.BankName, &company.AccountNumber, &company.AccountDocumentPath, &company.RegistrationNumber, &company.RegistrationDocumentPath, &company.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (db *DB) UpdateCompany(company *Company) error {
	_, err := db.Exec(`UPDATE companies SET
		name = ?, bank_name = ?, account_number = ?, account_document_path = ?, registration_number = ?, registration_document_path = ? WHERE id = ?`,
		company.Name, company.BankName, company.AccountNumber, company.AccountDocumentPath, company.RegistrationNumber, company.RegistrationDocumentPath, company.ID)
	return err
}

func (db *DB) DeleteCompany(id string) error {
	_, err := db.Exec("DELETE FROM companies WHERE id = ?", id)
	return err
}

func (db *DB) GetAllCompanies() ([]Company, error) {
	rows, err := db.Query("SELECT id, name, bank_name, account_number, registration_number, created_at FROM companies ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var company Company
		err := rows.Scan(&company.ID, &company.Name, &company.BankName, &company.AccountNumber, &company.RegistrationNumber, &company.CreatedAt)
		if err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}
	return companies, nil
}

func (db *DB) GetCompanies() ([]Company, error) {
	rows, err := db.Query("SELECT id, name FROM companies ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var company Company
		err := rows.Scan(&company.ID, &company.Name)
		if err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}
	return companies, nil
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
		(id, contact_type, first_name, last_name, email, phone, password, company_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		contact.ID, contact.ContactType, contact.FirstName, contact.LastName, contact.Email, contact.Phone, contact.Password, contact.CompanyID)
	return err
}

func (db *DB) GetContact(id string) (*Contact, error) {
	var contact Contact
	var companyID sql.NullString
	err := db.QueryRow(`SELECT id, contact_type, first_name, last_name, email, phone, password, company_id FROM contacts WHERE id = ?`, id).Scan(
		&contact.ID, &contact.ContactType, &contact.FirstName, &contact.LastName, &contact.Email, &contact.Phone, &contact.Password, &companyID)
	if err != nil {
		return nil, err
	}
	if companyID.Valid {
		contact.CompanyID = &companyID.String
	}
	return &contact, nil
}

func (db *DB) UpdateContact(contact *Contact) error {
	_, err := db.Exec(`UPDATE contacts SET contact_type = ?, first_name = ?, last_name = ?, email = ?, phone = ?, password = ?, company_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		contact.ContactType, contact.FirstName, contact.LastName, contact.Email, contact.Phone, contact.Password, contact.CompanyID, contact.ID)
	return err
}

func (db *DB) DeleteContact(id string) error {
	_, err := db.Exec("DELETE FROM contacts WHERE id = ?", id)
	return err
}

func (db *DB) GetAllContacts() ([]Contact, error) {
	rows, err := db.Query("SELECT id, contact_type, first_name, last_name, email, phone, company_id FROM contacts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		var companyID sql.NullString
		err := rows.Scan(&contact.ID, &contact.ContactType, &contact.FirstName, &contact.LastName, &contact.Email, &contact.Phone, &companyID)
		if err != nil {
			return nil, err
		}
		if companyID.Valid {
			contact.CompanyID = &companyID.String
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

func (db *DB) SearchContacts(keyword string) ([]Contact, error) {
	query := `SELECT id, contact_type, first_name, last_name, email, phone, company_id FROM contacts WHERE first_name LIKE ? OR last_name LIKE ? OR email LIKE ? or phone LIKE ?`

	likeKeyword := "%" + keyword + "%"
	rows, err := db.Query(query, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		var companyID sql.NullString
		err := rows.Scan(&contact.ID, &contact.ContactType, &contact.FirstName, &contact.LastName, &contact.Email, &contact.Phone, &companyID)
		if err != nil {
			return nil, err
		}
		if companyID.Valid {
			contact.CompanyID = &companyID.String
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

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
