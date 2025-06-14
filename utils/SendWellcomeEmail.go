package utils

import (
	"fmt"
	"net/smtp"
)

func SendWellcomeEmail(name, toEmail string) error {
	from := "mdsazibhossin2021@gmail.com"
	password := "csav ngvr llur pmfr"

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message.
	// message := []byte("Subject: Welcome to Our Platform!\r\n\r\nWelcome! Thanks for registering.")

	message := []byte("Subject: 🎉 Welcome to Our Platform! 🎉\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		`
	<!DOCTYPE html>
	<html>
	<head>
	  <meta charset="UTF-8">
	  <style>
		body {
		  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
		  background-color: #f9f9f9;
		  color: #333;
		  padding: 20px;
		}
		.container {
		  max-width: 600px;
		  margin: auto;
		  background-color: #ffffff;
		  padding: 30px;
		  border-radius: 8px;
		  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
		}
		.header {
		  text-align: center;
		  margin-bottom: 30px;
		}
		.header h1 {
		  color: #4CAF50;
		}
		.button {
		  display: inline-block;
		  padding: 12px 25px;
		  margin-top: 20px;
		  background-color: #4CAF50;
		  color: white;
		  text-decoration: none;
		  border-radius: 5px;
		}
		.footer {
		  margin-top: 30px;
		  font-size: 13px;
		  color: #888;
		  text-align: center;
		}
	  </style>
	</head>
	<body>
	  <div class="container">
		<div class="header">
		  <h1>Welcome Aboard! e-com 🚀</h1>
		  <p>Thank you for joining us, <strong>` + name + ` Friend!</strong></p>
		</div>
		<p>We’re thrilled to have you here. Your journey starts now and we’ll be with you every step of the way.</p>
		<p>If you have any questions or need help, just reply to this email — we’re always happy to help.</p>
		<a class="button" href="https://your-website.com">Get Started</a>
		<div class="footer">
		  <p>— The Team</p>
		  <p>Need help? <a href="mailto:mdsazibhossin2021@gmail.com">Contact Support</a></p>
		</div>
	  </div>
	</body>
	</html>
	`)

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
	if err != nil {
		return err
	}

	fmt.Println("Welcome email sent to:", toEmail)
	return nil
}
