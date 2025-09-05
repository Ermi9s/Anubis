package config

import (
	"github.com/Ermi9s/Anubis/model"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)


type PostgresDb struct {
	Client *pgxpool.Pool
}

func ConnectPostgres(cfg *model.Configuration) (*pgxpool.Pool, error) {
    connString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s", cfg.Postgres.Host, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DbName, cfg.Postgres.Port, cfg.Postgres.TimeZone)

	log.Printf("[Anubis] Connection string -> %s", connString)
	
	err := RunMigrations(connString, "pg_db/migrations")
	if err != nil {
		log.Fatalf("[Anubis Error] failed to run migrations: %s", err)
	}

	pgxp, err := pgxpool.New(context.Background(), connString)

	if err != nil {
		return &pgxpool.Pool{}, err
	}

	log.Println("[Anubis] Postgres connected succesfully")
	return pgxp, nil
}


func NewPostgresDb(client *pgxpool.Pool) *PostgresDb {
	return &PostgresDb{
		Client: client,
	}
}


func (pg *PostgresDb) CreateEvent(event model.AuditEvent) error {
	query := `
		INSERT INTO audit_event (
			event_id, timestamp, action, status, actor_id, actor_type,
			ip_address, user_agent, resource, resource_id, details, service_name
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`

	_, err := pg.Client.Exec(context.Background(), query,
		event.EventID,
		event.Timestamp,
		event.Action,
		event.Status,
		event.ActorID,
		event.ActorType,
		event.IPAddress,
		event.UserAgent,
		event.Resource,
		event.ResourceID,
		event.Details,
		event.ServiceName,
	)
	return err
}


func (pg *PostgresDb) Find(filter model.AuditEventFilter) (model.PaginatedResponse, error) {
	// base query
	baseQuery := `
		SELECT event_id, timestamp, action, status, actor_id, actor_type,
		       ip_address, user_agent, resource, resource_id, details, service_name
		FROM audit_event
	`
	countQuery := `SELECT COUNT(*) FROM audit_event`


	whereClauses := []string{}
	args := []interface{}{}
	argID := 1

	if filter.Action != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("action = $%d", argID))
		args = append(args, *filter.Action)
		argID++
	}
	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argID))
		args = append(args, *filter.Status)
		argID++
	}
	if filter.ActorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("actor_id = $%d", argID))
		args = append(args, *filter.ActorID)
		argID++
	}
	if filter.ActorType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("actor_type = $%d", argID))
		args = append(args, *filter.ActorType)
		argID++
	}
	if filter.StartTime != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("timestamp >= $%d", argID))
		args = append(args, *filter.StartTime)
		argID++
	}
	if filter.EndTime != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("timestamp <= $%d", argID))
		args = append(args, *filter.EndTime)
		argID++
	}
	if filter.ServiceName != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("service_name = $%d", argID))
		args = append(args, *filter.ServiceName)
		argID++
	}

	if len(whereClauses) > 0 {
		whereSQL := " WHERE " + strings.Join(whereClauses, " AND ")
		baseQuery += whereSQL
		countQuery += whereSQL
	}


	var total int64
	if err := pg.Client.QueryRow(context.Background(), countQuery, args...).Scan(&total); err != nil {
		return model.PaginatedResponse{}, err
	}

	sortBy := "timestamp"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if strings.ToLower(filter.SortOrder) == "asc" {
		sortOrder = "ASC"
	}


	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize


	baseQuery += fmt.Sprintf(" ORDER BY %s %s LIMIT %d OFFSET %d", sortBy, sortOrder, pageSize, offset)


	rows, err := pg.Client.Query(context.Background(), baseQuery, args...)
	if err != nil {
		return model.PaginatedResponse{}, err
	}
	defer rows.Close()

	events := []model.AuditEvent{}
	for rows.Next() {
		var e model.AuditEvent
		if err := rows.Scan(
			&e.EventID,
			&e.Timestamp,
			&e.Action,
			&e.Status,
			&e.ActorID,
			&e.ActorType,
			&e.IPAddress,
			&e.UserAgent,
			&e.Resource,
			&e.ResourceID,
			&e.Details,
			&e.ServiceName,
		); err != nil {
			return model.PaginatedResponse{}, err
		}
		events = append(events, e)
	}

	return model.PaginatedResponse{
		Data: events,
		Pagination: model.Pagination{
			Page:  page,
			Total: total,
		},
	}, nil
}
