package main

import "time"

type License struct {
	CompanyName string    `json:"company_name"`
	Email       string    `json:"email"`
	MaxUsers    int       `json:"max_users"`
	ExpiryDate  time.Time `json:"expiry_date"`
	Domain      string    `json:"domain"`
	Version     string    `json:"version"`
	IssueDate   time.Time `json:"issue_date"`
	LicenseType string    `json:"license_type"`
}
