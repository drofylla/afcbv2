package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

var db *DB

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

const dataFile = "AFcb.db" // Now using SQLite database

var conCard = template.Must(template.New("card").Funcs(template.FuncMap{
	"getCompanyName": func(companyID *string) string {
		if companyID == nil || *companyID == "" {
			return ""
		}
		company, err := db.GetCompany(*companyID)
		if err != nil {
			fmt.Printf("Error getting company %s: %v\n", *companyID, err)
			return ""
		}
		return company.Name
	},
}).Parse(`
	<div class="card bg-white rounded-xl shadow-md p-6 hover:shadow-lg transition-all duration-300" id="contact-{{.ID}}">
    <div class="details">
        <span class="id text-xs font-semibold text-gray-500">ID: {{.ID}}</span>
        <strong class="name block text-xl font-bold text-gray-800 mt-1">{{.FirstName}} {{.LastName}}</strong>
        {{if .CompanyID}}
        <div class="company mt-1">
            <span class="text-sm text-gray-600">{{getCompanyName .CompanyID}}</span>
        </div>
        {{end}}
        <span class="type inline-block mt-2 px-3 py-1 rounded-full text-sm font-medium
            {{if eq .ContactType "Personal"}}bg-blue-100 text-blue-800
            {{else if eq .ContactType "Work"}}bg-green-100 text-green-800
            {{else if eq .ContactType "Family"}}bg-purple-100 text-purple-800
            {{else}}bg-gray-100 text-gray-800{{end}}">
            {{.ContactType}}
        </span>
        <div class="details mt-3 text-gray-600">
            <div class="flex items-center mb-1">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
                <span id="email-{{.ID}}">{{.Email}}</span>
                <button onclick="copyEmail('email-{{.ID}}')" class="ml-2 p-1 rounded-full hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500" title="Copy Email">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2.5a1.5 1.5 0 011.5 1.5v4.5m-14-6.5h3v-3h-3v3z" />
                    </svg>
                </button>
            </div>
            <div class="flex items-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                </svg>
                <span>{{.Phone}}</span>
                <a href="https://wa.me/{{.Phone}}" target="_blank" class="ml-2 p-1 rounded-full text-green-500 hover:bg-green-100 transition-colors" title="WhatsApp">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12.04 2.87c-5.42 0-9.82 4.4-9.82 9.82 0 1.94.57 3.8.14 5.39l-1.39 5.09 5.25-1.36c1.5.25 3.09.4 4.56.4 5.42 0 9.82-4.4 9.82-9.82-.01-5.42-4.4-9.81-9.8-9.81zm-.04 17.1c-1.36 0-2.7-.22-3.9-.66l-2.61.68.68-2.55c-.5-1.16-.76-2.43-.76-3.75 0-4.41 3.59-8 8-8s8 3.59 8 8-3.59 8-8 8zm4.53-5.59c-.25-.13-.49-.2-.72-.2-.23 0-.46.07-.69.21-.23.14-.52.28-.84.38-.32.1-.64.16-.96.06-.32-.1-.6-.24-.87-.45-.27-.2-.5-.45-.7-.7-.19-.24-.34-.49-.49-.77s-.27-.58-.33-.89c-.06-.31-.05-.59-.01-.84.04-.26.13-.5.26-.72.13-.22.25-.4.36-.57.11-.17.18-.32.22-.44.04-.12.02-.27-.04-.43-.06-.16-.18-.32-.34-.48-.16-.16-.36-.31-.6-.44-.24-.13-.49-.2-.73-.2-.24 0-.48.05-.72.15-.24.1-.46.25-.66.44-.2.19-.38.41-.54.67-.16.26-.28.53-.4.81s-.2 0-.25-.06c-.05-.06-.2-.25-.37-.47s-.35-.4-.5-.54c-.16-.14-.28-.2-.37-.2s-.22 0-.36-.05c-.14-.05-.3-.08-.5-.09-.19-.01-.39-.01-.58 0-.19 0-.4.04-.61.09-.2.05-.4.14-.57.26-.17.12-.3.27-.4.45-.1.18-.15.39-.15.63s.06.48.19.74c.12.26.3.52.54.78.24.26.54.55.89.87.35.31.75.63 1.18.96 1.05.78 1.95 1.48 2.5 1.77.55.29 1.01.44 1.39.44.38 0 .82-.13 1.34-.38.52-.25.96-.54 1.33-.88.37-.34.6-.78.71-1.32.11-.54.06-1.04-.08-1.52z"/>
                    </svg>
                </a>
                <a href="signal://send?text=&phone={{.Phone}}" target="_blank" class="ml-2 p-1 rounded-full text-gray-800 hover:bg-gray-200 transition-colors" title="Signal">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm.8 14.8c-.37.37-.87.5-1.37.5-.5 0-1-.13-1.37-.5-.75-.75-.75-1.99 0-2.74L12 11.39l-1.44-1.44c-.75-.75-.75-1.99 0-2.74s1.99-.75 2.74 0L12 8.61l1.44-1.44c.75-.75 1.99-.75 2.74 0s.75 1.99 0 2.74L12.8 12.8l1.44 1.44c.75.75.75 1.99 0 2.74zm0 0"/>
                    </svg>
                </a>
            </div>
        </div>
    </div>
    <div class="actions flex justify-end mt-4 space-x-2">
        <button class="edit-btn p-2 rounded-lg border border-gray-300 hover:border-blue-500 hover:bg-blue-50 transition-colors"
            hx-get="/modal/edit/{{.ID}}"
            hx-target="#modal-container"
            hx-swap="innerHTML"
            title="Edit">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 20h9"/>
                <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4 12.5-12.5z"/>
            </svg>
        </button>
        <button class="delete-btn p-2 rounded-lg border border-gray-300 hover:border-red-500 hover:bg-red-50 transition-colors"
                hx-delete="/contacts/{{.ID}}"
                hx-target="#contact-{{.ID}}"
                hx-swap="outerHTML"
                hx-confirm="Are you sure you want to delete this contact?"
                title="Delete">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="3 6 5 6 21 6"/>
                <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6m5 0V4a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v2"/>
                <line x1="10" y1="11" x2="10" y2="17"/>
                <line x1="14" y1="11" x2="14" y2="17"/>
            </svg>
        </button>
    </div>
</div>
`))

