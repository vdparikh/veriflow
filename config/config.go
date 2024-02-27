package config

import (
	"fmt"
	"log"
	"os"

	"github.com/vdparikh/veriflow/utils"
	"gopkg.in/yaml.v2"
)

type Config struct {
	BaseURL string `yaml:"base_url"`

	Communication struct {
		ActiveService string `yaml:"active_service"`
		Services      struct {
			Slack struct {
				AppToken string `yaml:"app_token"`
				BotToken string `yaml:"bot_token"`
			} `yaml:"slack"`
			MSTeams struct {
				WebhookURL string `yaml:"webhook_url"`
			} `yaml:"ms_teams"`
		} `yaml:"services"`
	} `yaml:"communication"`

	Email struct {
		Enabled    bool   `yaml:"enabled"`
		From       string `yaml:"from"`
		SMTPServer string `yaml:"smtp_server"`
		Port       int    `yaml:"port"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
	} `yaml:"email"`

	Auth struct {
		Provider     string
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		Issuer       string `yaml:"issuer"`
		CallbackURL  string `yaml:"callback_url"`
	} `yaml:"auth"`

	Report struct {
		URL string `yaml:"url"`
	} `yaml:"report"`

	Authenticator struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"authenticator"`

	Messages Messages `yaml:"messages"`
}

type Messages struct {
	VerificationMessage                 string `yaml:"verification_message"`
	RequestConfirmationMessage          string `yaml:"request_confirmation_message"`
	RequestorCompletionMessage          string `yaml:"requestor_completion_message"`
	RecipientCompletionMessage          string `yaml:"recipient_completion_message"`
	RequestorVerificationFailureMessage string `yaml:"requestor_verification_failure_message"`
	RecipientVerificationFailureMessage string `yaml:"recipient_verification_failure_message"`
}

func LoadConfig() (Config, error) {
	var config Config

	var data []byte
	var err error

	if os.Getenv("LOCAL") != "" {
		data, err = os.ReadFile("config.yaml")
	} else {
		// Assuming downloadFileFromS3 returns a string and error
		bucketName := os.Getenv("VERIFLOW_BUCKET")

		if bucketName == "" {
			log.Fatal("environment variable for VERIFLOW_BUCKET missing")
		}

		var configData string
		configData, err = utils.DownloadFileFromS3(bucketName, "config.yaml")
		data = []byte(configData)
	}

	if err != nil {
		return config, fmt.Errorf("failed to load config: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
