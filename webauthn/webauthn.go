package webauthn

import (
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/vdparikh/veriflow/models"
)

type WebAuthnService struct {
	WebAuthn  *webauthn.WebAuthn
	UserStore models.UserStore
}

func NewWebAuthnService(webAuthnConfig *webauthn.Config, userStore models.UserStore) (*WebAuthnService, error) {
	webAuthn, err := webauthn.New(webAuthnConfig)
	if err != nil {
		return nil, err
	}
	return &WebAuthnService{
		WebAuthn:  webAuthn,
		UserStore: userStore,
	}, nil
}

func (s *WebAuthnService) BeginRegistration(userID string) (interface{}, error) {
	user, err := s.UserStore.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	options, sessionData, err := s.WebAuthn.BeginRegistration(
		user,
	)
	if err != nil {
		return nil, err
	}

	s.UserStore.SaveSessionData(userID, sessionData)
	return options, nil
}

func (s *WebAuthnService) FinishRegistration(userID string, response *http.Request) error {
	user, err := s.UserStore.GetUserByID(userID)
	if err != nil {
		return err
	}
	sessionData, err := s.UserStore.GetSessionData(userID)
	if err != nil {
		return err
	}

	parsedCredential, err := s.WebAuthn.FinishRegistration(user, *sessionData, response)
	if err != nil {
		return err
	}

	s.UserStore.SaveCredentials(userID, parsedCredential)
	s.UserStore.DeleteSessionData(userID)

	return nil
}

func (s *WebAuthnService) BeginLogin(userID string) (interface{}, error) {
	user, err := s.UserStore.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	credentials := s.UserStore.GetCredentials(userID)
	allowCredentials := make([]protocol.CredentialDescriptor, len(credentials))
	for i, cred := range credentials {
		allowCredentials[i] = protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.ID,
		}
	}

	options, sessionData, err := s.WebAuthn.BeginLogin(user)
	if err != nil {
		return nil, err
	}

	s.UserStore.SaveSessionData(userID, sessionData)
	return options, nil
}

func (s *WebAuthnService) FinishLogin(userID string, response *http.Request) error {
	user, err := s.UserStore.GetUserByID(userID)
	if err != nil {
		return err
	}

	sessionData, _ := s.UserStore.GetSessionData(userID)
	_, err = s.WebAuthn.FinishLogin(user, *sessionData, response)
	if err != nil {
		return err
	}

	s.UserStore.DeleteSessionData(userID)
	return nil
}
