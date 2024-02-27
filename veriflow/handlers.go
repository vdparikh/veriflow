package veriflow

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/vdparikh/veriflow/models"
)

// Handle Verification Subcommands
func (veriflow *Veriflow) HandleVeriflow(c *gin.Context) {
	userID, responseURL, commandText := c.PostForm("user_id"), c.PostForm("response_url"), c.PostForm("text")
	args := strings.Fields(commandText)

	switch {
	case len(args) == 0 || args[0] == "help":
		veriflow.HandleHelp(c)
	case args[0] == "verify" && len(args) >= 2:
		veriflow.HandleVerify(c, "slack", userID, extractUserID(args[1]), responseURL, commandText)
	case args[0] == "configure":
		veriflow.HandleConfigure(c, userID)
	default:
		c.JSON(http.StatusOK, gin.H{"response_type": "ephemeral", "text": "Unknown subcommand. Available subcommands are `verify` and `configure`."})
	}
}

// Help
func (veriflow *Veriflow) HandleHelp(c *gin.Context) {
	helpText := "Here are some things you can do with /veriflow command:\n" +
		"• `/veriflow verify <user>` - Start a Verification process for a user.\n" +
		"• `/veriflow configure` - Configure your Veriflow Setings.\n" +
		"• `/veriflow help` - Get help on how to use the Veriflow slash command."

	c.JSON(http.StatusOK, gin.H{"response_type": "ephemeral", "text": helpText})
}

// Handle Verification Request
func (veriflow *Veriflow) HandleVerify(c *gin.Context, service, requestor, recipient, responseURL, commandText string) {
	// The slackrequest model should be just Request which stores details of incoming request
	// Generic across teams, slack, webex, zoom etc.
	request := models.VerifyRequest{
		Start:             time.Now(),
		ID:                uuid.New().String(),
		Status:            "RECEIVED",
		CommunicationTool: service,
		Message:           commandText,
		Slack:             models.SlackRequest{UserID: requestor, RecipientID: recipient, Text: commandText, ResponseURL: responseURL},
	}

	err := veriflow.Svc.InitVerification(&request)

	if err != nil {
		request.DoneWithError(err.Error())
		switch service {
		case "slack":
			c.JSON(http.StatusOK, gin.H{"response_type": "ephemeral", "text": fmt.Sprintf("[API] Verification Failed with error %s", err.Error())})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// request.Status = "QUEUED"
	// veriflow.VerificationQueue <- request
	// veriflow.Svc.NewVerification(&request)

	c.JSON(http.StatusOK, gin.H{"response_type": "ephemeral", "text": "[API] Verification Initiated. Please head on to the Veriflow app to monitor the progress of your verification."})
}

// Configure authenticator
func (veriflow *Veriflow) HandleConfigure(c *gin.Context, userID string) {
	authLink, err := veriflow.Svc.ConfigureUser(userID)

	if err != nil {
		veriflow.Logger.Errorf("Error configuring user [%s] %s\n", userID, err.Error())
		c.JSON(http.StatusOK, gin.H{
			"response_type": "ephemeral",
			"text":          err.Error(),
		})
		return
	}

	// authLink, _ := veriflow.Svc.Auth.GenerateConfigLink(userID)

	buttonText := slack.NewTextBlockObject("plain_text", "Configure Veriflow", false, false)
	buttonElement := slack.NewButtonBlockElement(userID, "btn_configure", buttonText).WithStyle(slack.StylePrimary).WithURL(authLink)

	actionBlock := slack.NewActionBlock("", buttonElement)
	blocks := slack.Blocks{BlockSet: []slack.Block{actionBlock}}

	c.JSON(http.StatusOK, gin.H{
		"response_type": "ephemeral",
		"blocks":        blocks,
	})
}

func (veriflow *Veriflow) GetVerifyCodeHandler(c *gin.Context) {

	uuid := c.Param("uuid")
	vr, _ := models.GetByID(uuid)

	value, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find custom data"})
		return
	}

	sessionUser, ok := value.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid type"})
		return
	}

	if vr.Recipient.Email != sessionUser.Email {
		c.Data(http.StatusUnauthorized, "text/html", []byte("Invalid Request"))
		return
	}
	data := gin.H{
		"UUID":    uuid,
		"Request": vr,
		"User":    sessionUser,
	}

	c.HTML(http.StatusOK, "authenticator.html", data)
}

func (veriflow *Veriflow) PostVerifyCodeHandler(c *gin.Context) {
	uuid := c.Param("uuid")
	vr, _ := models.GetByID(uuid)

	value, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find custom data"})
		return
	}

	sessionUser, ok := value.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid type"})
		return
	}

	if vr.Recipient.Email != sessionUser.Email {
		c.Data(http.StatusUnauthorized, "text/html", []byte("Invalid Request"))
		return
	}

	code := c.PostForm("code")
	valid := veriflow.Svc.VerifyOTP(sessionUser.Authenticator.Secret, code)
	if valid {
		veriflow.Svc.SendConfirmation(vr)
		c.Redirect(http.StatusFound, "/requests/"+uuid)
		return
	}

	data := gin.H{
		"UUID":  uuid,
		"User":  sessionUser,
		"Error": "Invalid Code",
	}

	c.HTML(http.StatusOK, "authenticator.html", data)
}
