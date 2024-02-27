package veriflow

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vdparikh/veriflow/models"
	"golang.org/x/oauth2"
)

func (veriflow *Veriflow) AuthTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("auth_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userEmail, isValid := veriflow.ValidateAccessToken(accessToken)
		if !isValid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		user, err := models.GetUserByEmail(userEmail)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Set("user", user)
		c.Set("userEmail", user.Email)
		c.Next()
	}
}

func (veriflow *Veriflow) ValidateAccessToken(accessToken string) (string, bool) {
	provider, _, _ := veriflow.Svc.Auth.GetConfig()
	ctx := context.Background()
	userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	}))

	if err != nil {
		return "", false
	}
	return userInfo.Email, true
}
