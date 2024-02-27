package communication

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/slack-go/slack"
	"github.com/vdparikh/veriflow/config"
	"github.com/vdparikh/veriflow/models"
)

type SlackService struct {
	Client   *slack.Client
	Messages config.Messages
	BaseURL  string
}

func (s *SlackService) getChannelID(userIDs ...string) string {
	channel, _, _, err := s.Client.OpenConversation(&slack.OpenConversationParameters{
		Users: userIDs,
	})

	if err != nil {
		log.Printf("Failed to open conversation: %v", err)
		return ""
	}
	return channel.ID
}

func (s *SlackService) GetUserInfo(userID string) (models.User, error) {
	user, err := s.Client.GetUserInfo(userID)
	if err != nil {
		return models.User{}, err
	}

	return models.User{
		ID:    user.ID,
		Name:  user.Profile.RealName,
		Email: user.Profile.Email,
		Image: user.Profile.Image192,
	}, nil
}

func (s *SlackService) SendVerificationMessage(request *models.VerifyRequest) error {
	text := fmt.Sprintf(s.Messages.VerificationMessage, request.Recipient.Name, request.Requestor.Name)

	channelID := s.getChannelID(request.Recipient.ID)
	if channelID == "" {
		return errors.New("failed to create channel")
	}

	authButtonTxt := slack.NewTextBlockObject("plain_text", "Authenticate", false, false)
	authButton := slack.NewButtonBlockElement("authenticate", request.ID, authButtonTxt).WithStyle(slack.StylePrimary).WithURL(request.AuthLink) // URL button for authentication

	reportButtonTxt := slack.NewTextBlockObject("plain_text", "Report Issue", false, false)
	reportButton := slack.NewButtonBlockElement("report_issue", request.ID, reportButtonTxt).WithStyle(slack.StyleDanger).WithURL(request.ReportLink) // URL button for reporting

	var err error
	request.Slack.Channel, request.Slack.MessageTS, err = s.sendMessage(channelID, request.Status, "Request for Verification", text, request.Requestor.Image, authButton, reportButton)

	return err
}

func (s *SlackService) SendInitConfirmation(request *models.VerifyRequest) error {
	text := fmt.Sprintf(s.Messages.RequestConfirmationMessage, request.Requestor.Name, request.Recipient.Name, time.Now().Format("2006-01-02 03:04 PM"))

	channelID := s.getChannelID(request.Requestor.ID)
	if channelID == "" {
		return errors.New("failed to create channel")
	}

	authButtonTxt := slack.NewTextBlockObject("plain_text", "Request Status", false, false)
	statusButton := slack.NewButtonBlockElement("authenticate", request.ID, authButtonTxt).WithStyle(slack.StylePrimary).WithURL(request.StatusLink) // URL button for authentication

	var err error
	request.Slack.InitChannel, request.Slack.InitMessageTS, err = s.sendMessage(channelID, request.Status, "Verification Request Initiated", text, request.Recipient.Image, statusButton)
	return err
}

