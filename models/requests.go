package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/coreos/go-oidc/v3/oidc"
)

type VerifyRequest struct {
	ID             string `dynamodbav:"id"`
	RequestorEmail string `dynamodbav:"requestor_email"`

	Start             time.Time `dynamodbav:"start_time"`
	End               time.Time `dynamodbav:"end_time,omitempty"`
	CommunicationTool string    `dynamodbav:"communication_tool"`
	Status            string    `dynamodbav:"status"`

	Requestor User `dynamodbav:"requestor"`
	Recipient User `dynamodbav:"recipient"`

	Message string `dynamodbav:"message"`
	Error   string `dynamodbav:"error,omitempty"`

	OIDCInfo string `dynamodbav:"user_info,omitempty"`

	StatusLink string `dynamodbav:"status_link"`
	AuthLink   string `dynamodbav:"auth_link"`
	ReportLink string `dynamodbav:"report_link"`

	Slack SlackRequest `json:"slack"  dynamodbav:"slack"`

	Permalink string `json:"permalink"  dynamodbav:"permalink"`

	DurationSinceStart string `json:"duration_since_start" dynamodbav:"-"`
	Duration           string `json:"duration" dynamodbav:"-"`
}

type SlackRequest struct {
	UserID      string `json:"user_id"`
	ResponseURL string `json:"response_url"`
	RecipientID string `json:"recipient_id"`
	Text        string `json:"text"`

	MessageTS string `json:"message_ts"`
	Channel   string `json:"channel"`

	InitMessageTS string `json:"init_message_ts"`
	InitChannel   string `json:"init_channel"`
}

func GetByID(id string) (VerifyRequest, error) {
	var vr VerifyRequest

	input := &dynamodb.QueryInput{
		TableName:              aws.String(VeriflowRequestsTable),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: id},
		}, Limit: aws.Int32(1), // Add a limit to minimize read consumption
	}

	dbClient.QueryOne(input, &vr)

	if !vr.End.IsZero() {
		duration := vr.End.Sub(vr.Start)
		vr.Duration = formatDuration(duration)
	} else {
		duration := time.Since(vr.Start)
		vr.DurationSinceStart = formatDuration(duration)
	}

	return vr, nil
}

func (vr *VerifyRequest) Get() error {

	key := map[string]types.AttributeValue{
		"id":              &types.AttributeValueMemberS{Value: vr.ID},
		"requestor_email": &types.AttributeValueMemberS{Value: vr.RequestorEmail}, // Ensure this is correctly populated
	}

	err := dbClient.Get(VeriflowRequestsTable, key, vr)

	if !vr.End.IsZero() {
		duration := vr.End.Sub(vr.Start)
		vr.Duration = formatDuration(duration)
	} else {
		duration := time.Since(vr.Start)
		vr.DurationSinceStart = formatDuration(duration)
	}

	return err
}

func (vr *VerifyRequest) PrepareForSave() {
	if vr.Requestor.Email != "" {
		vr.RequestorEmail = vr.Requestor.Email
	}
}

func (vr *VerifyRequest) Save() error {
	vr.PrepareForSave()

	av, err := attributevalue.MarshalMap(vr)
	if err != nil {
		return err
	}
	return dbClient.Save(VeriflowRequestsTable, av)
}

func (vr *VerifyRequest) SetOIDCInfo(info *oidc.UserInfo) error {
	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}
	vr.OIDCInfo = string(bytes)
	return nil
}

func (vr *VerifyRequest) GetOIDCInfo() (*oidc.UserInfo, error) {
	var info oidc.UserInfo
	if err := json.Unmarshal([]byte(vr.OIDCInfo), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (vr *VerifyRequest) Done() error {
	vr.End = time.Now()
	vr.Status = "COMPLETED"
	return vr.Save()
}

func (vr *VerifyRequest) DoneWithError(errorString string) error {
	vr.End = time.Now()
	vr.Status = "FAILED"
	vr.Error = errorString
	return vr.Save()
}

func GetAllByUser(email string) ([]VerifyRequest, error) {
	var requests []VerifyRequest

	input := &dynamodb.QueryInput{
		TableName:              aws.String("VeriflowRequests"),
		IndexName:              aws.String("EmailIndex"), // Specify the name of your GSI
		KeyConditionExpression: aws.String("requestor_email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
	}

	dbClient.QueryAll(input, &requests)

	return requests, nil
}

func GetAllByRecipientEmail(email string) ([]VerifyRequest, error) {
	filt := expression.Name("recipient.email").Equal(expression.Value(email))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()

	if err != nil {
		return nil, fmt.Errorf("error building expression: %v", err)
	}

	params := &dynamodb.ScanInput{
		TableName:                 aws.String("VeriflowRequests"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}

	var verifyRequests []VerifyRequest
	err = dbClient.Scan(params, &verifyRequests)
	if err != nil {
		return nil, fmt.Errorf("error scanning table: %v", err)
	}

	return verifyRequests, nil
}

func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds()) % 60
	minutes := int(d.Minutes()) % 60
	hours := int(d.Hours()) % 24
	days := int(d.Hours() / 24)

	// Construct your string based on the values
	result := ""
	if days > 0 {
		result += fmt.Sprintf("%d days ", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%d hours ", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%d minutes ", minutes)
	}
	result += fmt.Sprintf("%d seconds", seconds)

	return result
}
