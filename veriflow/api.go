package veriflow

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/vdparikh/veriflow/models"
)

// LogrusLogger is the logging middleware using Logrus
func LogrusLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		statusCode := c.Writer.Status()
		log.WithFields(log.Fields{
			"status_code": statusCode,
			"latency":     latency,
			"client_ip":   c.ClientIP(),
			"method":      c.Request.Method,
			"path":        c.Request.RequestURI,
			"user_agent":  c.Request.UserAgent(),
			"errors":      c.Errors.ByType(gin.ErrorTypePrivate).String(),
		}).Info("Request completed")
	}
}

func (veriflow *Veriflow) SetupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Use(LogrusLogger())

	router.LoadHTMLGlob("templates/*")

	router.Any("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Dummy endpoint for Slack interactivity.
	// If removed it throws 404 when users clicks on any button
	router.POST("/veriflow", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	router.GET("/login", func(c *gin.Context) {
		authLink, _ := veriflow.Svc.Auth.GenerateAuthLink("login", "")
		c.Redirect(302, authLink)
		// c.JSON(http.StatusOK, gin.H{})
	})

	// Web pages behind Aht
	app := router.Group("")
	app.Use(veriflow.AuthTokenMiddleware())

	app.GET("/veriflow", func(c *gin.Context) {
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
		data := gin.H{
			"User":   sessionUser,
			"Config": veriflow.Cfg,
		}
		c.HTML(http.StatusOK, "home.html", data)
	})

	app.GET("/requests", func(c *gin.Context) {
		requestsOutgoing, _ := models.GetAllByUser(c.GetString("userEmail"))
		requestsIncoming, _ := models.GetAllByRecipientEmail(c.GetString("userEmail"))

		data := gin.H{
			"Sent":     requestsOutgoing,
			"Recieved": requestsIncoming,
		}

		c.HTML(http.StatusOK, "requests.html", data)
	})

	app.GET("/requests/:uuid", func(c *gin.Context) {
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

		uuid := c.Param("uuid")
		vr, err := models.GetByID(uuid)
		if err != nil {
			c.Data(http.StatusUnauthorized, "text/html", []byte("Invalid Request"))
			return
		}
		userEmail := c.GetString("userEmail")
		if vr.Requestor.Email != userEmail && vr.Recipient.Email != userEmail {
			c.Data(http.StatusUnauthorized, "text/html", []byte("Invalid Request"))
			return
		}

		data := gin.H{
			"User":    sessionUser,
			"Request": vr,
		}

		c.HTML(http.StatusOK, "request.html", data)
	})

	app.GET("/requests/:uuid/report", func(c *gin.Context) {
		uuid := c.Param("uuid")
		vr, err := models.GetByID(uuid)
		if err != nil {
			c.Data(http.StatusUnauthorized, "text/html", []byte("Invalid Request"))
			return
		}
		userEmail := c.GetString("userEmail")
		if vr.Requestor.Email != userEmail && vr.Recipient.Email != userEmail {
			c.Data(http.StatusUnauthorized, "text/html", []byte("Invalid Request"))
			return
		}

		vr.Error = "Incident reported to security. Please avoid communication with the user"
		veriflow.Svc.SendFailure(vr)

		redirectURL := "/requests/" + uuid
		c.Redirect(http.StatusFound, redirectURL)
	})

	app.GET("/requests/:uuid/approve", func(c *gin.Context) {
		uuid := c.Param("uuid")
		vr, err := models.GetByID(uuid)
		if err != nil {
			c.Data(http.StatusUnauthorized, "text/html", []byte("Invalid Request"))
			return
		}
		userEmail := c.GetString("userEmail")

		if vr.Recipient.Email != userEmail {
			if vr.Status != "COMPLETED" && vr.Status != "FAILED" {
				vr.Error = fmt.Sprintf("Request Failed! The authenticated user email [%s] does not match requestor [%s]  or recipient [%s]", userEmail, vr.Recipient.Email, userEmail)
				veriflow.Svc.SendFailure(vr)
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized Access"})
			return
		}

		redirectURL := "/requests/" + uuid
		if veriflow.Svc.Config.Authenticator.Enabled {
			vr.Save()
			redirectURL = "/requests/" + uuid + "/verify-code"
		} else {
			veriflow.Svc.SendConfirmation(vr)
		}

		c.Redirect(http.StatusFound, redirectURL)
	})

	// Authenticator endpoints for verification and QR code generation
	app.GET("/requests/:uuid/verify-code", veriflow.GetVerifyCodeHandler)
	app.POST("/requests/:uuid/verify-code", veriflow.PostVerifyCodeHandler)

	// Configuration page for users
	app.GET("/user/configure", func(c *gin.Context) {
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

		data := gin.H{"User": sessionUser}
		c.HTML(http.StatusOK, "configure-authenticator.html", data)
	})

	// Veriflow Primary Endpoints
	// Support help, verify and configure subcommands
	// Should we have enroll which combines configure and will store user information in DB?
	slack := router.Group("/slack")
	slack.POST("/veriflow", veriflow.HandleVeriflow)

	// OIDC Callback
	auth := router.Group("/auth")
	auth.GET("/callback", veriflow.AuthCallback)

	// API endpoints
	api := router.Group("/api")
	api.Use(veriflow.AuthTokenMiddleware())

	api.POST("/veriflow", func(c *gin.Context) {
		var requestPayload struct {
			RecipientID string `json:"recipient_id"`
		}

		if err := c.ShouldBindJSON(&requestPayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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

		// This is temporary to make sure all users are in the system
		// TODO: this seems redundant now
		models.GetOrCreateUser(requestPayload.RecipientID)

		veriflow.HandleVerify(c, "api", sessionUser.Email, requestPayload.RecipientID, "", "")
	})

	// Enable authenticator support
	api.POST("/authenticator", func(c *gin.Context) {
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

		if sessionUser.Authenticator.Secret == "" {
			veriflow.Svc.EnableAuthenticator(&sessionUser)
		}

		decodedData, err := veriflow.Svc.GenerateQRCode(sessionUser.Authenticator.Secret, "Veriflow", sessionUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate QR Code"})
			return
		}

		image := "data:image/png;base64," + base64.StdEncoding.EncodeToString(decodedData)

		c.JSON(http.StatusOK, gin.H{
			"encoded_qr_code": image,
		})
	})

	api.POST("/authenticator/validate", func(c *gin.Context) {
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

		var requestPayload struct {
			Code string `json:"code"`
		}

		if err := c.ShouldBindJSON(&requestPayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		valid := veriflow.Svc.VerifyOTP(sessionUser.Authenticator.Secret, requestPayload.Code)
		if !valid {
			c.JSON(http.StatusOK, gin.H{
				"error": "invalid code",
			})
			return
		}

		sessionUser.Authenticator.Validated = true
		models.SaveUser(sessionUser)

		c.JSON(http.StatusOK, gin.H{
			"verified": "ok",
		})
	})

	// WebauthN endpoints
	api.POST("/begin-registration", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")

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

		options, err := veriflow.WebAuthnService.BeginRegistration(sessionUser.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, options)
	})

	api.POST("/finish-registration", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")

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

		err := veriflow.WebAuthnService.FinishRegistration(sessionUser.ID, c.Request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	api.POST("/begin-login/:uuid", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")

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

		options, err := veriflow.WebAuthnService.BeginLogin(sessionUser.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, options)
	})

	api.POST("/finish-login/:uuid", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")

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

		err := veriflow.WebAuthnService.FinishLogin(sessionUser.ID, c.Request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		uuid := c.Param("uuid")
		vr, _ := models.GetByID(uuid)
		veriflow.Svc.SendConfirmation(vr)

		c.JSON(http.StatusOK, gin.H{})
	})

	return router
}
