package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditEvent struct {
	EventID     uuid.UUID       `json:"event_id"`
	Timestamp   time.Time       `json:"timestamp"`
	Action      string          `json:"action"`
	Status      string          `json:"status"`
	ActorID     string          `json:"actor_id"`
	ActorType   string          `json:"actor_type"`
	IPAddress   string          `json:"ip_address,omitempty"`
	UserAgent   string          `json:"user_agent,omitempty"`
	Resource    string          `json:"resource,omitempty"`
	ResourceID  string          `json:"resource_id,omitempty"`
	Details     json.RawMessage `json:"details,omitempty"`
	ServiceName string          `json:"service_name"`
}

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type PaginatedResponse struct {
	Data       []AuditEvent `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

type AuditEventFilter struct {
	Action      *string
	Status      *string
	ActorID     *string
	ActorType   *string
	StartTime   *time.Time
	EndTime     *time.Time
	ServiceName *string
	Page        int
	PageSize    int
	SortBy      string
	SortOrder   string
}

type RabbitMQConf struct {
	Address    string                 `yaml:"address"`
	QueueName  string                 `yaml:"queue_name"`
	Durable    bool                   `yaml:"durable"`
	AutoDelete bool                   `yaml:"auto_delete"`
	Exclusive  bool                   `yaml:"exclusive"`
	NoWait     bool                   `yaml:"no_wait"`
	Args       map[string]interface{} `yaml:"args"`
}

type Postgres struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"db_name"`
	Port     string `yaml:"port"`
	Sslmode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

type Configuration struct {
	RabbitMQ   *RabbitMQConf `yaml:"rabbit_mq,omitempty"`
	Postgres   *Postgres     `yaml:"postgres"`
	Concurency int           `yaml:"read_concurency"`
}