var addCompanyModalHTML = `
<div id="company-modal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full">
    <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div class="flex justify-end">
            <button hx-target="#company-modal" hx-swap="outerHTML" hx-get="/modal/close" class="text-gray-400 hover:text-gray-600">&times;</button>
        </div>
        <h3 class="text-xl font-bold mb-4">Add New Company</h3>
        <form id="companyForm" enctype="multipart/form-data"
              hx-post="/companies"
              hx-target="#company-list"
              hx-swap="afterbegin"
              hx-on::after-request="if(event.detail.successful) htmx.remove(htmx.find('#company-modal'))">
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="companyName">Company Name</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                       id="companyName" name="name" type="text" placeholder="Company Name" required>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="bankName">Bank Name</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                       id="bankName" name="bank_name" type="text" placeholder="Bank Name">
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="accountNumber">Account Number</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                       id="accountNumber" name="account_number" type="text" placeholder="Account Number">
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="accountDocument">Account Document</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                       id="accountDocument" name="account_document" type="file" accept=".pdf,.jpg,.jpeg,.png">
                <p class="text-xs text-gray-500 mt-1">Upload bank statement or account proof (PDF, JPG, PNG)</p>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="registrationNumber">Registration Number</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                       id="registrationNumber" name="registration_number" type="text" placeholder="Registration Number">
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="registrationDocument">Registration Document</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                       id="registrationDocument" name="registration_document" type="file" accept=".pdf,.jpg,.jpeg,.png">
                <p class="text-xs text-gray-500 mt-1">Upload company registration document (PDF, JPG, PNG)</p>
            </div>
            <div class="flex items-center justify-end">
                <button type="button" hx-target="#company-modal" hx-swap="outerHTML" hx-get="/modal/close"
                        class="bg-gray-500 text-white font-bold py-2 px-4 rounded-lg shadow-md hover:bg-gray-600 transition-colors duration-300 mr-2">Cancel</button>
                <button type="submit" class="bg-blue-600 text-white font-bold py-2 px-4 rounded-lg shadow-md hover:bg-blue-700 transition-colors duration-300">Save Company</button>
            </div>
        </form>
    </div>
</div>
`

