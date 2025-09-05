package model


type DataBase interface {
	CreateEvent(AuditEvent) error
	Find(AuditEventFilter) (PaginatedResponse, error)
}


