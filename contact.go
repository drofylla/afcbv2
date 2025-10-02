package main

import (
	"fmt"
	"strings"

	gonanoid "github.com/matoous/go-nanoid"
)

// struct for contact details
type Contact struct {
	ID          string
	ContactType string
	FirstName   string
	LastName    string
	Email       string
	Phone       string
	Password    string
}

// generate unique 6-character ID using custom alphabet & numbers
func genID() (string, error) {
	id, err := gonanoid.Generate("drofylla12301993", 6)
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %w", err)
	}
	return id, nil
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
