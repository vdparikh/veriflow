package models

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-webauthn/webauthn/webauthn"
)

type UserStore interface {
	GetUserByID(userID string) (webauthn.User, error)
	SaveSessionData(userID string, sessionData *webauthn.SessionData) error
	GetSessionData(userID string) (*webauthn.SessionData, error)
	DeleteSessionData(userID string) error
	SaveCredentials(userID string, credential *webauthn.Credential) error
	GetCredentials(userID string) []webauthn.Credential
}

type User struct {
	ID              string                `dynamodbav:"id"`
	Name            string                `dynamodbav:"name"`
	Email           string                `dynamodbav:"email"`
	Image           string                `dynamodbav:"image"`
	Authenticator   Authenticator         `dynamodbav:"authenticator"`
	CredentialsJSON string                `dynamodbav:"credentialsJSON"`
	Credentials     []webauthn.Credential `dynamodbav:"-"`
}

type Authenticator struct {
	Secret    string `dynamodbav:"secret"`
	FilePath  string `dynamodbav:"filePath"`
	Permalink string `dynamodbav:"permalink"`
	Validated bool   `dynamodbav:"validated"`
}

func NewUserStore() UserStore {
	return &User{}
}

func SaveUser(user User) error {
	// Encode Credentials to JSON
	if len(user.Credentials) > 0 {
		credsJSON, err := json.Marshal(user.Credentials)
		if err != nil {
			return fmt.Errorf("failed to marshal credentials: %w", err)
		}
		user.CredentialsJSON = string(credsJSON)
	}

	av, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	return dbClient.Save(VeriflowUsersTable, av)
}

func GetUser(userID string) (User, error) {
	var user User

	dbClient.GetItemById(VeriflowUsersTable, userID, &user)
	// Decode CredentialsJSON
	if user.CredentialsJSON != "" {
		err := json.Unmarshal([]byte(user.CredentialsJSON), &user.Credentials)
		if err != nil {
			return user, fmt.Errorf("failed to unmarshal credentials JSON: %w", err)
		}
	}

	return user, nil
}

func GetOrCreateUser(userEmail string) (User, error) {
	user, err := GetUser(userEmail)
	if user.ID == "" {
		user.ID = userEmail
		user.Email = userEmail
		user.Name = userEmail

		SaveUser(user)
		return user, err
	}

	if user.CredentialsJSON != "" {
		err := json.Unmarshal([]byte(user.CredentialsJSON), &user.Credentials)
		if err != nil {
			return user, fmt.Errorf("failed to unmarshal credentials JSON: %w", err)
		}
	}

	return user, nil
}

func GetUserByEmail(userEmail string) (User, error) {
	var user User

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(VeriflowUsersTable),
		IndexName:              aws.String("EmailIndex"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: userEmail},
		},
		Limit: aws.Int32(1),
	}

	err := dbClient.QueryOne(queryInput, &user)

	if user.CredentialsJSON != "" {
		err := json.Unmarshal([]byte(user.CredentialsJSON), &user.Credentials)
		if err != nil {
			return user, fmt.Errorf("failed to unmarshal credentials JSON: %w", err)
		}
	}

	return user, err
}

func (u *User) GetUserByID(userID string) (webauthn.User, error) {
	user, err := GetUser(userID)
	if err != nil {
		return &user, err
	}
	return &user, nil
}

func (u *User) SaveCredentials(userID string, credentials *webauthn.Credential) error {
	user, _ := GetUser(userID)
	user.Credentials = append(user.Credentials, *credentials)
	return SaveUser(user)
}

func (u *User) GetCredentials(userID string) []webauthn.Credential {
	user, _ := GetUser(userID)
	return user.Credentials
}

func (u *User) WebAuthnID() []byte {
	return []byte(u.ID)
}

func (u *User) WebAuthnName() string {
	return u.Name
}

func (u *User) WebAuthnDisplayName() string {
	return u.Name
}

func (u *User) WebAuthnIcon() string {
	return u.Image
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}
