package veriflow

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vdparikh/veriflow/models"
)

// OIDC Auth Callback
func (veriflow *Veriflow) AuthCallback(c *gin.Context) {
	// TODO: this needs to be changed. ID should be seperate than state
	state := c.Request.URL.Query().Get("state")
	stateBytes, _ := base64.StdEncoding.DecodeString(state)
	stateSlice := strings.Split(string(stateBytes), "=")

	if len(stateSlice) < 2 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid state"})
		return
	}

	fn, uuid := stateSlice[0], stateSlice[1]

	userInfo, oauth2Token, err := veriflow.Svc.Auth.GetAccessToken(c.Request.URL.Query().Get("code"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	redirectURL := "/user/configure"

	if fn == "login" {
		redirectURL = "/veriflow"
		models.GetOrCreateUser(userInfo.Email)
	} else if fn == "configure" {
		user, err := models.GetUser(uuid)
		if err != nil {
			user, err = veriflow.Svc.Communicator.GetUserInfo(uuid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user: " + err.Error()})
				return
			}
			models.SaveUser(user)
		}

		// Check if the user who autenticated is the same email address from Slack registration
		if user.Email != userInfo.Email {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Mismatch Emails"})
			return
		}
	} else {
		redirectURL = "/requests/" + uuid

		// Get request
		vr, err := models.GetByID(uuid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get request: " + err.Error()})
			return
		}

		// If user is neither requestor or recipient... Error out and don't create session
		if vr.Recipient.Email != userInfo.Email && vr.Requestor.Email != userInfo.Email {
			if vr.Status != "COMPLETED" && vr.Status != "FAILED" {
				vr.Error = fmt.Sprintf("Request Failed! The authenticated user email [%s] does not match requestor [%s]  or recipient [%s]", userInfo.Email, vr.Recipient.Email, userInfo.Email)
				veriflow.Svc.SendFailure(vr)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized Access"})
			return
		}

		// If REPORT and Status is not completed and clicked by Recipient
		if fn == "report" && vr.Status == "SENT" && userInfo.Email == vr.Recipient.Email {
			//TODO: Trigger security block
			vr.Error = "Reported by recipient. Please refrain from the converation and the incident has been reported to security"
			veriflow.Svc.SendFailure(vr)
		}

		// If Auth and Status is not completed and clicked by Recipient
		if fn == "auth" && vr.Status == "SENT" && userInfo.Email == vr.Recipient.Email {
			vr.SetOIDCInfo(userInfo)
			redirectURL = "/requests/" + uuid
		}
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    oauth2Token.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   int(time.Until(oauth2Token.Expiry).Seconds()),
	})

	c.Redirect(http.StatusFound, redirectURL)
}
