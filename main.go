package main

import (
	"fmt"
	"io"

	"gopkg.in/gomail.v2"
)

type Contact struct {
	FirstName   string
	LastName    string
	Email       string
	CountryCode string
	Phone       string
}

func main() {

	ok := Contact{
		FirstName:   "Orm Kornnaphat",
		LastName:    "Sethratanapong",
		Email:       "ok@kshhh.co",
		CountryCode: "",
		Phone:       "",
	}

	mail := gomail.NewMessage()
	mail.SetHeader("From", "af@drofylla.com")
	mail.SetHeader("To", ok.Email)
	mail.SetHeader("Subject", "AFcb Registration")
	mail.SetBody("text/plain", fmt.Sprintf("Hi %s! Your contact details has been saved in AFcb.", ok.FirstName))

	send := gomail.SendFunc(func(from string, to []string, msg io.WriterTo) error {
		fmt.Println("From:", from)
		fmt.Println("To:", to)
		fmt.Println("Subject:", mail.GetHeader("Subject"))

		return nil
	})

	if err := gomail.Send(send, mail); err != nil {
		panic(err)
	}
}
