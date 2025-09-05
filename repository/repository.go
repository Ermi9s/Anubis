package repository

import "github.com/Ermi9s/Anubis/model"




type Repository struct {
	DB model.DataBase
}


func NewRepository(database model.DataBase) *Repository {
	return &Repository{
		DB: database,
	}
}


func (r *Repository) CreateAudit(event model.AuditEvent) error {
	return r.DB.CreateEvent(event)
}


func (r *Repository) FindAudit(filter model.AuditEventFilter) (model.PaginatedResponse, error) {
	return r.DB.Find(filter)
}


