package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var db *DB

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

const dataFile = "AFcb.db" // Now using SQLite database

var conCard = template.Must(template.New("card").Funcs(template.FuncMap{
	"getCompanyName": func(companyID *string) string {
		if companyID == nil {
			return ""
		}
		company, err := db.GetCompany(*companyID)
		if err != nil {
			return ""
		}
		return company.Name
	},
}).Parse(`
    <div class="card bg-white rounded-xl shadow-md p-6 hover:shadow-lg transition-all duration-300" id="contact-{{.Contact.ID}}">
    <div class="details">
        <span class="id text-xs font-semibold text-gray-500">ID: {{.Contact.ID}}</span>
        <strong class="name block text-xl font-bold text-gray-800 mt-1">{{.Contact.FirstName}} {{.Contact.LastName}}</strong>
        {{if .Contact.CompanyID}}
        <div class="company mt-1">
            <span class="text-sm text-gray-600">{{getCompanyName .Contact.CompanyID}}</span>
        </div>
        {{end}}
        <span class="type inline-block mt-2 px-3 py-1 rounded-full text-sm font-medium
            {{if .IsCurrentUser}}bg-yellow-100 text-yellow-800
            {{else if eq .Contact.ContactType "Personal"}}bg-blue-100 text-blue-800
            {{else if eq .Contact.ContactType "Work"}}bg-green-100 text-green-800
            {{else if eq .Contact.ContactType "Family"}}bg-purple-100 text-purple-800
            {{else}}bg-gray-100 text-gray-800{{end}}">
            {{if .IsCurrentUser}}Myself{{else}}{{.Contact.ContactType}}{{end}}
        </span>
        <div class="details mt-3 text-gray-600">
            <div class="flex items-center mb-1">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
                <span id="email-{{.Contact.ID}}">{{.Contact.Email}}</span>
                <button onclick="copyEmail('email-{{.Contact.ID}}')" class="ml-2 p-1 rounded-full hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500" title="Copy Email">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2.5a1.5 1.5 0 011.5 1.5v4.5m-14-6.5h3v-3h-3v3z" />
                    </svg>
                </button>
            </div>
            <div class="flex items-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                </svg>
                <span>{{.Contact.Phone}}</span>
                <a href="https://wa.me/{{.Contact.Phone}}" target="_blank" class="ml-2 p-1 rounded-full text-green-500 hover:bg-green-100 transition-colors" title="WhatsApp">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12.04 2.87c-5.42 0-9.82 4.4-9.82 9.82 0 1.94.57 3.8.14 5.39l-1.39 5.09 5.25-1.36c1.5.25 3.09.4 4.56.4 5.42 0 9.82-4.4 9.82-9.82-.01-5.42-4.4-9.81-9.8-9.81zm-.04 17.1c-1.36 0-2.7-.22-3.9-.66l-2.61.68.68-2.55c-.5-1.16-.76-2.43-.76-3.75 0-4.41 3.59-8 8-8s8 3.59 8 8-3.59 8-8 8zm4.53-5.59c-.25-.13-.49-.2-.72-.2-.23 0-.46.07-.69.21-.23.14-.52.28-.84.38-.32.1-.64.16-.96.06-.32-.1-.6-.24-.87-.45-.27-.2-.5-.45-.7-.7-.19-.24-.34-.49-.49-.77s-.27-.58-.33-.89c-.06-.31-.05-.59-.01-.84.04-.26.13-.5.26-.72.13-.22.25-.4.36-.57.11-.17.18-.32.22-.44.04-.12.02-.27-.04-.43-.06-.16-.18-.32-.34-.48-.16-.16-.36-.31-.6-.44-.24-.13-.49-.2-.73-.2-.24 0-.48.05-.72.15-.24.1-.46.25-.66.44-.2.19-.38.41-.54.67-.16.26-.28.53-.4.81s-.2 0-.25-.06c-.05-.06-.2-.25-.37-.47s-.35-.4-.5-.54c-.16-.14-.28-.2-.37-.2s-.22 0-.36-.05c-.14-.05-.3-.08-.5-.09-.19-.01-.39-.01-.58 0-.19 0-.4.04-.61.09-.2.05-.4.14-.57.26-.17.12-.3.27-.4.45-.1.18-.15.39-.15.63s.06.48.19.74c.12.26.3.52.54.78.24.26.54.55.89.87.35.31.75.63 1.18.96 1.05.78 1.95 1.48 2.5 1.77.55.29 1.01.44 1.39.44.38 0 .82-.13 1.34-.38.52-.25.96-.54 1.33-.88.37-.34.6-.78.71-1.32.11-.54.06-1.04-.08-1.52z"/>
                    </svg>
                </a>
                <a href="signal://send?text=&phone={{.Contact.Phone}}" target="_blank" class="ml-2 p-1 rounded-full text-gray-800 hover:bg-gray-200 transition-colors" title="Signal">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm.8 14.8c-.37.37-.87.5-1.37.5-.5 0-1-.13-1.37-.5-.75-.75-.75-1.99 0-2.74L12 11.39l-1.44-1.44c-.75-.75-.75-1.99 0-2.74s1.99-.75 2.74 0L12 8.61l1.44-1.44c.75-.75 1.99-.75 2.74 0s.75 1.99 0 2.74L12.8 12.8l1.44 1.44c.75.75.75 1.99 0 2.74zm0 0"/>
                    </svg>
                </a>
            </div>
        </div>
    </div>
    <div class="actions flex justify-end mt-4 space-x-2">
   		<button class="pdf-btn p-2 rounded-lg border border-gray-300 hover:border-green-500 hover:bg-green-50 transition-colors"
            onclick="generateContactPDF('{{.Contact.ID}}')"
            title="Download PDF">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            	<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
            </svg>
        </button>
        <button class="edit-btn p-2 rounded-lg border border-gray-300 hover:border-blue-500 hover:bg-blue-50 transition-colors"
            hx-get="/modal/edit/{{.Contact.ID}}"
            hx-target="#modal-container"
            hx-swap="innerHTML"
            title="Edit">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 20h9"/>
                <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4 12.5-12.5z"/>
            </svg>
        </button>
        <button class="delete-btn p-2 rounded-lg border border-gray-300 hover:border-red-500 hover:bg-red-50 transition-colors"
                hx-delete="/contacts/{{.Contact.ID}}"
                hx-target="#contact-{{.Contact.ID}}"
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
            <button hx-target="#company-modal" hx-swap="outerHTML" hx-get="/modal/close" class="text-gray-400 hover:text-gray-600 text-2xl">&times;</button>
        </div>
        <h3 class="text-xl font-bold mb-4">Add New Company</h3>
        <form id="companyForm" enctype="multipart/form-data"
              hx-post="/companies"
              hx-target="#companies-table-body"
              hx-swap="beforeend"
              hx-on::after-request="if(event.detail.successful) { document.getElementById('company-modal').remove(); }">
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
              hx-on::after-request="if(event.detail.successful) document.getElementById('contact-modal').remove()">
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
                <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                       id="email" name="Email" type="email" placeholder="Email" required>
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
                    <option value="{{.ID}}" {{if and $.Contact.CompanyID (eq .ID $.Contact.CompanyID)}}selected{{end}}>{{.Name}}</option>
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

// License activation handler
func activateLicenseHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !isAdmin(r) {
		http.Error(w, "Forbidden - Admin access required", http.StatusForbidden)
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		licenseKey := r.FormValue("licenseKey")
		if licenseKey == "" {
			fmt.Fprintf(w, `
                <div class="bg-red-50 border border-red-200 rounded-xl p-4">
                    <div class="flex items-center">
                        <svg class="w-5 h-5 text-red-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"/>
                        </svg>
                        <p class="text-red-700 font-medium">License key is required</p>
                    </div>
                </div>
            `)
			return
		}

		// Validate the license
		licenseManager, err := NewLicenseManager()
		if err != nil {
			fmt.Fprintf(w, `
                <div class="bg-red-50 border border-red-200 rounded-xl p-4">
                    <div class="flex items-center">
                        <svg class="w-5 h-5 text-red-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"/>
                        </svg>
                        <p class="text-red-700 font-medium">License system error: %v</p>
                    </div>
                </div>
            `, err)
			return
		}

		license, err := licenseManager.ValidateLicense(licenseKey)
		if err != nil {
			fmt.Fprintf(w, `
                <div class="bg-red-50 border border-red-200 rounded-xl p-4">
                    <div class="flex items-center">
                        <svg class="w-5 h-5 text-red-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"/>
                        </svg>
                        <p class="text-red-700 font-medium">Invalid license: %v</p>
                    </div>
                </div>
            `, err)
			return
		}

		// Save license to environment or database
		// For now, we'll just validate and show success

		fmt.Fprintf(w, `
        <div class="bg-green-50 border border-green-200 rounded-xl p-6">
            <div class="flex items-center mb-4">
                <div class="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center mr-4">
                    <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                    </svg>
                </div>
                <div>
                    <h3 class="text-lg font-semibold text-green-800">License Activated Successfully!</h3>
                    <p class="text-green-600">Your license has been validated and activated.</p>
                </div>
            </div>

            <div class="bg-white rounded-lg p-4 border border-green-100">
                <div class="grid grid-cols-2 gap-4 text-sm">
                    <div>
                        <p class="text-gray-500 font-medium">Company</p>
                        <p class="text-gray-900 font-semibold">%s</p>
                    </div>
                    <div>
                        <p class="text-gray-500 font-medium">License Type</p>
                        <p class="text-gray-900 font-semibold">%s</p>
                    </div>
                    <div>
                        <p class="text-gray-500 font-medium">Expiration</p>
                        <p class="text-gray-900 font-semibold">%s</p>
                    </div>
                    <div>
                        <p class="text-gray-500 font-medium">Max Users</p>
                        <p class="text-gray-900 font-semibold">%d</p>
                    </div>
                    <div>
                        <p class="text-gray-500 font-medium">Domain</p>
                        <p class="text-gray-900 font-semibold">%s</p>
                    </div>
                    <div>
                        <p class="text-gray-500 font-medium">Issue Date</p>
                        <p class="text-gray-900 font-semibold">%s</p>
                    </div>
                </div>
            </div>

            <div class="mt-4 bg-blue-50 rounded-lg p-4">
                <p class="text-blue-700 text-sm">
                    <strong>Note:</strong> For permanent activation across server restarts, set the
                    <code class="bg-blue-100 px-1 rounded">AFCB_LICENSE_KEY</code> environment variable to this license key.
                </p>
            </div>
        </div>
        `, license.CompanyName, license.LicenseType,
			license.ExpiryDate.Format("January 2, 2006"),
			license.MaxUsers, license.Domain,
			license.IssueDate.Format("January 2, 2006"))
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	isAdmin := false
	currentUser, err := getCurrentUser(r)
	if err == nil && currentUser == "af" {
		isAdmin = true
	}

	data := struct {
		IsAdmin bool
	}{
		IsAdmin: isAdmin,
	}

	tmpl := template.Must(template.ParseFiles("static/index.html"))
	tmpl.Execute(w, data)
}

// helper for admin check
func isAdmin(r *http.Request) bool {
	currentUser, err := getCurrentUser(r)
	if err != nil {
		return false
	}
	return currentUser == "af"
}

func licenseContentHandler(w http.ResponseWriter, r *http.Request) {
	licenseAdminHandler(w, r) // This will render the actual license content
}

func licenseAdminHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !isAdmin(r) {
		http.Error(w, "Forbidden - Admin access required", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	licenseManager, err := NewLicenseManager()
	if err != nil {
		fmt.Fprintf(w, `
            <div class="bg-red-50 border border-red-200 rounded-xl p-6">
                <div class="flex items-center">
                    <svg class="w-6 h-6 text-red-500 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"/>
                    </svg>
                    <h3 class="text-lg font-semibold text-red-800">License System Error</h3>
                </div>
                <p class="mt-2 text-red-600">%v</p>
            </div>`, err)
		return
	}

	licenseKey := os.Getenv("AFCB_LICENSE_KEY")

	fmt.Fprintf(w, `
    <div class="max-w-6xl mx-auto px-4 py-8">
        <!-- Header -->
        <div class="text-center mb-12">
            <h1 class="text-3xl font-bold text-gray-900 mb-2">License Management</h1>
            <p class="text-gray-600">Manage your AFCB license and activation</p>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
            <!-- Current License Card -->
            <div class="bg-white rounded-2xl shadow-lg border border-gray-200 p-6">
                <div class="flex items-center mb-6">
                    <div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mr-4">
                        <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
                        </svg>
                    </div>
                    <h2 class="text-xl font-semibold text-gray-900">Current License Status</h2>
                </div>
    `)

	if licenseKey == "" {
		fmt.Fprintf(w, `
                <div class="bg-yellow-50 border border-yellow-200 rounded-xl p-6 text-center">
                    <div class="w-16 h-16 bg-yellow-100 rounded-full flex items-center justify-center mx-auto mb-4">
                        <svg class="w-8 h-8 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"/>
                        </svg>
                    </div>
                    <h3 class="text-lg font-semibold text-yellow-800 mb-2">Trial Mode Active</h3>
                    <p class="text-yellow-600 mb-4">You're currently running with limited features. Activate a license to unlock all capabilities.</p>
                    <div class="bg-yellow-100 rounded-lg p-3">
                        <p class="text-yellow-700 text-sm font-medium">Features may be restricted in trial mode</p>
                    </div>
                </div>
        `)
	} else {
		license, err := licenseManager.ValidateLicense(licenseKey)
		if err != nil {
			fmt.Fprintf(w, `
                <div class="bg-red-50 border border-red-200 rounded-xl p-6">
                    <div class="flex items-center mb-4">
                        <svg class="w-6 h-6 text-red-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                        <h3 class="text-lg font-semibold text-red-800">Invalid License</h3>
                    </div>
                    <p class="text-red-600 mb-4">The current license key is invalid or corrupted.</p>
                    <div class="bg-red-100 rounded-lg p-3">
                        <p class="text-red-700 text-sm">Error: %v</p>
                    </div>
                </div>
            `, err)
		} else {
			// Check license status
			isExpired := time.Now().After(license.ExpiryDate)
			statusIcon := "✅"
			statusText := "Active"
			badgeClass := "bg-green-100 text-green-800"

			if isExpired {
				statusIcon = "⚠️"
				statusText = "Expired"
				badgeClass = "bg-red-100 text-red-800"
			} else if license.LicenseType == "trial" {
				statusIcon = "⏱️"
				statusText = "Trial"
				badgeClass = "bg-blue-100 text-blue-800"
			}

			daysLeft := int(time.Until(license.ExpiryDate).Hours() / 24)

			fmt.Fprintf(w, `
                <div class="space-y-4">
                    <div class="flex justify-between items-start">
                        <div>
                            <h3 class="text-lg font-semibold text-gray-900">%s</h3>
                            <p class="text-gray-600">%s License</p>
                        </div>
                        <span class="px-3 py-1 rounded-full text-sm font-semibold %s">%s %s</span>
                    </div>

                    <div class="grid grid-cols-2 gap-4 text-sm">
                        <div class="bg-gray-50 rounded-lg p-3">
                            <p class="text-gray-500 font-medium">Company</p>
                            <p class="text-gray-900 font-semibold">%s</p>
                        </div>
                        <div class="bg-gray-50 rounded-lg p-3">
                            <p class="text-gray-500 font-medium">Email</p>
                            <p class="text-gray-900 font-semibold">%s</p>
                        </div>
                        <div class="bg-gray-50 rounded-lg p-3">
                            <p class="text-gray-500 font-medium">Expires</p>
                            <p class="text-gray-900 font-semibold">%s</p>
                        </div>
                        <div class="bg-gray-50 rounded-lg p-3">
                            <p class="text-gray-500 font-medium">Max Users</p>
                            <p class="text-gray-900 font-semibold">%d</p>
                        </div>
                    </div>
                `, license.CompanyName, license.LicenseType, badgeClass, statusIcon, statusText,
				license.CompanyName, license.Email,
				license.ExpiryDate.Format("January 2, 2006"),
				license.MaxUsers)

			if !isExpired {
				if daysLeft <= 30 {
					fmt.Fprintf(w, `
                        <div class="bg-orange-50 border border-orange-200 rounded-lg p-4">
                            <div class="flex items-center">
                                <svg class="w-5 h-5 text-orange-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
                                </svg>
                                <p class="text-orange-700 font-medium">%d days remaining</p>
                            </div>
                        </div>
                    `, daysLeft)
				}
			} else {
				fmt.Fprintf(w, `
                    <div class="bg-red-50 border border-red-200 rounded-lg p-4">
                        <div class="flex items-center">
                            <svg class="w-5 h-5 text-red-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                            </svg>
                            <p class="text-red-700 font-medium">License expired on %s</p>
                        </div>
                    </div>
                `, license.ExpiryDate.Format("January 2, 2006"))
			}

			fmt.Fprintf(w, `</div>`)
		}
	}

	fmt.Fprintf(w, `
            </div>

            <!-- Activation Card -->
            <div class="bg-white rounded-2xl shadow-lg border border-gray-200 p-6">
                <div class="flex items-center mb-6">
                    <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center mr-4">
                        <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                        </svg>
                    </div>
                    <h2 class="text-xl font-semibold text-gray-900">Activate License</h2>
                </div>

                <div class="space-y-6">
                    <div>
                        <label for="licenseKey" class="block text-sm font-medium text-gray-700 mb-2">
                            License Key
                        </label>
                        <textarea
                            id="licenseKey"
                            name="licenseKey"
                            rows="8"
                            class="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-none font-mono text-sm"
                            placeholder="Paste your license key here..."
                            required></textarea>
                        <p class="mt-2 text-sm text-gray-500">
                            Enter the complete license key provided for your organization.
                        </p>
                    </div>

                    <div class="flex items-center justify-between pt-4 border-t border-gray-200">
                        <div class="text-sm text-gray-500">
                            Need a license? Contact support
                        </div>
                        <button
                            onclick="activateLicense()"
                            class="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-semibold py-3 px-8 rounded-xl shadow-lg transition-all duration-200 transform hover:scale-105 focus:ring-4 focus:ring-blue-200">
                            Activate License
                        </button>
                    </div>
                </div>

                <div id="license-result" class="mt-6"></div>
            </div>
        </div>

        <!-- Information Section -->
        <div class="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-2xl p-8 border border-blue-200">
            <div class="text-center mb-6">
                <h2 class="text-2xl font-bold text-gray-900 mb-2">License Benefits</h2>
                <p class="text-gray-600">Unlock the full potential of AFCB</p>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <div class="text-center">
                    <div class="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center mx-auto mb-4">
                        <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
                        </svg>
                    </div>
                    <h3 class="font-semibold text-gray-900 mb-2">Full Access</h3>
                    <p class="text-sm text-gray-600">Unlock all features and capabilities</p>
                </div>

                <div class="text-center">
                    <div class="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center mx-auto mb-4">
                        <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"/>
                        </svg>
                    </div>
                    <h3 class="font-semibold text-gray-900 mb-2">User Management</h3>
                    <p class="text-sm text-gray-600">Support for multiple users based on your tier</p>
                </div>

                <div class="text-center">
                    <div class="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center mx-auto mb-4">
                        <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
                        </svg>
                    </div>
                    <h3 class="font-semibold text-gray-900 mb-2">Priority Support</h3>
                    <p class="text-sm text-gray-600">Get help when you need it most</p>
                </div>

                <div class="text-center">
                    <div class="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center mx-auto mb-4">
                        <svg class="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                        </svg>
                    </div>
                    <h3 class="font-semibold text-gray-900 mb-2">Regular Updates</h3>
                    <p class="text-sm text-gray-600">Stay current with the latest features</p>
                </div>
            </div>
        </div>

        <script>
            function activateLicense() {
                const licenseKey = document.getElementById('licenseKey').value;
                const resultDiv = document.getElementById('license-result');

                if (!licenseKey) {
                    resultDiv.innerHTML = '<div class="bg-red-50 border border-red-200 rounded-xl p-4"><div class="flex items-center"><svg class="w-5 h-5 text-red-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"/></svg><p class="text-red-700 font-medium">License key is required</p></div></div>';
                    return;
                }

                // Show loading state
                resultDiv.innerHTML = '<div class="bg-blue-50 border border-blue-200 rounded-xl p-4 text-center"><div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 mx-auto"></div><p class="mt-2 text-blue-700">Activating license...</p></div>';

                // Use HTMX to submit the form
                htmx.ajax('POST', '/admin/activate-license', {
                    values: { licenseKey: licenseKey },
                    target: '#license-result',
                    swap: 'innerHTML'
                });
            }
        </script>
    </div>
    `)
}

// Helper function to get current username from session
func getCurrentUser(r *http.Request) (string, error) {
	// First check if user is authenticated
	sessionCookie, err := r.Cookie("session")
	if err != nil || sessionCookie.Value != "authenticated" {
		return "", fmt.Errorf("not authenticated")
	}

	// Try to get username from current_user cookie first (set during login)
	if userCookie, err := r.Cookie("current_user"); err == nil {
		return userCookie.Value, nil
	}

	// Fallback: try password_change_user cookie
	if userCookie, err := r.Cookie("password_change_user"); err == nil {
		return userCookie.Value, nil
	}

	return "", fmt.Errorf("user not found in session cookies")
}

// Helper function to check if contact belongs to current user
func isCurrentUserContact(contactEmail string, r *http.Request) bool {
	// Get current logged-in user from session
	currentUser, err := getCurrentUser(r)
	if err != nil {
		fmt.Printf("DEBUG: Cannot get current user: %v\n", err)
		return false
	}

	fmt.Printf("DEBUG: Current user: %s, Contact email: %s\n", currentUser, contactEmail)

	// Direct comparison: if current user's username IS the contact email
	// This works because users log in with email as username
	if currentUser == contactEmail {
		return true
	}

	// Additional check: if the current user has a contact record, check if it matches
	user, err := db.GetUser(currentUser)
	if err != nil {
		fmt.Printf("DEBUG: Cannot get user from DB: %v\n", err)
		return false
	}

	// If user has a contact ID, check if the contact email matches
	if user.ContactID != nil {
		contact, err := db.GetContact(*user.ContactID)
		if err == nil && contact.Email == contactEmail {
			return true
		}
	}

	return false
}

func renderCard(w http.ResponseWriter, r *http.Request, c Contact) {
	w.Header().Set("Content-Type", "text/html")

	// Check if this contact belongs to the current user
	isCurrentUser := isCurrentUserContact(c.Email, r)

	data := struct {
		Contact       Contact
		IsCurrentUser bool
	}{
		Contact:       c,
		IsCurrentUser: isCurrentUser,
	}

	conCard.Execute(w, data)
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

// Search companies handler
func searchCompanies(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("q")
	fmt.Printf("Search companies request received for keyword: '%s'\n", keyword)

	w.Header().Set("Content-Type", "text/html")

	var results []Company
	var err error

	if keyword == "" {
		fmt.Println("No keyword provided, returning all companies")
		results, err = db.GetAllCompanies()
	} else {
		results, err = db.SearchCompanies(keyword)
	}

	if err != nil {
		http.Error(w, "Search failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Found %d companies for keyword '%s'\n", len(results), keyword)

	if len(results) == 0 {
		fmt.Fprintf(w, `<tr><td colspan="6" class="px-6 py-4 text-center text-gray-500">No companies found for "%s"</td></tr>`, template.HTMLEscapeString(keyword))
		return
	}

	for _, company := range results {
		// Format created date
		createdDate := company.CreatedAt
		if len(company.CreatedAt) > 10 {
			createdDate = company.CreatedAt[:10]
		}

		fmt.Fprintf(w, `
        <tr id="company-row-%s">
        <td class="px-6 py-4 whitespace-nowrap">
            <div class="text-sm font-medium text-gray-900">%s</div>
            <div class="text-sm text-gray-500">ID: %s</div>
        </td>
        <td class="px-6 py-4">
            <div class="text-sm text-gray-900"><strong>Bank:</strong> %s</div>
            <div class="text-sm text-gray-500"><strong>Account:</strong> %s</div>
            <div class="text-sm text-gray-500">
                <strong>Document:</strong>
                %s
            </div>
        </td>
        <td class="px-6 py-4">
            <div class="text-sm text-gray-900"><strong>Number:</strong> %s</div>
            <div class="text-sm text-gray-500">
                <strong>Document:</strong>
                %s
            </div>
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            %s
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            %s
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
            <button class="text-blue-600 hover:text-blue-900 mr-3 p-1 rounded hover:bg-blue-50 transition-colors"
                    hx-get="/modal/edit-company/%s"
                    hx-target="#modal-container"
                    hx-swap="innerHTML"
                    title="Edit">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                </svg>
            </button>
            <button class="text-red-600 hover:text-red-900 p-1 rounded hover:bg-red-50 transition-colors"
                    hx-delete="/companies/%s"
                    hx-target="#company-row-%s"
                    hx-swap="outerHTML"
                    hx-confirm="Are you sure you want to delete this company?"
                    title="Delete">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                </svg>
            </button>
        </td>
        </tr>`,
			company.ID,
			template.HTMLEscapeString(company.Name),
			company.ID,
			template.HTMLEscapeString(company.BankName),
			template.HTMLEscapeString(company.AccountNumber),
			getDocumentLinkWithPreview(company.AccountDocumentPath, "Account Document"),
			template.HTMLEscapeString(company.RegistrationNumber),
			getDocumentLinkWithPreview(company.RegistrationDocumentPath, "Registration Document"),
			getCreatedByDisplay(company.CreatedBy), // Created By column
			createdDate,                            // Created Date column
			company.ID,
			company.ID,
			company.ID)
	}
}

// Enhanced document link function with preview
func getDocumentLinkWithPreview(filename, documentType string) string {
	if filename == "" {
		return "<span class='text-gray-400'>None</span>"
	}
	return fmt.Sprintf(`<a href="javascript:void(0)" onclick="previewDocument('%s', '%s')" class="text-blue-600 hover:text-blue-800">View</a>`, filename, documentType)
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

	fmt.Println("=== DEBUG: Starting addCompany ===")

	// Parse multipart form for file uploads
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		fmt.Printf("DEBUG: ParseMultipartForm error: %v\n", err)
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}
	fmt.Println("DEBUG: Multipart form parsed successfully")

	// Get current logged-in user
	currentUser, err := getCurrentUser(r)
	if err != nil {
		fmt.Printf("DEBUG: Could not get current user: %v\n", err)
		currentUser = "unknown" // fallback
	}
	fmt.Printf("DEBUG: Current user: %s\n", currentUser)

	// Print all form values for debugging
	fmt.Println("DEBUG: Form values:")
	for key, values := range r.Form {
		fmt.Printf("  %s: %v\n", key, values)
	}

	// Generate company ID
	id, err := genID()
	if err != nil {
		fmt.Printf("DEBUG: genID error: %v\n", err)
		http.Error(w, "Failed to generate ID", http.StatusInternalServerError)
		return
	}
	fmt.Printf("DEBUG: Generated company ID: %s\n", id)

	// Handle file uploads
	fmt.Println("DEBUG: Handling file uploads...")
	accountDoc, err := handleFileUpload(r, "account_document")
	if err != nil {
		fmt.Printf("DEBUG: Account document upload error: %v\n", err)
		http.Error(w, "Failed to upload account document: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("DEBUG: Account document: %s\n", accountDoc)

	registrationDoc, err := handleFileUpload(r, "registration_document")
	if err != nil {
		fmt.Printf("DEBUG: Registration document upload error: %v\n", err)
		// Clean uploaded file if fails
		if accountDoc != "" {
			deleteUploadedFile(accountDoc)
		}
		http.Error(w, "Failed to upload registration document: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("DEBUG: Registration document: %s\n", registrationDoc)

	// Get form values
	name := r.FormValue("name")
	bankName := r.FormValue("bank_name")
	accountNumber := r.FormValue("account_number")
	registrationNumber := r.FormValue("registration_number")

	fmt.Printf("DEBUG: Form data - Name: '%s', Bank: '%s', Account: '%s', Reg: '%s'\n",
		name, bankName, accountNumber, registrationNumber)

	// Validate required fields
	if name == "" {
		fmt.Println("DEBUG: Company name is required")
		http.Error(w, "Company name is required", http.StatusBadRequest)
		return
	}

	company := &Company{
		ID:                       id,
		Name:                     name,
		BankName:                 bankName,
		AccountNumber:            accountNumber,
		AccountDocumentPath:      accountDoc,
		RegistrationNumber:       registrationNumber,
		RegistrationDocumentPath: registrationDoc,
		CreatedBy:                &currentUser, // Set the logged-in user
	}

	fmt.Printf("DEBUG: Attempting to create company: %+v\n", company)

	// Create company in database
	if err := db.CreateCompany(company); err != nil {
		fmt.Printf("DEBUG: CreateCompany error: %v\n", err)
		// Clean uploaded files if database operation fails
		if accountDoc != "" {
			deleteUploadedFile(accountDoc)
		}
		if registrationDoc != "" {
			deleteUploadedFile(registrationDoc)
		}
		http.Error(w, "Failed to create company: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("DEBUG: Company created successfully in database")

	w.Header().Set("Content-Type", "text/html")

	// Get the actual created timestamp from the database
	createdCompany, err := db.GetCompany(company.ID)
	var createdDate string
	if err != nil {
		fmt.Printf("DEBUG: Could not get created company for timestamp: %v\n", err)
		// Fallback to current time
		createdDate = "Just now"
	} else {
		// Format the timestamp to show date and time
		createdDate = formatTimestamp(createdCompany.CreatedAt)
		if createdDate == "" {
			createdDate = "Just now"
		}
	}

	fmt.Fprintf(w, `
    <tr id="company-row-%s">
        <td class="px-6 py-4 whitespace-nowrap">
            <div class="text-sm font-medium text-gray-900">%s</div>
            <div class="text-sm text-gray-500">ID: %s</div>
        </td>
        <td class="px-6 py-4">
            <div class="text-sm text-gray-900"><strong>Bank:</strong> %s</div>
            <div class="text-sm text-gray-500"><strong>Account:</strong> %s</div>
            <div class="text-sm text-gray-500">
                <strong>Document:</strong>
                %s
            </div>
        </td>
        <td class="px-6 py-4">
            <div class="text-sm text-gray-900"><strong>Number:</strong> %s</div>
            <div class="text-sm text-gray-500">
                <strong>Document:</strong>
                %s
            </div>
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            %s
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            %s
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
            <button class="text-blue-600 hover:text-blue-900 mr-3 p-1 rounded hover:bg-blue-50 transition-colors"
                    hx-get="/modal/edit-company/%s"
                    hx-target="#modal-container"
                    hx-swap="innerHTML"
                    title="Edit">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                </svg>
            </button>
            <button class="text-red-600 hover:text-red-900 p-1 rounded hover:bg-red-50 transition-colors"
                    hx-delete="/companies/%s"
                    hx-target="#company-row-%s"
                    hx-swap="outerHTML"
                    hx-confirm="Are you sure you want to delete this company?"
                    title="Delete">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                </svg>
            </button>
        </td>
    </tr>`,
		company.ID,                              // 1. %s - company.ID
		template.HTMLEscapeString(company.Name), // 2. %s - company.Name
		company.ID,                              // 3. %s - company.ID (for the ID display)
		template.HTMLEscapeString(company.BankName),                                           // 4. %s - bank name
		template.HTMLEscapeString(company.AccountNumber),                                      // 5. %s - account number
		getDocumentLinkWithPreview(company.AccountDocumentPath, "Account Document"),           // 6. %s - account doc
		template.HTMLEscapeString(company.RegistrationNumber),                                 // 7. %s - registration number
		getDocumentLinkWithPreview(company.RegistrationDocumentPath, "Registration Document"), // 8. %s - reg doc
		getCreatedByDisplay(company.CreatedBy),                                                // 9. %s - created by
		createdDate,                                                                           // 10. %s - created date with timestamp
		company.ID,                                                                            // 11. %s - company.ID (for edit button)
		company.ID,                                                                            // 12. %s - company.ID (for delete button)
		company.ID)                                                                            // 13. %s - company.ID (for delete target)

	fmt.Println("=== DEBUG: addCompany completed successfully ===")
}

func getCreatedByDisplay(createdBy *string) string {
	if createdBy == nil || *createdBy == "" {
		return "System"
	}
	return *createdBy
}

// formatTimestamp formats database timestamp to readable format
func formatTimestamp(timestamp string) string {
	if timestamp == "" {
		return ""
	}

	// If it's a full timestamp from database (e.g., "2024-01-15 14:30:25")
	if len(timestamp) > 16 {
		// Try to parse and format nicely
		// For SQLite, the format is usually "YYYY-MM-DD HH:MM:SS"
		if len(timestamp) >= 19 {
			return timestamp[:16] // Show "YYYY-MM-DD HH:MM"
		}
		return timestamp
	}

	return timestamp
}

// Check if email already exists
func checkEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Validate email format
	if !emailRegex.MatchString(email) {
		fmt.Fprintf(w, `<span class="text-red-500 text-xs">Invalid email format</span>`)
		return
	}

	// Check if email exists
	_, err := db.GetContactByEmail(email)
	if err == nil {
		fmt.Fprintf(w, `<span class="text-red-500 text-xs">Email already exists</span>`)
	} else {
		fmt.Fprintf(w, `<span class="text-green-500 text-xs">Email available</span>`)
	}
}

// Delete company handler
func deleteCompany(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	fmt.Println("DELETE company request received for id:", id)

	// Get company to find associated documents
	company, err := db.GetCompany(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Delete uploaded files
	if company.AccountDocumentPath != "" {
		deleteUploadedFile(company.AccountDocumentPath)
	}
	if company.RegistrationDocumentPath != "" {
		deleteUploadedFile(company.RegistrationDocumentPath)
	}

	// Delete company from database
	if err := db.DeleteCompany(id); err != nil {
		fmt.Println("Delete company error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return empty content - HTMX will remove element
	w.WriteHeader(http.StatusOK)
}

// Update company handler
func updateCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form for file uploads
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	// Get existing company
	company, err := db.GetCompany(id)
	if err != nil {
		http.Error(w, "Company not found", http.StatusNotFound)
		return
	}

	// Update basic fields
	company.Name = r.FormValue("name")
	company.BankName = r.FormValue("bank_name")
	company.AccountNumber = r.FormValue("account_number")
	company.RegistrationNumber = r.FormValue("registration_number")

	// Handle file uploads - only update if new files are provided
	if accountDoc, err := handleFileUpload(r, "account_document"); err == nil && accountDoc != "" {
		// Delete old account document
		if company.AccountDocumentPath != "" {
			deleteUploadedFile(company.AccountDocumentPath)
		}
		company.AccountDocumentPath = accountDoc
	}

	if registrationDoc, err := handleFileUpload(r, "registration_document"); err == nil && registrationDoc != "" {
		// Delete old registration document
		if company.RegistrationDocumentPath != "" {
			deleteUploadedFile(company.RegistrationDocumentPath)
		}
		company.RegistrationDocumentPath = registrationDoc
	}

	// Update company in database
	if err := db.UpdateCompany(company); err != nil {
		http.Error(w, "Failed to update company: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated table row
	w.Header().Set("Content-Type", "text/html")

	// Format created date
	createdDate := formatTimestamp(company.CreatedAt)
	if createdDate == "" {
		createdDate = "Unknown"
	}

	fmt.Fprintf(w, `
    <tr id="company-row-%s">
        <td class="px-6 py-4 whitespace-nowrap">
            <div class="text-sm font-medium text-gray-900">%s</div>
            <div class="text-sm text-gray-500">ID: %s</div>
        </td>
        <td class="px-6 py-4">
            <div class="text-sm text-gray-900"><strong>Bank:</strong> %s</div>
            <div class="text-sm text-gray-500"><strong>Account:</strong> %s</div>
            <div class="text-sm text-gray-500">
                <strong>Document:</strong>
                %s
            </div>
        </td>
        <td class="px-6 py-4">
            <div class="text-sm text-gray-900"><strong>Number:</strong> %s</div>
            <div class="text-sm text-gray-500">
                <strong>Document:</strong>
                %s
            </div>
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            %s
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
            <button class="text-blue-600 hover:text-blue-900 mr-3 p-1 rounded hover:bg-blue-50 transition-colors"
                    hx-get="/modal/edit-company/%s"
                    hx-target="#modal-container"
                    hx-swap="innerHTML"
                    title="Edit">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                </svg>
            </button>
            <button class="text-red-600 hover:text-red-900 p-1 rounded hover:bg-red-50 transition-colors"
                    hx-delete="/companies/%s"
                    hx-target="#company-row-%s"
                    hx-swap="outerHTML"
                    hx-confirm="Are you sure you want to delete this company?"
                    title="Delete">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                </svg>
            </button>
        </td>
    </tr>`,
		company.ID,
		template.HTMLEscapeString(company.Name),
		company.ID,
		template.HTMLEscapeString(company.BankName),
		template.HTMLEscapeString(company.AccountNumber),
		getDocumentLinkWithPreview(company.AccountDocumentPath, "Account Document"),
		template.HTMLEscapeString(company.RegistrationNumber),
		getDocumentLinkWithPreview(company.RegistrationDocumentPath, "Registration Document"),
		createdDate,
		company.ID,
		company.ID,
		company.ID)
}

// Helper function to display current document
func getCurrentDocumentDisplay(filename string) string {
	if filename == "" {
		return "No document uploaded"
	}
	return fmt.Sprintf(`<a href="/uploads/%s" target="_blank" class="text-blue-600 hover:text-blue-800">View current</a>`, filename)
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

	// SINGLE RENDER LOOP - FIXED
	for _, c := range contacts {
		isCurrentUser := isCurrentUserContact(c.Email, r)
		data := struct {
			Contact       Contact
			IsCurrentUser bool
		}{
			Contact:       c,
			IsCurrentUser: isCurrentUser,
		}
		if err := conCard.Execute(w, data); err != nil {
			fmt.Printf("Error rendering contact %s: %v\n", c.ID, err)
			continue
		}
	}
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

	// Check if email already exists
	existingContact, err := db.GetContactByEmail(email)
	if err == nil && existingContact != nil {
		fmt.Printf("Email already exists: %s\n", email)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `
			<div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full">
				<div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
					<div class="flex justify-end">
						<button hx-target="#contact-modal" hx-swap="outerHTML" hx-get="/modal/close" class="text-gray-400 hover:text-gray-600">&times;</button>
					</div>
					<h3 class="text-xl font-bold mb-4 text-red-600">Error</h3>
					<div class="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
						<p class="text-red-800">A contact with email <strong>%s</strong> already exists.</p>
					</div>
					<div class="flex justify-end">
						<button hx-target="#contact-modal" hx-swap="outerHTML" hx-get="/modal/close"
								class="bg-gray-500 text-white font-bold py-2 px-4 rounded-lg shadow-md hover:bg-gray-600 transition-colors duration-300">
							Close
						</button>
					</div>
				</div>
			</div>`, template.HTMLEscapeString(email))
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
		// Handle unique constraint error gracefully
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `
				<div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full">
					<div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
						<div class="flex justify-end">
							<button hx-target="#contact-modal" hx-swap="outerHTML" hx-get="/modal/close" class="text-gray-400 hover:text-gray-600">&times;</button>
						</div>
						<h3 class="text-xl font-bold mb-4 text-red-600">Error</h3>
						<div class="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
							<p class="text-red-800">A contact with email <strong>%s</strong> already exists.</p>
							<p class="text-red-600 text-sm mt-2">Please use a different email address.</p>
						</div>
						<div class="flex justify-end">
							<button hx-target="#contact-modal" hx-swap="outerHTML" hx-get="/modal/close"
									class="bg-gray-500 text-white font-bold py-2 px-4 rounded-lg shadow-md hover:bg-gray-600 transition-colors duration-300">
								Close
							</button>
						</div>
					</div>
				</div>`, template.HTMLEscapeString(email))
			return
		}
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
	renderCard(w, r, *newContact)
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
	companyID := r.FormValue("CompanyID")

	// Only update password if provided
	if password != "" {
		contact.Password = password
	}

	if companyID == "" {
		contact.CompanyID = nil
	} else {
		contact.CompanyID = &companyID
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
	renderCard(w, r, *contact)
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

	// SINGLE RENDER LOOP - FIXED
	for _, c := range results {
		isCurrentUser := isCurrentUserContact(c.Email, r)
		data := struct {
			Contact       Contact
			IsCurrentUser bool
		}{
			Contact:       c,
			IsCurrentUser: isCurrentUser,
		}
		if err := conCard.Execute(w, data); err != nil {
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

	// Convert CompanyID pointer to string for template
	companyIDStr := ""
	if contact.CompanyID != nil {
		companyIDStr = *contact.CompanyID
	}

	data := struct {
		Contact      *Contact
		Companies    []Company
		CompanyIDStr string
	}{
		Contact:      contact,
		Companies:    companies,
		CompanyIDStr: companyIDStr,
	}

	w.Header().Set("Content-Type", "text/html")

	// Use a simpler template without pointer comparison issues
	tmpl := template.Must(template.New("edit-modal").Parse(`
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
                        <option value="{{.ID}}" {{if eq .ID $.CompanyIDStr}}selected{{end}}>{{.Name}}</option>
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
    `))

	if err := tmpl.Execute(w, data); err != nil {
		fmt.Printf("Error executing edit modal template: %v\n", err)
		http.Error(w, "Failed to render modal", http.StatusInternalServerError)
		return
	}
}

// Edit company modal handler
func editCompanyModal(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	company, err := db.GetCompany(id)
	if err != nil {
		http.Error(w, "Company not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	// Create edit company modal HTML
	editCompanyModalHTML := `
    <div id="company-modal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full">
        <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
            <div class="flex justify-end">
                <button hx-target="#company-modal" hx-swap="outerHTML" hx-get="/modal/close" class="text-gray-400 hover:text-gray-600">&times;</button>
            </div>
            <h3 class="text-xl font-bold mb-4">Edit Company</h3>
            <form id="companyForm" enctype="multipart/form-data"
                  hx-put="/companies/` + company.ID + `"
                  hx-target="#company-row-` + company.ID + `"
                  hx-swap="outerHTML"
                  hx-on::after-request="if(event.detail.successful) htmx.remove(htmx.find('#company-modal'))">
                <div class="mb-4">
                    <label class="block text-gray-700 text-sm font-bold mb-2" for="companyName">Company Name</label>
                    <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                           id="companyName" name="name" type="text" value="` + template.HTMLEscapeString(company.Name) + `" required>
                </div>
                <div class="mb-4">
                    <label class="block text-gray-700 text-sm font-bold mb-2" for="bankName">Bank Name</label>
                    <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                           id="bankName" name="bank_name" type="text" value="` + template.HTMLEscapeString(company.BankName) + `">
                </div>
                <div class="mb-4">
                    <label class="block text-gray-700 text-sm font-bold mb-2" for="accountNumber">Account Number</label>
                    <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                           id="accountNumber" name="account_number" type="text" value="` + template.HTMLEscapeString(company.AccountNumber) + `">
                </div>
                <div class="mb-4">
                    <label class="block text-gray-700 text-sm font-bold mb-2" for="accountDocument">Account Document</label>
                    <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                           id="accountDocument" name="account_document" type="file" accept=".pdf,.jpg,.jpeg,.png">
                    <p class="text-xs text-gray-500 mt-1">Current: ` + getCurrentDocumentDisplay(company.AccountDocumentPath) + `</p>
                </div>
                <div class="mb-4">
                    <label class="block text-gray-700 text-sm font-bold mb-2" for="registrationNumber">Registration Number</label>
                    <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                           id="registrationNumber" name="registration_number" type="text" value="` + template.HTMLEscapeString(company.RegistrationNumber) + `">
                </div>
                <div class="mb-4">
                    <label class="block text-gray-700 text-sm font-bold mb-2" for="registrationDocument">Registration Document</label>
                    <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                           id="registrationDocument" name="registration_document" type="file" accept=".pdf,.jpg,.jpeg,.png">
                    <p class="text-xs text-gray-500 mt-1">Current: ` + getCurrentDocumentDisplay(company.RegistrationDocumentPath) + `</p>
                </div>
                <div class="flex items-center justify-end">
                    <button type="button" hx-target="#company-modal" hx-swap="outerHTML" hx-get="/modal/close"
                            class="bg-gray-500 text-white font-bold py-2 px-4 rounded-lg shadow-md hover:bg-gray-600 transition-colors duration-300 mr-2">Cancel</button>
                    <button type="submit" class="bg-blue-600 text-white font-bold py-2 px-4 rounded-lg shadow-md hover:bg-blue-700 transition-colors duration-300">Save Changes</button>
                </div>
            </form>
        </div>
    </div>`

	fmt.Fprint(w, editCompanyModalHTML)
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

		//Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: "authenticated",
			Path:  "/",
		})

		// Set current_user cookie for session tracking
		http.SetCookie(w, &http.Cookie{
			Name:  "current_user",
			Value: username,
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
	http.SetCookie(w, &http.Cookie{
		Name:   "current_user",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "password_change_user",
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

// server companies page
func companiesPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/companies.html")
}

// get companies for table view
func getCompaniesTable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	companies, err := db.GetAllCompanies()
	if err != nil {
		http.Error(w, "Failed to fetch companies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(companies) == 0 {
		w.Write([]byte(`<tr><td colspan="6" class="px-6 py-4 text-center text-gray-500">No companies found</td></tr>`))
		return
	}

	for _, company := range companies {
		// Format created date
		createdDate := formatTimestamp(company.CreatedAt)
		if createdDate == "" {
			createdDate = "Unknown"
		}

		fmt.Fprintf(w, `
        <tr id="company-row-%s">
        <td class="px-6 py-4 whitespace-nowrap">
            <div class="text-sm font-medium text-gray-900">%s</div>
            <div class="text-sm text-gray-500">ID: %s</div>
        </td>
        <td class="px-6 py-4">
            <div class="text-sm text-gray-900"><strong>Bank:</strong> %s</div>
            <div class="text-sm text-gray-500"><strong>Account:</strong> %s</div>
            <div class="text-sm text-gray-500">
                <strong>Document:</strong>
                %s
            </div>
        </td>
        <td class="px-6 py-4">
            <div class="text-sm text-gray-900"><strong>Number:</strong> %s</div>
            <div class="text-sm text-gray-500">
                <strong>Document:</strong>
                %s
            </div>
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            %s
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            %s
        </td>
        <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
            <button class="text-blue-600 hover:text-blue-900 mr-3 p-1 rounded hover:bg-blue-50 transition-colors"
                    hx-get="/modal/edit-company/%s"
                    hx-target="#modal-container"
                    hx-swap="innerHTML"
                    title="Edit">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                </svg>
            </button>
            <button class="text-red-600 hover:text-red-900 p-1 rounded hover:bg-red-50 transition-colors"
                    hx-delete="/companies/%s"
                    hx-target="#company-row-%s"
                    hx-swap="outerHTML"
                    hx-confirm="Are you sure you want to delete this company?"
                    title="Delete">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                </svg>
            </button>
        </td>
        </tr>`,
			company.ID,
			template.HTMLEscapeString(company.Name),
			company.ID,
			template.HTMLEscapeString(company.BankName),
			template.HTMLEscapeString(company.AccountNumber),
			getDocumentLinkWithPreview(company.AccountDocumentPath, "Account Document"),
			template.HTMLEscapeString(company.RegistrationNumber),
			getDocumentLinkWithPreview(company.RegistrationDocumentPath, "Registration Document"),
			getCreatedByDisplay(company.CreatedBy), // Created By column
			createdDate,                            // Created Date column
			company.ID,
			company.ID,
			company.ID)
	}
}

// get document link
func getDocumentLink(filename string) string {
	if filename == "" {
		return "<span class='text-gray-400'>None</span>"
	}
	return fmt.Sprintf(`<a href="/uploads/%s" target="_blank" class="text-blue-600 hover:text-blue-800">View</a>`, filename)
}

// PDF Handlers
func generateContactPDFHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	//Get contact
	contact, err := db.GetContact(id)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	//Get Company name if available
	companyName := ""
	if contact.CompanyID != nil {
		company, err := db.GetCompany(*contact.CompanyID)
		if err == nil {
			companyName = company.Name
		}
	}

	//GenPDF
	pdfService := NewPDFService()
	pdfBytes, err := pdfService.GenerateContactCardPDF(contact, companyName)
	if err != nil {
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	//Set response headers
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"contact_%s_%s_%s.pdf\"", contact.FirstName, contact.LastName, contact.ID))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	// Write PDF to response
	w.Write(pdfBytes)
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

	// Make sure the uploads directory exists
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		log.Printf("Warning: Could not create uploads directory: %v", err)
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

	// authRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// http.ServeFile(w, r, "static/index.html")
	// })

	authRouter.HandleFunc("/", indexHandler).Methods("GET")

	// Static file server
	authRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	authRouter.HandleFunc("/admin/license", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/license.html")
	}).Methods("GET")

	// Licensing
	authRouter.HandleFunc("/admin/license", licenseAdminHandler).Methods("GET")
	authRouter.HandleFunc("/admin/activate-license", activateLicenseHandler).Methods("GET", "POST")
	authRouter.HandleFunc("/admin/license-content", licenseContentHandler).Methods("GET")

	// Contact API endpoints
	authRouter.HandleFunc("/contacts", getContacts).Methods("GET")
	authRouter.HandleFunc("/contacts", addContact).Methods("POST")
	authRouter.HandleFunc("/contacts/{id}", updateContact).Methods("PUT", "PATCH")
	authRouter.HandleFunc("/contacts/{id}", deleteContact).Methods("DELETE")

	// Modal endpoints
	authRouter.HandleFunc("/modal/add", addModal).Methods("GET")
	authRouter.HandleFunc("/modal/edit/{id}", editModal).Methods("GET")
	authRouter.HandleFunc("/modal/close", closeForm).Methods("GET")

	// Companies page routes
	authRouter.HandleFunc("/companies-page", companiesPageHandler).Methods("GET")
	authRouter.HandleFunc("/companies-table", getCompaniesTable).Methods("GET")

	// Company search endpoint
	authRouter.HandleFunc("/search-companies", searchCompanies).Methods("GET")

	// Email validation endpoint
	authRouter.HandleFunc("/check-email", checkEmail).Methods("GET")

	// Company edit and delete routes
	authRouter.HandleFunc("/companies/{id}", deleteCompany).Methods("DELETE")
	authRouter.HandleFunc("/modal/edit-company/{id}", editCompanyModal).Methods("GET")
	authRouter.HandleFunc("/companies/{id}", updateCompany).Methods("PUT")

	// File uploads serving
	authRouter.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads", http.FileServer(http.Dir("./uploads"))))

	authRouter.HandleFunc("/modal/add-company", addCompanyModal).Methods("GET")
	authRouter.HandleFunc("/companies", addCompany).Methods("POST")
	authRouter.HandleFunc("/companies", getCompanies).Methods("GET")

	// Search endpoint
	authRouter.HandleFunc("/search", searchContacts).Methods("GET")

	//PDF CC
	authRouter.HandleFunc("/contacts/{id}/pdf", generateContactPDFHandler).Methods("GET")

	// Server start
	fmt.Println("AFcb started at http://localhost:1330")
	fmt.Println("Default admin login: af / afcb")
	log.Fatal(http.ListenAndServe(":1330", router))
}
