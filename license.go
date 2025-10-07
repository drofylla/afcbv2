package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

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

type LicenseManager struct {
	publicKey *rsa.PublicKey
}

func NewLicenseManager() (*LicenseManager, error) {
	//Public key should match private key used to generate licenses
	publicKeyPEM := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzPH31szhqbLkNxChlinG
zHhepT7fmDNbnY4ziB1O3mMKECvvKZ+WkdeaBS0YTgOeoDYcSZ/Y41wLUOdWmw4G
JTL2PmCU/PDuew350Kr2SU4JhA817Q3TOioFJBU6ImMgEeB4R77JB5xmo3r5byEV
B1cP7KOnWg88duouZdvGc2+VXIRQvioj61Z0ufmZ4pVdVQCXiK5D1TStju3rcYa0
ZdnD1IdNinwtSJMmS6dMm7YVi5R6dF2jRbxCHNWgNCiDo/GhFATKN1RJ97VGmTyV
pjPbiFu9dEvcDuB5ud3G025CJJ/QwZuw32qxgo/Okk48FBLWWTBHsnIIMUDVmR0j
BQIDAQAB
-----END PUBLIC KEY-----`

	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("Failed to parse PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Not an RSA public key")
	}

	return &LicenseManager{publicKey: rsaPub}, nil
}

func (lm *LicenseManager) ValidateLicense(licenseKey string) (*License, error) {
	//decode base64 license
	licenseJSON, err := base64.StdEncoding.DecodeString(licenseKey)
	if err != nil {
		return nil, fmt.Errorf("Invalid license format: %v", err)
	}

	//parse license structure
	var licenseStruct struct {
		Data      []byte `json:"data"`
		Signature []byte `json:"signature"`
	}

	if err := json.Unmarshal(licenseJSON, &licenseStruct); err != nil {
		return nil, fmt.Errorf("Invalid license structure: %v", err)
	}

	//verify signature
	hashed := sha256.Sum256(licenseStruct.Data)
	if err := rsa.VerifyPKCS1v15(lm.publicKey, crypto.SHA256, hashed[:], licenseStruct.Signature); err != nil {
		return nil, fmt.Errorf("Invalid license signature: %v", err)
	}

	//parse license data
	var license License
	if err := json.Unmarshal(licenseStruct.Data, &license); err != nil {
		return nil, fmt.Errorf("Invalid license data: %v", err)
	}

	return &license, nil
}

func (lm *LicenseManager) CheckLicenseRequirements() error {
	licenseKey := os.Getenv("AFCB_LICENSE_KEY")
	if licenseKey == "" {
		return fmt.Errorf("License key not found. Please set AFCB_LICENSE_KEY environment variable")
	}

	license, err := lm.ValidateLicense(licenseKey)
	if err != nil {
		return fmt.Errorf("Invalid license: %v", err)
	}

	//check expiry date
	if time.Now().After(license.ExpiryDate) {
		return fmt.Errorf("License expired on %s", license.ExpiryDate.Format("2006-01-02"))
	}

	//check user limits
	currentUsers, err := lm.getCurrentUserCount()
	if err == nil && currentUsers > license.MaxUsers {
		return fmt.Errorf("User limit exceeded: %d/%d users", currentUsers, license.MaxUsers)
	}
	return nil
}

func (lm *LicenseManager) getCurrentUserCount() (int, error) {
	return 0, nil
}

func (lm *LicenseManager) GetLicenseInfo(licenseKey string) (*License, error) {
	return lm.ValidateLicense(licenseKey)
}
