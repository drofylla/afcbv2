package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

// Generate PDF for contact card
func (p *PDFService) GenerateContactCardPDF(contact *Contact, companyName string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// PDF setup
	pdf.SetHeaderFunc(func() {
		//Header
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(40, 10, "AFcb Contact Card")
		pdf.Ln(12)
	})

	pdf.SetFooterFunc(func() {
		//Footer
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.Cell(0, 10, fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 at 15:04")))
	})

	pdf.AddPage()

	//Title
	pdf.SetFont("Arial", "B", 13)
	pdf.Cell(0, 15, "Contact Information")
	pdf.Ln(9)

	//Contact details
	//Name & IDF
	// pdf.SetXY(45, 40)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, fmt.Sprintf("%s %s", contact.FirstName, contact.LastName))
	pdf.Ln(5)
	//ID
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(13, 7, fmt.Sprintf("(%s)", contact.ID))
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(0, 7, fmt.Sprintf("-%s", contact.ContactType))
	//Company
	if companyName != "" {
		pdf.SetX(45)
		pdf.Cell(0, 7, fmt.Sprintf("Company: %s", companyName))
		pdf.Ln(3)
	}
	pdf.Ln(6)
	//Contact Methods
	pdf.SetFont("Arial", "B", 13)
	pdf.Cell(0, 10, "Contact Methods")
	pdf.Ln(6)
	//Email
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 7, "Email:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 7, contact.Email)
	pdf.Ln(5)
	//Phone
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(25, 7, "Phone:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 7, contact.Phone)
	pdf.Ln(5)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}
