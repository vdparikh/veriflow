package veriflow

import (
	"net/url"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/vdparikh/veriflow/config"
	"github.com/vdparikh/veriflow/models"
	"github.com/vdparikh/veriflow/service"

	log "github.com/sirupsen/logrus"

	webauthnservice "github.com/vdparikh/veriflow/webauthn"
)

type Veriflow struct {
	Cfg               config.Config
	Svc               *service.VerificationService
	WebAuthnService   *webauthnservice.WebAuthnService
	VerificationQueue chan models.VerifyRequest
	UserStore         models.UserStore
	Logger            *log.Logger
}

func New() *Veriflow {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	logger := log.New()
	logger.Formatter = &log.JSONFormatter{}
	logger.Level = log.InfoLevel

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize webAuthn
	userStore := models.NewUserStore()
	parsedURL, _ := url.Parse(cfg.BaseURL)
	rpid := parsedURL.Hostname()
	webAuthnService, err := webauthnservice.NewWebAuthnService(&webauthn.Config{
		RPDisplayName: "Veriflow",
		RPID:          rpid,
		RPOrigin:      cfg.BaseURL,
	}, userStore)

	if err != nil {
		log.Fatalf("Failed to create WebAuthn service: %v", err)
	}

	svc, err := service.NewVerificationService(&cfg)
	if err != nil {
		log.Fatalf("Failed to create Verification service: %v", err)
	}

	veriflow := Veriflow{
		Cfg:               cfg,
		Svc:               svc,
		WebAuthnService:   webAuthnService,
		UserStore:         userStore,
		VerificationQueue: make(chan models.VerifyRequest, 100),
		Logger:            logger,
	}

	// go veriflow.StartVerificationWorker()

	return &veriflow
}

// // Queue to manage influx of verification requests
// func (veriflow *Veriflow) StartVerificationWorker() {
// 	for request := range veriflow.VerificationQueue {
// 		err := veriflow.Svc.NewVerification(request)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 	}
// }