var addModalHTML = `
<div id="contact-modal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full">
    <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div class="flex justify-end">
            <button hx-target="#contact-modal" hx-swap="outerHTML" hx-get="/modal/close" class="text-gray-400 hover:text-gray-600">&times;</button>
        </div>
        <h3 class="text-xl font-bold mb-4">Add New Contact</h3>
        <form id="contactForm"
              hx-post="/contacts"
              hx-target="#contact-list"
              hx-swap="afterbegin"
              hx-on::after-request="if(event.detail.successful) htmx.remove(htmx.find('#contact-modal'))">
            <input type="hidden" id="contact-id" name="id">
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="contactType">Contact Type</label>
                <select id="contactType" name="ContactType" class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline">
                    <option value="Personal">Personal</option>
                    <option value="Work">Work</option>
                    <option value="Family">Family</option>
                </select>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="firstName">First Name</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="firstName" name="FirstName" type="text" placeholder="First Name" required>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="lastName">Last Name</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="lastName" name="LastName" type="text" placeholder="Last Name" required>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="email">Email</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="email" name="Email" type="email" placeholder="Email" required>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="phone">Phone</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="phone" name="Phone" type="tel" placeholder="Phone" required>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="company">Company</label>
                <select id="company" name="CompanyID" class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline">
                    <option value="">No Company</option>
                    {{range .Companies}}
                    <option value="{{.ID}}">{{.Name}}</option>
                    {{end}}
                </select>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="password">Password</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                       id="password" name="Password" type="password" placeholder="Password for login">
                <p class="text-xs text-gray-500 mt-1">Leave empty to auto-generate</p>
            </div>
            <div class="flex items-center justify-end">
                <button type="button" hx-target="#contact-modal" hx-swap="outerHTML" hx-get="/modal/close" class="bg-gray-500 text-white font-bold py-2 px-4 rounded-lg shadow-md hover:bg-gray-600 transition-colors duration-300 mr-2">Cancel</button>
                <button type="submit" class="bg-blue-600 text-white font-bold py-2 px-4 rounded-lg shadow-md hover:bg-blue-700 transition-colors duration-300">Save Contact</button>
            </div>
        </form>
    </div>
</div>
`

var editModalHTML = `
<div id="contact-modal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full">
    <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div class="flex justify-end">
            <button hx-target="#contact-modal" hx-swap="outerHTML" hx-get="/modal/close" class="text-gray-400 hover:text-gray-600">&times;</button>
        </div>
        <h3 class="text-xl font-bold mb-4">Edit Contact</h3>
        <form id="contactForm"
              hx-put="/contacts/{{.Contact.ID}}"
              hx-target="#contact-{{.Contact.ID}}"
              hx-swap="outerHTML"
              hx-on::after-request="if(event.detail.successful) htmx.remove(htmx.find('#contact-modal'))">
            <input type="hidden" name="id" value="{{.Contact.ID}}">
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="contactType">Contact Type</label>
                <select id="contactType" name="ContactType" class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline">
                    <option value="Personal" {{if eq .Contact.ContactType "Personal"}}selected{{end}}>Personal</option>
                    <option value="Work" {{if eq .Contact.ContactType "Work"}}selected{{end}}>Work</option>
                    <option value="Family" {{if eq .Contact.ContactType "Family"}}selected{{end}}>Family</option>
                </select>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="firstName">First Name</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="firstName" name="FirstName" type="text" value="{{.Contact.FirstName}}" required>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="lastName">Last Name</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="lastName" name="LastName" type="text" value="{{.Contact.LastName}}" required>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="email">Email</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="email" name="Email" type="email" value="{{.Contact.Email}}" required>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="phone">Phone</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="phone" name="Phone" type="tel" value="{{.Contact.Phone}}" required>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="company">Company</label>
                <select id="company" name="CompanyID" class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline">
                    <option value="">No Company</option>
                    {{range .Companies}}
                    <option value="{{.ID}}" {{if $.Contact.CompanyID}}{{if eq .ID $.Contact.CompanyID}}selected{{end}}{{end}}>{{.Name}}</option>
                    {{end}}
                </select>
            </div>
            <div class="mb-4">
                <label class="block text-gray-700 text-sm font-bold mb-2" for="password">Password</label>
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                       id="password" name="Password" type="password" placeholder="Leave empty to keep current"
                       value="{{if .Contact.Password}}{{.Contact.Password}}{{end}}">
                <p class="text-xs text-gray-500 mt-1">Leave empty to keep current password</p>
            </div>
            <div class="flex items-center justify-end">
                <button type="button" hx-target="#contact-modal" hx-swap="outerHTML" hx-get="/modal/close" class="bg-gray-500 text-white font-bold py-2 px-4 rounded-lg shadow-md hover:bg-gray-600 transition-colors duration-300 mr-2">Cancel</button>
                <button type="submit" class="bg-blue-600 text-white font-bold py-2 px-4 rounded-lg shadow-md hover:bg-blue-700 transition-colors duration-300">Save Changes</button>
            </div>
        </form>
    </div>
</div>
`

