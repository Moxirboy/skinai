package sms

import (
	"math/rand"
	"net/smtp"
	"strconv"
	"time"
)

const smtpServer = "smtp.gmail.com"
const smtpPort = 587
const smtpUsername = "curifyapp1@gmail.com"
const smtpPassword = "wuyk nvub juvi ojzk"
const subject = "Curify Verification "
const body = "Your verification code is: "

func GenerateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	return strconv.Itoa(rand.Intn(9999))
}

func SendEmail(to, code string) error {
	from := smtpUsername
	password := smtpPassword

	auth := smtp.PlainAuth("", from, password, smtpServer)

	message := []byte("Subject: " + subject + "\r\n" +
		"\r\n" +
		body + code)

	err := smtp.SendMail(smtpServer+":"+strconv.Itoa(smtpPort), auth, from, []string{to}, message)
	if err != nil {
		return err
	}

	return nil
}
