package _pkg

type EmailPKG interface {
	SendEmail() (bool, error)
	ParseTemplate(templateFileName string, data interface{}) error
}
