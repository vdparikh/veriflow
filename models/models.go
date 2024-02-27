package models

import "github.com/vdparikh/veriflow/db"

var dbClient db.DB

const (
	VeriflowUsersTable           = "VeriflowUsers"
	VeriflowWebAuthnSessionTable = "VeriflowWebAuthnSessions"
	VeriflowRequestsTable        = "VeriflowRequests"
)

func init() {
	dbClient = db.Init()
}
