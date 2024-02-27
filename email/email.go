package email

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/vdparikh/veriflow/config"
	"github.com/vdparikh/veriflow/models"

	"github.com/gomarkdown/markdown/parser"
)

type Email interface {
	SendInitConfirmation(request *models.VerifyRequest) error
	SendVerificationMessage(request *models.VerifyRequest) error
	SendCompletionMessage(request *models.VerifyRequest) error
	SendFailedVerificationMessage(request *models.VerifyRequest) error
}

func NewEmail(cfg *config.Config) Email {

	return &EmailService{
		SMTPServer: cfg.Email.SMTPServer,
		Port:       cfg.Email.Port,
		Username:   cfg.Email.Username,
		Password:   cfg.Email.Password,
		From:       cfg.Email.From,
		Messages:   cfg.Messages,
		Enabled:    cfg.Email.Enabled,
	}

}

type EmailService struct {
	SMTPServer string
	Port       int
	Username   string
	Password   string
	From       string
	Messages   config.Messages
	Enabled    bool
}

func mdToHTML(md string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return string(markdown.Render(doc, renderer))
}

func (s *EmailService) sendEmail(from string, to []string, subject string, body string) error {
	auth := smtp.PlainAuth("", s.Username, s.Password, s.SMTPServer)

	html := mdToHTML(body)

	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Avenir, 'Segoe UI', Arial, sans-serif; background: #f2f2f2; }
				.box { border: 2px solid #f2f2f2;  }
				.header { background: #f2f2f2; padding: 10px; text-align: center; }
				.header img { width: 200px; }
				.content { padding: 20px; background: #fff;}
				.footer { background: #f2f2f2; padding: 10px; text-align: center; }
				.buttons { margin: 20px 0px; }
				a.btn { border: 1px solid #fff; border-radius: 0.5rem; padding: 5px 50px; background: #eee; text-decoration: none; color: #fff; font-size: 16px; font-weight: bold; margin-right: 10px; }
				a.btn.auth { background: green; }
				a.btn.report { background: crimson; }
			</style>
		</head>
		<body>
		<div class="box">
			<div class="header">
				<p><strong>Veriflow Verification</strong></p>
			</div>
			<div class="content">
				%s
			</div>
			<div class="footer">
				<!-- Footer content here -->
			</div>
		</div>
		</body>
		</html>
	`, html)

	msg := []byte("MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"From: " + from + "\r\n" +
		"To: " + strings.Join(to, ", ") + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		htmlBody)

	err := smtp.SendMail("smtp.gmail.com:587", auth, s.From, to, msg)
	return err
}

func (s *EmailService) SendVerificationMessage(request *models.VerifyRequest) error {
	if !s.Enabled {
		return nil
	}

	to := []string{
		request.Recipient.Email,
	}

	subject := "Verification Request"
	body := fmt.Sprintf(s.Messages.VerificationMessage, request.Recipient.Name, request.Requestor.Name)
	body = body + fmt.Sprintf("<div class='buttons'><a class='btn auth' href='%s'>Verify</a> <a class='btn report' href='%s'>Report</a></div>", request.AuthLink, request.ReportLink)
	return s.sendEmail(s.From, to, subject, body)
}

func (s *EmailService) SendInitConfirmation(request *models.VerifyRequest) error {
	if !s.Enabled {
		return nil
	}

	to := []string{
		request.Requestor.Email,
	}
	subject := "Verification Request Initiated"
	body := fmt.Sprintf(s.Messages.RequestConfirmationMessage, request.Requestor.Name, request.Recipient.Name, time.Now().Format("2006-01-02 03:04 PM"))

	return s.sendEmail(s.From, to, subject, body)
}

func (s *EmailService) SendCompletionMessage(request *models.VerifyRequest) error {
	if !s.Enabled {
		return nil
	}

	to := []string{
		request.Recipient.Email,
		request.Requestor.Email,
	}
	subject := "Verification Completed"
	body := fmt.Sprintf(s.Messages.RequestorCompletionMessage, request.Requestor.Name, request.Recipient.Name, time.Now().Format("2006-01-02 03:04 PM"))

	return s.sendEmail(s.From, to, subject, body)
}

func (s *EmailService) SendFailedVerificationMessage(request *models.VerifyRequest) error {
	if !s.Enabled {
		return nil
	}

	to := []string{
		request.Requestor.Email,
	}
	subject := "Verification Failed"
	body := fmt.Sprintf(s.Messages.RequestorVerificationFailureMessage, request.Requestor.Name, request.Recipient.Name, time.Now().Format("2006-01-02 03:04 PM"))

	return s.sendEmail(s.From, to, subject, body)
}
