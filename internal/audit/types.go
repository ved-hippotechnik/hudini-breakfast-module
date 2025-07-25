package audit

// AuditAction represents the actions that can be audited
type AuditAction string

const (
	ActionCreate   AuditAction = "CREATE"
	ActionUpdate   AuditAction = "UPDATE"
	ActionDelete   AuditAction = "DELETE"
	ActionRead     AuditAction = "READ"
	ActionLogin    AuditAction = "LOGIN"
	ActionLogout   AuditAction = "LOGOUT"
	ActionConsume  AuditAction = "CONSUME_BREAKFAST"
	ActionExport   AuditAction = "EXPORT"
	ActionReport   AuditAction = "GENERATE_REPORT"
)

// AuditResource represents the resources that can be audited
type AuditResource string

const (
	ResourceGuest       AuditResource = "GUEST"
	ResourceRoom        AuditResource = "ROOM"
	ResourceConsumption AuditResource = "BREAKFAST_CONSUMPTION"
	ResourceStaff       AuditResource = "STAFF"
	ResourceProperty    AuditResource = "PROPERTY"
	ResourceOutlet      AuditResource = "OUTLET"
	ResourceAuth        AuditResource = "AUTHENTICATION"
	ResourceReport      AuditResource = "REPORT"
	ResourceAnalytics   AuditResource = "ANALYTICS"
)