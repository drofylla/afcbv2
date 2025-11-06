package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
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
		// Header - Centered
		pdf.SetFont("Arial", "B", 16)
		pageWidth, _ := pdf.GetPageSize()
		titleWidth := pdf.GetStringWidth("AFcb Contact Card")
		titleX := (pageWidth - titleWidth) / 2
		pdf.SetX(titleX)
		pdf.Cell(titleWidth, 10, "AFcb Contact Card")
		pdf.Ln(15)
	})

	pdf.SetFooterFunc(func() {
		// Footer
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		footerText := fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 at 15:04"))
		pageWidth, _ := pdf.GetPageSize()
		textWidth := pdf.GetStringWidth(footerText)
		footerX := (pageWidth - textWidth) / 2
		pdf.SetX(footerX)
		pdf.Cell(textWidth, 10, footerText)
	})

	pdf.AddPage()

	startY := 40.0
	pdf.SetY(startY)

	// QR Code Section - Right Column
	qrCodeBytes, err := p.generateVCardQRCode(contact, companyName)
	if err != nil {
		log.Printf("Warning: Could not generate QR code: %v", err)
	} else {
		// Calculate positions for right column
		pageWidth, _ := pdf.GetPageSize()
		qrWidth := 45.0
		qrX := pageWidth - qrWidth - 25 // Right margin
		qrY := startY

		// "Scan to Save Contact" title above QR code
		pdf.SetXY(qrX, qrY)
		pdf.SetFont("Arial", "B", 14)
		scanTitle := "Scan to Save Contact"
		scanTitleWidth := pdf.GetStringWidth(scanTitle)
		scanTitleX := qrX + (qrWidth-scanTitleWidth)/2
		pdf.SetX(scanTitleX)
		pdf.Cell(scanTitleWidth, 8, scanTitle)
		pdf.Ln(8)

		// QR Code image
		qrOpts := gofpdf.ImageOptions{
			ImageType: "PNG",
			ReadDpi:   true,
		}

		qrReader := bytes.NewReader(qrCodeBytes)
		info := pdf.RegisterImageOptionsReader("qrcode", qrOpts, qrReader)
		if info == nil {
			log.Printf("Warning: Failed to register QR code image")
		} else {
			pdf.ImageOptions("qrcode", qrX, pdf.GetY(), qrWidth, qrWidth, false, qrOpts, 0, "")

			// QR code instruction below the image - Centered
			pdf.SetXY(qrX, pdf.GetY()+qrWidth+3)
			pdf.SetFont("Arial", "I", 8)
			instruction1 := "Scan with phone camera"
			instruction1Width := pdf.GetStringWidth(instruction1)
			instruction1X := qrX + (qrWidth-instruction1Width)/2
			pdf.SetX(instruction1X)
			pdf.Cell(instruction1Width, 4, instruction1)
			pdf.Ln(4)

			instruction2 := "to save contact"
			instruction2Width := pdf.GetStringWidth(instruction2)
			instruction2X := qrX + (qrWidth-instruction2Width)/2
			pdf.SetX(instruction2X)
			pdf.Cell(instruction2Width, 4, instruction2)
		}
	}

	// Contact Information Section - Left Column (Left Aligned)
	pdf.SetY(startY) // Reset to same starting position as QR code
	pdf.SetX(20)     // Left margin

	// Contact Information Title - Left Aligned
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Contact Information")
	pdf.Ln(12)

	// Name - Left Aligned
	pdf.SetX(20)
	pdf.SetFont("Arial", "B", 14)
	name := fmt.Sprintf("%s %s", contact.FirstName, contact.LastName)
	pdf.Cell(0, 8, name)
	pdf.Ln(8)

	// ID and Type - Left Aligned with indentation
	pdf.SetX(25)
	pdf.SetFont("Arial", "", 11)
	idText := fmt.Sprintf("ID: %s", contact.ID)
	idWidth := pdf.GetStringWidth(idText) + 5
	pdf.Cell(idWidth, 7, idText)

	pdf.SetFont("Arial", "I", 11)
	contactTypeText := fmt.Sprintf("Type: %s", contact.ContactType)
	pdf.Cell(0, 7, contactTypeText)
	pdf.Ln(10)

	// Company - Left Aligned with indentation
	if companyName != "" {
		pdf.SetX(25)
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(0, 8, fmt.Sprintf("Company: %s", companyName))
		pdf.Ln(12)
	}

	// Contact Methods Section - Left Aligned
	pdf.SetX(20)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Contact Methods")
	pdf.Ln(12)

	// Email - Left Aligned with indentation
	pdf.SetX(25)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(18, 7, "Email:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, contact.Email)
	pdf.Ln(8)

	// Phone - Left Aligned with indentation
	pdf.SetX(25) // Indented 5mm
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(18, 7, "Phone:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, contact.Phone)
	pdf.Ln(15)

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}

	return buf.Bytes(), nil
}

func (p *PDFService) generateVCardQRCode(contact *Contact, companyName string) ([]byte, error) {
	vcard := p.generateVCardContent(contact, companyName)

	// Generate QR code
	qrBytes, err := qrcode.Encode(vcard, qrcode.Medium, 200)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %v", err)
	}

	return qrBytes, nil
}

func (p *PDFService) generateVCardContent(contact *Contact, companyName string) string {
	var vcard strings.Builder

	vcard.WriteString("BEGIN:VCARD\n")
	vcard.WriteString("VERSION:3.0\n")
	vcard.WriteString(fmt.Sprintf("FN:%s %s\n", contact.FirstName, contact.LastName))
	vcard.WriteString(fmt.Sprintf("N:%s;%s;;;\n", contact.LastName, contact.FirstName))

	if contact.Email != "" {
		vcard.WriteString(fmt.Sprintf("EMAIL:%s\n", contact.Email))
	}

	if contact.Phone != "" {
		cleanPhone := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' || r == '+' {
				return r
			}
			return -1
		}, contact.Phone)
		vcard.WriteString(fmt.Sprintf("TEL:%s\n", cleanPhone))
	}

	if companyName != "" {
		vcard.WriteString(fmt.Sprintf("ORG:%s\n", companyName))
	}

	// Timestamp
	vcard.WriteString(fmt.Sprintf("REV:%s\n", time.Now().Format("20060102T150405Z")))

	vcard.WriteString("END:VCARD")

	return vcard.String()
}
