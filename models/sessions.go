package models

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-webauthn/webauthn/webauthn"
)

func (u *User) SaveSessionData(userID string, sessionData *webauthn.SessionData) error {
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	err = dbClient.Save(VeriflowWebAuthnSessionTable, map[string]types.AttributeValue{
		"id":   &types.AttributeValueMemberS{Value: userID},
		"data": &types.AttributeValueMemberS{Value: string(sessionDataJSON)},
	})
	return err
}

func (u *User) GetSessionData(userID string) (*webauthn.SessionData, error) {

	out, err := dbClient.GetItem(VeriflowWebAuthnSessionTable, map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: userID},
	})
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, fmt.Errorf("session data not found for user ID: %s", userID)
	}

	dataVal, ok := out["data"]
	if !ok {
		return nil, fmt.Errorf("missing session data for user ID: %s", userID)
	}

	var sessionData webauthn.SessionData
	err = json.Unmarshal([]byte(dataVal.(*types.AttributeValueMemberS).Value), &sessionData)
	if err != nil {
		return nil, err
	}

	return &sessionData, nil
}

func (u *User) DeleteSessionData(userID string) error {
	return dbClient.Delete(VeriflowWebAuthnSessionTable, userID)
}
