package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"github.com/vdparikh/veriflow/auth"
	"github.com/vdparikh/veriflow/communication"
	"github.com/vdparikh/veriflow/config"
	"github.com/vdparikh/veriflow/email"
	"github.com/vdparikh/veriflow/models"
)

type VerificationService struct {
	Config       *config.Config
	Auth         auth.Authenticator
	Communicator communication.Communicator
	EmailSvc     email.Email
}

func NewVerificationService(cfg *config.Config) (*VerificationService, error) {
	svc := &VerificationService{
		Config:       cfg,
		Auth:         auth.NewAuthenticator(cfg),
		Communicator: communication.NewCommunicator(cfg),
		EmailSvc:     email.NewEmail(cfg),
	}

	if svc.Communicator == nil {
		return svc, errors.New("no active service communication")
	}

	return svc, nil
}

func (svc *VerificationService) FetchUserDetails(request *models.VerifyRequest) error {
	//TODO: update the user details on every fetch to make sure things are latest and greatest
	var err error
	request.Requestor, err = svc.Communicator.GetUserInfo(request.Slack.UserID)
	if err != nil {
		log.Printf("failed to get user details: %v\n", err)
		return err
	}

	request.Requestor, err = models.GetUser(request.Slack.UserID)
	if err != nil {
		models.SaveUser(request.Requestor)
	}

	request.Recipient, err = svc.Communicator.GetUserInfo(request.Slack.RecipientID)
	if err != nil {
		log.Printf("failed to get recipient: %v\n", err)
		return errors.New("failed to get recipient")
	}

	request.Recipient, err = models.GetUser(request.Slack.RecipientID)
	if err != nil {
		request.Recipient.CredentialsJSON = ""
		request.Recipient.Credentials = []webauthn.Credential{}
		request.Recipient.Authenticator.Validated = false
		models.SaveUser(request.Recipient)
	}

	return nil
}

func (svc *VerificationService) InitVerification(request *models.VerifyRequest) error {
	err := svc.FetchUserDetails(request)
	if err != nil {
		return err
	}

	fmt.Printf("starting verification for %s initiated by %s\n", request.Requestor.Email, request.Recipient.Email)
	return svc.NewVerification(request)
}

func (svc *VerificationService) NewVerification(request *models.VerifyRequest) error {
	request.AuthLink, _ = svc.Auth.GenerateAuthLink("auth", request.ID)
	request.ReportLink, _ = svc.Auth.GenerateAuthLink("report", request.ID)
	request.StatusLink, _ = svc.Auth.GenerateAuthLink("status", request.ID)

	err := svc.Communicator.SendVerificationMessage(request)
	if err != nil {
		return err
	}

	request.Status = "SENT"
	err = svc.Communicator.SendInitConfirmation(request)
	if err != nil {
		return err
	}
	go svc.EmailSvc.SendVerificationMessage(request)
	return request.Save()
}

func (svc *VerificationService) SendFailure(request models.VerifyRequest) (*models.VerifyRequest, error) {
	request.DoneWithError(request.Error)
	svc.Communicator.SendFailedVerificationMessage(&request)
	go svc.EmailSvc.SendFailedVerificationMessage(&request)
	return &request, nil
}

func (svc *VerificationService) SendConfirmation(request models.VerifyRequest) (*models.VerifyRequest, error) {
	request.Done()
	svc.Communicator.SendCompletionMessage(&request)
	go svc.EmailSvc.SendCompletionMessage(&request)
	return &request, nil
}

func (svc *VerificationService) VerifyOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

func (svc *VerificationService) GenerateQRCode(secret, issuer, accountName string) ([]byte, error) {
	otpauthURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", issuer, accountName, secret, issuer)
	var png []byte
	png, err := qrcode.Encode(otpauthURL, qrcode.Medium, 256)
	return png, err
}

func (svc *VerificationService) EnableAuthenticator(user *models.User) error {
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "Veriflow", AccountName: user.ID})
	if err != nil {
		log.Fatalf("Error generating TOTP key: %v", err)
	}
	user.Authenticator = models.Authenticator{Secret: key.Secret(), FilePath: "", Permalink: "", Validated: false}
	return models.SaveUser(*user)
}

func (svc *VerificationService) ConfigureUser(userID string) (string, error) {
	existingUser, err := models.GetUser(userID)
	if err == nil {
		// User exists
		return svc.Auth.GenerateAuthLink("configure", existingUser.ID)
	}

	user, err := svc.Communicator.GetUserInfo(userID)
	if err != nil {
		return "", err
	}

	err = models.SaveUser(user)
	if err != nil {
		return "", err
	}

	return svc.Auth.GenerateAuthLink("configure", user.ID)

}
