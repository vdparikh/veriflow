package communication

import (
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/vdparikh/veriflow/config"
	"github.com/vdparikh/veriflow/models"
)

// This allows Veriflow messages to various communication tools
// Right now it only allows Slack and Email

type Communicator interface {
	GetUserInfo(userID string) (models.User, error)
	SendInitConfirmation(request *models.VerifyRequest) error
	SendVerificationMessage(request *models.VerifyRequest) error
	SendCompletionMessage(request *models.VerifyRequest) error
	SendFailedVerificationMessage(request *models.VerifyRequest) error
}

func NewCommunicator(cfg *config.Config) Communicator {
	switch cfg.Communication.ActiveService {
	case "slack":
		client := slack.New(
			cfg.Communication.Services.Slack.BotToken,
			slack.OptionDebug(false),
			slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
			slack.OptionAppLevelToken(cfg.Communication.Services.Slack.AppToken),
		)

		return &SlackService{
			Client:   client,
			Messages: cfg.Messages,
			BaseURL:  cfg.BaseURL,
		}
	case "ms_teams":
		return nil
	default:
		return nil
	}

}