var changePasswordHTML = `
<!doctype HTML>
<html lang="en">
	<head>
		<meta charset="UTF-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<title>Change Password - AFCB</title>
		<script src="https://cdn.tailwindcss.com"></script>
        <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    </head>
    <body class="bg-gray-200 flex items-center justify-center min-h-screen">
        <div class="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
            <h2 class="text-2xl font-bold text-center text-gray-800 mb-6">
                Change Your Password
            </h2>
            <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
                <p class="text-yellow-800 text-sm">For security reasons, please change your default password.</p>
            </div>
            <form
                hx-post="/change-password"
                hx-trigger="submit"
                hx-target="#password-message"
                hx-swap="innerHTML"
            >
                <div class="mb-4">
                	<label class="block text-gray-700 font-bold mb-2" for="newPassword">New Password</label>
                    <input
                         class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                        id="newPassword"
                        name="newPassword"
                        type="password"
                        placeholder="Enter new password"
                        required
                        minlength="6"
                    />
                </div>
                <div class="mb-6">
                    <label class="block text-gray-700 font-bold mb-2" for="confirmPassword">Confirm New Password</label>
                    <input
                        class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                        id="confirmPassword"
                        name="confirmPassword"
                        type="password"
                        placeholder="Confirm new password"
                        required
                        minlength="6"
                        />
                </div>
                <div class="flex items-center justify-between">
                     <button
                        class="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline w-full"
                        type="submit"
                    >
                        Change Password
                    </button>
                </div>
            </form>
            <div id="password-message" class="mt-4 text-center"></div>
        </div>
    </body>
</html>
`

func renderCard(w http.ResponseWriter, c Contact) {
	w.Header().Set("Content-Type", "text/html")
	conCard.Execute(w, c)
}

func getCompanies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	companies, err := db.GetCompanies()
	if err != nil {
		http.Error(w, "Failed to fetch companies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Returning %d companies\n", len(companies))

	if len(companies) == 0 {
		fmt.Fprintf(w, `<div class="flex items-center justify-center p-8 bg-gray-100 text-gray-500 rounded-lg shadow-md">
            No companies found. Add your first company!
        </div>`)
		return
	}

	for _, company := range companies {
		fmt.Fprintf(w, `
        <div class="bg-white rounded-lg shadow-md p-6 mb-4" id="company-%s">
            <h3 class="text-lg font-bold text-gray-800">%s</h3>
            <div class="mt-2 text-sm text-gray-600">
                <p><strong>Bank:</strong> %s</p>
                <p><strong>Account:</strong> %s</p>
                <p><strong>Registration:</strong> %s</p>
            </div>
            <div class="mt-4 flex justify-end space-x-2">
                <button class="text-blue-600 hover:text-blue-800"
                        hx-get="/modal/edit-company/%s"
                        hx-target="#modal-container"
                        hx-swap="innerHTML">Edit</button>
                <button class="text-red-600 hover:text-red-800"
                        hx-delete="/companies/%s"
                        hx-target="#company-%s"
                        hx-swap="outerHTML"
                        hx-confirm="Are you sure you want to delete this company?">Delete</button>
            </div>
        </div>`,
			company.ID, company.Name, company.BankName, company.AccountNumber,
			company.RegistrationNumber, company.ID, company.ID, company.ID)
	}
}

