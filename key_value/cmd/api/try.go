package main

import (
	"log"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func GenerateKey() error {
	KeyLength := 7
	t := time.Now().String()
	KeyHash, err := bcrypt.GenerateFromPassword([]byte(t), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// Make a Regex we only want letters and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return nil
	}
	processedString := reg.ReplaceAllString(string(KeyHash), "")
	log.Println(string(processedString))

	KeyNumber := strings.ToUpper(string(processedString[len(processedString)-KeyLength:]))
	log.Println(KeyNumber)
	return nil
}
