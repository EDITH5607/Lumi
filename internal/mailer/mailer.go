package mailer

import (
	"bytes"
	"embed"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// a new variable with the type embed.FS (embedded file system) to hold the templates
// we use 'go:embed "dir name"

//go:embed "templates"
var templateFS embed.FS


// dialer means the email provider  like amazon ses, and sender is whom u want to send
type Mailer struct {
	dialer *mail.Dialer
	sender  string
}

// initialize an dialer and return instance of Mailer ... provide for the main.go
func New(host string, port int, username, password, sender string) Mailer {
	// initialize newDailer with port host username password
	dialer := mail.NewDialer(host,port,username,password)
	// provide timeout of 5 seconds (which wait for 5 seconds for the response of smtp server)
	dialer.Timeout = 5 * time.Second

	// return the instance of the Mailer 
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// reusable send fn which recive recipient template file and data as parameter
func (m Mailer) Send(recipient, templateFile string, data any) error {

	// this will find and load template file and make a new template 
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/" + templateFile)
	if err!=nil {
		return err
	}

	// ExecuteTemplate takes the template block named "subject", fills in the dynamic data, and writes the result into the buffer. 
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject,"subject", data)
	if err!=nil {
		return err
	}

	// ExecuteTemplate takes the template block named "plainBody", fills in the dynamic data, and writes the result into the buffer. 
	plainBody := new(bytes.Buffer)
	err  = tmpl.ExecuteTemplate(plainBody, "plainBody",data)
	if err != nil {
		return err
	}

	// ExecuteTemplate takes the template block named "htmlbody", fills in the dynamic data, and writes the result into the buffer. 
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlbody", data)
	if err!=nil {
		return err
	}

	// sending message with specify the format with setting the header, body and alternative for new systems 
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject",subject.String())

	msg.SetBody("text/plain",plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())


	// retry sending upto three times before aborting and returning the final error, sleep 500milisecond before sending next
	for i:=1; i <= 3; i++ {
		// after setting all call the dialand send to the smtp provider
		err = m.dialer.DialAndSend(msg)
		// for visual not mistaking that why we use nil == err , both ways are correct
		if nil == err {
			return nil
		}
		// if didnt work sleep for 500 milliseconds and retry
		time.Sleep(500 * time.Millisecond )
	}

	return err
}