// TODO: The message needs to go to the same CHANNEL where the message was initiated
func (s *SlackService) SendCompletionMessage(request *models.VerifyRequest) error {

	authButtonTxt := slack.NewTextBlockObject("plain_text", "Request Status", false, false)
	statusButton := slack.NewButtonBlockElement("authenticate", request.ID, authButtonTxt).WithStyle(slack.StylePrimary).WithURL(request.StatusLink) // URL button for authentication

	text := fmt.Sprintf(s.Messages.RequestorCompletionMessage, request.Requestor.Name, request.Recipient.Name, time.Now().Format("2006-01-02 03:04 PM"))
	channelID := s.getChannelID(request.Requestor.ID)
	s.sendMessage(channelID, request.Status, "Verification Completed", text, request.Recipient.Image, statusButton)

	text = fmt.Sprintf(s.Messages.RecipientCompletionMessage, request.Recipient.Name, request.Requestor.Name, time.Now().Format("2006-01-02 03:04 PM"))
	channelID = s.getChannelID(request.Recipient.ID)
	ch, ts, err := s.sendMessage(channelID, request.Status, "Thank you! Verification Completed", text, request.Requestor.Image, statusButton)

	// Decide if we want to send message to both parties induvidually or seperate?
	// text := fmt.Sprintf(s.Messages.RequestorCompletionMessage, request.Requestor.Name, request.Recipient.Name, time.Now().Format("2006-01-02 03:04 PM"))
	// channelID := s.getChannelID(request.Requestor.ID, request.Recipient.ID)
	// if channelID == "" {
	// 	return errors.New("failed to create channel")
	// }

	// ch, ts, err := s.sendMessage(channelID, request.Status, "Thank you! Verification Completed", text, "")

	s.Client.DeleteMessage(request.Slack.Channel, request.Slack.MessageTS)
	s.Client.DeleteMessage(request.Slack.InitChannel, request.Slack.InitMessageTS)

	request.Permalink, _ = s.Client.GetPermalink(&slack.PermalinkParameters{
		Channel: ch,
		Ts:      ts,
	})

	return err
}

func (s *SlackService) SendFailedVerificationMessage(request *models.VerifyRequest) error {
	text := fmt.Sprintf(s.Messages.RequestorVerificationFailureMessage, request.Requestor.Name, request.Recipient.Name, time.Now().Format("2006-01-02 03:04 PM"), request.Error)

	channelID := s.getChannelID(request.Requestor.ID)
	if channelID == "" {
		return errors.New("failed to create channel")
	}

	ch, ts, err := s.sendMessage(channelID, request.Status, "Verification Failed", text, request.Recipient.Image)

	text = fmt.Sprintf(s.Messages.RecipientVerificationFailureMessage, request.Recipient.Name, request.Requestor.Name, time.Now().Format("2006-01-02 03:04 PM"), request.Error)

	channelID = s.getChannelID(request.Recipient.ID)
	s.sendMessage(channelID, request.Status, "Verification Failed", text, request.Requestor.Image)

	s.Client.DeleteMessage(request.Slack.Channel, request.Slack.MessageTS)
	s.Client.DeleteMessage(request.Slack.InitChannel, request.Slack.InitMessageTS)

	request.Permalink, _ = s.Client.GetPermalink(&slack.PermalinkParameters{
		Channel: ch,
		Ts:      ts,
	})

	return err
}

func (s *SlackService) sendMessage(channelID, status, header, text, imageURL string, buttons ...*slack.ButtonBlockElement) (string, string, error) {
	headerText := slack.NewTextBlockObject("plain_text", header, false, false)
	headerSection := slack.NewHeaderBlock(headerText)

	sectionText := slack.NewTextBlockObject("mrkdwn", text, false, false)

	var sectionBlock *slack.SectionBlock
	if imageURL != "" {
		imageElement := slack.NewImageBlockElement(imageURL, imageURL)
		sectionBlock = slack.NewSectionBlock(sectionText, nil, slack.NewAccessory(imageElement))
	} else {
		sectionBlock = slack.NewSectionBlock(sectionText, nil, nil)
	}

	footerText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Powered by Veriflow* | %s", status), false, false)
	footerBlock := slack.NewContextBlock("", footerText)

	var msgOptions slack.MsgOption
	if len(buttons) > 0 {
		var blockElements []slack.BlockElement
		for _, btn := range buttons {
			blockElements = append(blockElements, btn)
		}
		actionBlock := slack.NewActionBlock("", blockElements...)
		msgOptions = slack.MsgOptionBlocks(headerSection, sectionBlock, actionBlock, footerBlock)
	} else {
		msgOptions = slack.MsgOptionBlocks(headerSection, sectionBlock, footerBlock)
	}

	channel, ts, err := s.Client.PostMessage(channelID, msgOptions)

	return channel, ts, err
}
