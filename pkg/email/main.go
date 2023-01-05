package email

import (
	"bytes"
	"html/template"
	"log"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

// New
// Request struct
type Request struct {
	to      []string
	subject string
	body    string
}

func NewRequest(to []string, subject, body string) *Request {
	return &Request{
		to:      to,
		subject: subject,
		body:    body,
	}
}

func (r *Request) SendEmail() (bool, error) {

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

	auth := smtp.PlainAuth("", user, pass, host)

	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + r.subject + "!\n"
	// msg := []byte(subject + mime + "\n" + r.body)

	// Set the "Content-Type" header to "text/html".
	header := make(map[string]string)
	header["Content-Type"] = "text/html"

	msg := []byte("From: " + from + "\r\n" +
		"To: " + r.to[0] + "\r\n" +
		subject +
		// "Subject: " + subject + "\r\n" +
		mime +
		"\r\n" +
		r.body + "\r \n")

	// Add the headers to the message.
	for k, v := range header {
		msg = append([]byte(k+": "+v+"\r\n"), msg...)
	}

	addr := host + ":" + port

	if err := smtp.SendMail(addr, auth, from, r.to, msg); err != nil {
		return false, err
	}

	return true, nil
}

func (r *Request) ParseTemplate(templateFileName string, data interface{}) error {
	templateFileName = "pkg/email/templates/" + templateFileName

	t, err := template.ParseFiles(templateFileName)

	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	r.body = buf.String()

	return nil
}