func addCompanyModal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.New("company-modal").Parse(addCompanyModalHTML))
	tmpl.Execute(w, nil)
}

func addCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//parse multipart form for file uploads
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		//32MB max
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	//gen company ID
	id, err := genID()
	if err != nil {
		http.Error(w, "Failed to generate ID", http.StatusInternalServerError)
		return
	}

	//handle file uploads
	accountDoc, err := handleFileUpload(r, "account_document")
	if err != nil {
		http.Error(w, "Failed to upload account document: "+err.Error(), http.StatusInternalServerError)
		return
	}

	registrationDoc, err := handleFileUpload(r, "registration_document")
	if err != nil {
		//clean uploaded file if fails
		if accountDoc != "" {
			deleteUploadedFile(accountDoc)
		}
		http.Error(w, "Failed to upload registration document: "+err.Error(), http.StatusInternalServerError)
		return
	}

	company := &Company{
		ID:                       id,
		Name:                     r.FormValue("name"),
		BankName:                 r.FormValue("bank_name"),
		AccountNumber:            r.FormValue("account_number"),
		AccountDocumentPath:      accountDoc,
		RegistrationNumber:       r.FormValue("registration_number"),
		RegistrationDocumentPath: registrationDoc,
	}

	if err := db.CreateCompany(company); err != nil {
		if accountDoc != "" {
			deleteUploadedFile(accountDoc)
		}
		if registrationDoc != "" {
			deleteUploadedFile(registrationDoc)
		}
		http.Error(w, "Failed to create company "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<div class="bg-white rounded-lg shadow-md p-6 mb-4" id="company-%s">
        <h3 class="text-lg font-bold text-gray-800">%s</h3>
        <div class="mt-2 text-sm text-gray-600">
            <p><strong>Bank:</strong> %s</p>
            <p><strong>Account:</strong> %s</p>
            <p><strong>Registration:</strong> %s</p>
        </div>
        <div class="mt-4 flex justify-end space-x-2">
            <button class="text-blue-600 hover:text-blue-800">Edit</button>
            <button class="text-red-600 hover:text-red-800">Delete</button>
        </div>
    </div>`, company.ID, company.Name, company.BankName, company.AccountNumber, company.RegistrationNumber)
}

func getContacts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	fmt.Printf("=== GET /contacts called ===\n")

	contacts, err := db.GetAllContacts()
	if err != nil {
		http.Error(w, "Failed to fetch contacts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Returning %d contacts to client\n", len(contacts))

	if len(contacts) == 0 {
		fmt.Printf("No contacts found, returning empty message\n")
		fmt.Fprintf(w, `<div class="flex items-center justify-center p-8 bg-gray-100 text-gray-500 rounded-lg shadow-md">
            No contacts found. Add your first contact!
        </div>`)
		return
	}

	cardRendered := 0
	for _, c := range contacts {
		fmt.Printf("Rendering contact: %s %s (ID: %s)\n", c.FirstName, c.LastName, c.ID)
		if err := conCard.Execute(w, c); err != nil {
			fmt.Printf("Error rendering contact %s: %v\n", c.ID, err)
			continue // Continue with next contact instead of failing completely
		}
		cardRendered++
	}
	fmt.Printf("Successfully rendered %d contacts\n", cardRendered)
	fmt.Printf("=== END GET /contacts ===\n")
}

func addContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	contactType := r.FormValue("ContactType")
	firstName := r.FormValue("FirstName")
	lastName := r.FormValue("LastName")
	email := r.FormValue("Email")
	phone := r.FormValue("Phone")
	password := r.FormValue("Password")
	companyID := r.FormValue("CompanyID")

	fmt.Printf("Received form data - Type: '%s', Name: '%s %s', Email: '%s', Phone: '%s', Password: '%s', CompanyID: '%s'\n",
		contactType, firstName, lastName, email, phone, password, companyID)

	// Validate required fields
	if firstName == "" || lastName == "" || email == "" || phone == "" {
		fmt.Printf("Missing required fields: FirstName='%s', LastName='%s', Email='%s', Phone='%s'\n",
			firstName, lastName, email, phone)
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Email validation
	if !emailRegex.MatchString(email) {
		http.Error(w, "Invalid email address format", http.StatusBadRequest)
		return
	}

	// Generate ID for new contact
	newID, err := genID()
	if err != nil {
		http.Error(w, "Failed to generate ID", http.StatusInternalServerError)
		return
	}

	// Generate default password if not provided
	if password == "" {
		// Get last 3 characters of phone
		phoneLast3 := ""
		if len(phone) >= 3 {
			phoneLast3 = phone[len(phone)-3:]
		} else {
			phoneLast3 = phone
		}

		// Get first letter of first name (uppercase)
		firstNameFirstLetter := ""
		if len(firstName) > 0 {
			firstNameFirstLetter = strings.ToUpper(string(firstName[0]))
		}

		password = phoneLast3 + firstNameFirstLetter
		fmt.Printf("Auto-generated password: %s (from phone: %s, first name: %s)\n", password, phone, firstName)
	}

	// Handle companyID for new contact
	var companyIDPtr *string
	if companyID != "" {
		companyIDPtr = &companyID
	}

	newContact := &Contact{
		ID:          newID,
		ContactType: contactType,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		Phone:       phone,
		Password:    password,
		CompanyID:   companyIDPtr,
	}

	if err := db.CreateContact(newContact); err != nil {
		fmt.Printf("Error creating contact: %v\n", err)
		http.Error(w, "Failed to create contact: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create user account for this contact
	user := &User{
		Username:           email,
		Password:           password,
		ContactID:          &newID,
		NeedPasswordChange: true,
	}

	// Check if user already exists
	_, err = db.GetUser(email)
	if err != nil {
		// User doesn't exist, create new one
		if err := db.CreateUser(user); err != nil {
			fmt.Printf("Warning: Failed to create user account for contact: %v\n", err)
		} else {
			fmt.Printf("User account created for contact: %s\n", email)
		}
	} else {
		fmt.Printf("Warning: User account already exists for email: %s\n", email)
	}

	fmt.Printf("New contact created: %s %s (ID: %s)\n", newContact.FirstName, newContact.LastName, newContact.ID)
	renderCard(w, *newContact)
}

func updateContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Printf("UPDATE request received for id: %s\n", id)

	// Get existing contact
	contact, err := db.GetContact(id)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	// Update fields
	contact.ContactType = r.FormValue("ContactType")
	contact.FirstName = r.FormValue("FirstName")
	contact.LastName = r.FormValue("LastName")
	contact.Email = r.FormValue("Email")
	contact.Phone = r.FormValue("Phone")
	password := r.FormValue("Password")

	// Only update password if provided
	if password != "" {
		contact.Password = password
	}

	// Validate required fields
	if contact.ContactType == "" || contact.FirstName == "" || contact.LastName == "" ||
		contact.Email == "" || contact.Phone == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Email validation
	if !emailRegex.MatchString(contact.Email) {
		http.Error(w, "Invalid email address format", http.StatusBadRequest)
		return
	}

	fmt.Printf("Attempting to update contact %s\n", id)

	// Update contact in database
	if err := db.UpdateContact(contact); err != nil {
		fmt.Println("Update error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update user account
	user := &User{
		Username:  contact.Email,
		Password:  contact.Password,
		ContactID: &id,
	}

	existingUser, err := db.GetUser(contact.Email)
	if err != nil {
		// Create new user if doesn't exist
		if err := db.CreateUser(user); err != nil {
			fmt.Printf("Warning: Failed to create user account: %v\n", err)
		}
	} else {
		// Update existing user
		if existingUser.Password != user.Password {
			if err := db.UpdateUserPassword(contact.Email, user.Password); err != nil {
				fmt.Printf("Warning: Failed to update user password: %v\n", err)
			}
		}
	}

	fmt.Printf("Successfully updated contact: %+v\n", contact)
	renderCard(w, *contact)
}

func deleteContact(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	fmt.Println("DELETE request received for id:", id)

	// Get contact to find associated user
	contact, err := db.GetContact(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Delete contact
	if err := db.DeleteContact(id); err != nil {
		fmt.Println("Delete error:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Also delete associated user
	if err := db.DeleteUser(contact.Email); err != nil {
		fmt.Printf("Warning: Failed to delete user account: %v\n", err)
	}

	// Return empty content - HTMX will remove element
	w.WriteHeader(http.StatusOK)
}

func searchContacts(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("q")
	fmt.Printf("Search request received for keyword: '%s'\n", keyword)

	w.Header().Set("Content-Type", "text/html")

	var results []Contact
	var err error

	if keyword == "" {
		fmt.Println("No keyword provided, returning all contacts")
		results, err = db.GetAllContacts()
	} else {
		results, err = db.SearchContacts(keyword)
	}

	if err != nil {
		http.Error(w, "Search failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Found %d results for keyword '%s'\n", len(results), keyword)

	if len(results) == 0 {
		fmt.Fprintf(w, `<div class="no-results text-center p-8 text-gray-500">No contacts found for "%s"</div>`, template.HTMLEscapeString(keyword))
		return
	}

	for _, c := range results {
		if err := conCard.Execute(w, c); err != nil {
			fmt.Printf("Error rendering contact %s: %v\n", c.ID, err)
			continue
		}
	}
}

// PW CHANGE HANDLER
func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		_, err := r.Cookie("password_change_user")
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<div class="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
                    <h2 class="text-2xl font-bold text-center text-gray-800 mb-6">Access Denied</h2>
                    <div class="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
                        <p class="text-red-800 text-sm">Please login first to change your password.</p>
                    </div>
                    <a href="/login" class="block text-center bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700">Go to Login</a>
                </div>`))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(changePasswordHTML))
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		newPassword := r.FormValue("newPassword")
		confirmPassword := r.FormValue("confirmPassword")

		//Get username from cookie
		cookie, err := r.Cookie("password_change_user")
		if err != nil {
			w.Write([]byte(`<div class="text-red-500">Session expired. Please login again.</div>`))
			return
		}
		username := cookie.Value

		//Valudate password
		if newPassword != confirmPassword {
			w.Write([]byte(`<div class="text-red-500">Password must be at least 6 characters long.</div>`))
			return
		}

		if len(newPassword) < 6 {
			w.Write([]byte(`<div class="text-red-500">Password must be at least 6 characters long.</div>`))
			return
		}

		//Update pw in db
		if err := db.UpdateUserPassword(username, newPassword); err != nil {
			w.Write([]byte(`<div class="text-red-500">Failed to update password. Please try again</div>`))
			return
		}

		//clear cookie
		http.SetCookie(w, &http.Cookie{
			Name:   "password_change_user",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})

		//Update contact pw if it's a contact user
		user, err := db.GetUser(username)
		if err == nil && user.ContactID != nil {
			contact, err := db.GetContact(*user.ContactID)
			if err == nil {
				contact.Password = newPassword
				db.UpdateContact(contact)
			}
		}

		fmt.Printf("Password successfullt changed for user: %s\n", username)
		w.Write([]byte(`<div class="text-green-500">Password updated succesfully! Redirecting...</div>
			<script>setTimeout(() => window.location.href = "/", 2000)</script>`))
		return
	}
}

