package smtp

import (
	"crypto/tls"
	"github.com/jordan-wright/email"
	"github.com/revel/revel"
	"net/smtp"
	"os"
)

type smtp_type struct {
	host     string
	port     string
	username string
	password string
	from     string
}

var smtp_data smtp_type = smtp_type{}

func InitSMTP() {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	address := smtpHost + ":" + smtpPort
	tlsConfig := &tls.Config{ServerName: smtpHost}
	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return
	}
	defer conn.Close()
	revel.AppLog.Info("Connected to SMTP server")
	smtp_data.host = os.Getenv("SMTP_HOST")
	smtp_data.port = os.Getenv("SMTP_PORT")
	smtp_data.username = os.Getenv("SMTP_USER")
	smtp_data.password = os.Getenv("SMTP_PASS")
	smtp_data.from = smtp_data.username
}

func SendMessage(email_to string, subject string, message string) error {
	e := email.NewEmail()
	e.From = smtp_data.from
	e.To = []string{email_to}
	e.Subject = subject
	e.Text = []byte(message)
	return e.SendWithTLS(
		smtp_data.host+":"+smtp_data.port,
		smtp.PlainAuth("", smtp_data.username, smtp_data.password, smtp_data.host),
		&tls.Config{ServerName: smtp_data.host},
	)
}
