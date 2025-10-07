package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"
)

type LicenseData struct {
	CompanyName string    `json:"company_name"`
	Email       string    `json:"email"`
	MaxUsers    int       `json:"max_users"`
	ExpiryDate  time.Time `json:"expiry_date"`
	Domain      string    `json:"domain"`
	Version     string    `json:"version"`
	IssueDate   time.Time `json:"issue_date"`
	LicenseType string    `json:"license_type"` //"trial", "permanent"
}

type LicenseGenerator struct {
	privateKey *rsa.PrivateKey
}

func NewLicenseGenerator() (*LicenseGenerator, error) {
	//generate or load private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return &LicenseGenerator{privateKey: privateKey}, nil
}

func (lg *LicenseGenerator) LoadPrivateKeyFromFile(filename string) error {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return fmt.Errorf("Failed to parse PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	lg.privateKey = privateKey
	return nil
}

func (lg *LicenseGenerator) SavePrivateKey(filename string) error {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(lg.privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return os.WriteFile(filename, privateKeyPEM, 0600)
}

func (lg *LicenseGenerator) GetPublicKeyPEM() string {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&lg.privateKey.PublicKey)

	if err != nil {
		return ""
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM)
}

func (lg *LicenseGenerator) GenerateLicense(licenseData *LicenseData) (string, error) {
	//Serialize license data to JSON
	data, err := json.Marshal(licenseData)
	if err != nil {
		return "", err
	}

	//Create signature
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, lg.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	//Create license structure
	license := struct {
		Data      []byte `json:"data"`
		Signature []byte `json:"signature"`
	}{
		Data:      data,
		Signature: signature,
	}

	//Encode to JSON
	licenseJSON, err := json.Marshal(license)
	if err != nil {
		return "", err
	}

	//Base64 encode for easy distribution
	licenseKey := base64.StdEncoding.EncodeToString(licenseJSON)
	return licenseKey, nil
}

func (lg *LicenseGenerator) GenerateTrialLicense(companyName, email string, days int) (string, error) {
	licenseData := &LicenseData{
		CompanyName: companyName,
		Email:       email,
		MaxUsers:    3,
		ExpiryDate:  time.Now().AddDate(0, 0, days),
		Domain:      "*",
		Version:     "1.0",
		IssueDate:   time.Now(),
		LicenseType: "trial",
	}
	return lg.GenerateLicense(licenseData)
}

func (lg *LicenseGenerator) GeneratePermanentLicense(companyName, email, domain string, maxUsers int, months int) (string, error) {
	licenseData := &LicenseData{
		CompanyName: companyName,
		Email:       email,
		MaxUsers:    maxUsers,
		ExpiryDate:  time.Now().AddDate(0, months, 0),
		Domain:      domain,
		Version:     "1.0",
		IssueDate:   time.Now(),
		LicenseType: "permanent",
	}
	return lg.GenerateLicense(licenseData)
}

func GenerateLicenseCLI() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println(" go run license_gen.go trial <company> <email> <days>")
		fmt.Println(" go run license_gen.go keygen (generate new key pair")
		return
	}

	generator, err := NewLicenseGenerator()
	if err != nil {
		log.Fatal("Failed to create license generator", err)
	}

	//Load existing private key
	if err := generator.LoadPrivateKeyFromFile("private.key"); err != nil {
		fmt.Println("No existing private key found, generate new key pair")
		//save new private key
		if err := generator.SavePrivateKey("private.key"); err != nil {
			log.Fatal("Failed to save private key:", err)
		}
		fmt.Println("New key generated and saved to private.key")
	}

	command := os.Args[1]

	switch command {
	case "keygen":
		//generate & save keys
		fmt.Println("Public Key (add to your license.go):")
		fmt.Println(generator.GetPublicKeyPEM())
		fmt.Println("\nPrivate key saved to private.key")
	case "trial":
		if len(os.Args) != 5 {
			fmt.Println("Usage: go run license_gen.go trial <company> <email> <days>")
			return
		}
		company := os.Args[2]
		email := os.Args[3]
		days := os.Args[4]

		var daysInt int
		fmt.Sscanf(days, "%d", &daysInt)

		licenseKey, err := generator.GenerateTrialLicense(company, email, daysInt)

		if err != nil {
			log.Fatal("Failed to generate trial license:", err)
		}

		fmt.Printf("Trial License Generated:\n")
		fmt.Printf("Company: %s\n", company)
		fmt.Printf("Email: %s\n", email)
		fmt.Printf("Duration: %d days\n", daysInt)
		fmt.Printf("License Key:\n%s\n", licenseKey)
	case "permanent":
		if len(os.Args) != 7 {
			fmt.Println("Usage: go run license_gen.go permanent <company> <email> <domain> <max_users> <months>")
			return
		}
		company := os.Args[2]
		email := os.Args[3]
		domain := os.Args[4]
		maxUsers := os.Args[5]
		months := os.Args[6]

		var maxUsersInt, monthsInt int
		fmt.Sscanf(maxUsers, "%d", &maxUsersInt)
		fmt.Sscanf(months, "%d", &monthsInt)

		licenseKey, err := generator.GeneratePermanentLicense(company, email, domain, maxUsersInt, monthsInt)
		if err != nil {
			log.Fatal("Failed to generate permanent license:", err)
		}
		fmt.Printf("Permanent License Generated:\n")
		fmt.Printf("Company: %s\n", company)
		fmt.Printf("Email: %s\n", email)
		fmt.Printf("Domain: %s\n", domain)
		fmt.Printf("Max Users: %d\n", maxUsersInt)
		fmt.Printf("Duration: %d months\n", monthsInt)
		fmt.Printf("License Key:\n%s\n", licenseKey)
	default:
		fmt.Println("Unknown command:", command)
	}
}