// MODAL HANDLERS
func addModal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	companies, err := db.GetCompanies()
	if err != nil {
		http.Error(w, "Failed to fetch companies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Companies []Company
	}{
		Companies: companies,
	}

	tmpl := template.Must(template.New("modal").Parse(addModalHTML))
	if err := tmpl.Execute(w, data); err != nil {
		fmt.Printf("Template execution error: %v\n", err)
		http.Error(w, "Failed to render modal", http.StatusInternalServerError)
		return
	}
}

func editModal(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	contact, err := db.GetContact(id)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	companies, err := db.GetCompanies()
	if err != nil {
		http.Error(w, "Failed to fetch companies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Contact   *Contact
		Companies []Company
	}{
		Contact:   contact,
		Companies: companies,
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.New("edit-modal").Parse(editModalHTML))
	tmpl.Execute(w, data)
}

func closeForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
}

// AUTH HANDLERS
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	fmt.Printf("Login attempt: username=%s\n", username)

	user, err := db.GetUser(username)
	if err != nil {
		fmt.Printf("Login failed - user not found or error: %v\n", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`<div id="login-message" class="mt-4 text-center text-red-500">Invalid username or password.</div>`))
		return
	}

	fmt.Printf("Found user: %s, needs password change: %t\n", user.Username, user.NeedPasswordChange)

	if user.Password == password {
		fmt.Printf("Login successful for user: %s\n", username)
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: "authenticated",
			Path:  "/",
		})

		//Check if need password change
		if user.NeedPasswordChange {
			fmt.Printf("User %s needs passowrd change\n", username)
			//Cookie indicator password change needed
			http.SetCookie(w, &http.Cookie{
				Name:   "password_change_user",
				Value:  username,
				Path:   "/",
				MaxAge: 300,
			})
			w.Header().Set("HX-Redirect", "/change-password")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Login successful - password change required"))
			return
		}

		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Login successful"))
		return
	}

	fmt.Printf("Login failed - password mismatch for user: %s\n", username)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`<div id="login-message" class="mt-4 text-center text-red-500">Invalid username or password.</div>`))
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow login page without authentication
		if r.URL.Path == "/login" || r.URL.Path == "/change-password" {
			next.ServeHTTP(w, r)
			return
		}

		//check if user authenticated
		sessionCookie, err := r.Cookie("session")
		if err != nil || sessionCookie.Value != "authenticated" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		//check if user need password change
		if _, err := r.Cookie("password_change_user"); err == nil {
			http.Redirect(w, r, "/change-password", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize database
	var err error
	db, err = InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	fmt.Println("Database initialized successfully")

	// Debug: users table
	if err := db.DebugUserTable(); err != nil {
		fmt.Printf("Debug error: %v\n", err)
	}

	router := mux.NewRouter()

	// Debug: List all users
	rows, err := db.Query("SELECT username, password FROM users")
	if err != nil {
		fmt.Printf("Error querying users: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Println("Current users in database:")
		for rows.Next() {
			var username, password string
			if err := rows.Scan(&username, &password); err != nil {
				fmt.Printf("Error scanning user row: %v\n", err)
				continue
			}
			fmt.Printf("  User: %s, Password: %s\n", username, password)
		}
	}

	// Serve login page
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.ServeFile(w, r, "static/login.html")
		} else if r.Method == "POST" {
			loginHandler(w, r)
		}
	})

	router.HandleFunc("/change-password", changePasswordHandler).Methods("GET", "POST")
	router.HandleFunc("/logout", logoutHandler).Methods("GET")

	// Create sub-router for all authenticated routes
	authRouter := router.PathPrefix("/").Subrouter()
	authRouter.Use(authMiddleware)

	authRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// Static file server
	authRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	// Contact API endpoints
	authRouter.HandleFunc("/contacts", getContacts).Methods("GET")
	authRouter.HandleFunc("/contacts", addContact).Methods("POST")
	authRouter.HandleFunc("/contacts/{id}", updateContact).Methods("PUT", "PATCH")
	authRouter.HandleFunc("/contacts/{id}", deleteContact).Methods("DELETE")

	// Modal endpoints
	authRouter.HandleFunc("/modal/add", addModal).Methods("GET")
	authRouter.HandleFunc("/modal/edit/{id}", editModal).Methods("GET")
	authRouter.HandleFunc("/modal/close", closeForm).Methods("GET")

	authRouter.HandleFunc("/modal/add-company", addCompanyModal).Methods("GET")
	authRouter.HandleFunc("/companies", addCompany).Methods("POST")
	authRouter.HandleFunc("/companies", getCompanies).Methods("GET")

	// Search endpoint
	authRouter.HandleFunc("/search", searchContacts).Methods("GET")

	// Server start
	fmt.Println("AFcb started at http://localhost:1330")
	fmt.Println("Default admin login: af / afcb")
	log.Fatal(http.ListenAndServe(":1330", router))
}
