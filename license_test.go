package main

import (
	"os"
	"testing"
	"time"
)

func TestLicenseValidation(t *testing.T) {
	// Set up a test license key
	testLicenseKey := os.Getenv("TEST_LICENSE_KEY")
	if testLicenseKey == "" {
		t.Skip("TEST_LICENSE_KEY not set, skipping license tests")
	}

	licenseManager, err := NewLicenseManager()
	if err != nil {
		t.Fatalf("Failed to create license manager: %v", err)
	}

	// Test validation
	license, err := licenseManager.ValidateLicense(testLicenseKey)
	if err != nil {
		t.Fatalf("License validation failed: %v", err)
	}

	// Test license properties
	if license.CompanyName == "" {
		t.Error("License company name is empty")
	}
	if license.ExpiryDate.Before(time.Now()) {
		t.Error("License is already expired")
	}
	if license.MaxUsers <= 0 {
		t.Error("License max users is invalid")
	}
}

func TestLicenseRequirements(t *testing.T) {
	// Test without license (should fail)
	os.Unsetenv("AFCB_LICENSE_KEY")

	licenseManager, err := NewLicenseManager()
	if err != nil {
		t.Fatalf("Failed to create license manager: %v", err)
	}

	err = licenseManager.CheckLicenseRequirements()
	if err == nil {
		t.Error("Expected license check to fail without license key")
	}
}
