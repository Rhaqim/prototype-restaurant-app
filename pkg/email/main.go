package email

import (
	"context"
	"log"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

func SendEmail(ctx context.Context, to []string, subject string, body string) error {

	if ctx.Err() == context.Canceled {
		return ctx.Err()
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file fro email")
	}

	var (
		EmailHost     = os.Getenv("EMAIL_HOST")
		EmailPort     = os.Getenv("EMAIL_PORT")
		EmailUser     = os.Getenv("EMAIL_USER")
		EmailPassword = os.Getenv("EMAIL_PASSWORD")
		EmailFrom     = os.Getenv("EMAIL_FROM")
	)

	// Set up authentication information.
	host, port, user, pass, from := EmailHost, EmailPort, EmailUser, EmailPassword, EmailFrom

	log.Printf("Credentials \n from: %s, user: %s, host: %s, pass: %s, port: %s", from, user, host, pass, port)

	auth := smtp.PlainAuth("", user, pass, host)

	// Set the "Content-Type" header to "text/html".
	header := make(map[string]string)
	header["Content-Type"] = "text/html"

	// Set the HTML and CSS for the email template.
	template := `
		<html>
			<head>
				<style>
					h1 {
						color: blue;
					}
				</style>
			</head>
			<body>
				<h1>` + body + `</h1>
			</body>
		</html>
	`

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	msg := []byte("From: " + from + "\r\n" +
		"To: " + to[0] + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		template)

	// Add the headers to the message.
	for k, v := range header {
		msg = append([]byte(k+": "+v+"\r\n"), msg...)
	}

	// err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)
	err = smtp.SendMail(host+":"+port, auth, from, to, msg)
	if err != nil {
		return err
	}

	return nil
}
