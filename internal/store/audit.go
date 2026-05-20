package store

import (
	"context"
	"database/sql"
	"encoding/json"
)

type AuditEntry struct {
	Action   string
	AppID    string
	AppName  string
	Operator string
	Detail   map[string]any
}

type AuditStore struct {
	db *sql.DB
}

func NewAuditStore(db *sql.DB) *AuditStore {
	return &AuditStore{db: db}
}

func (s *AuditStore) Log(ctx context.Context, entry AuditEntry) error {
	var detail []byte
	if entry.Detail != nil {
		var err error
		detail, err = json.Marshal(entry.Detail)
		if err != nil {
			return err
		}
	}

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO audit_log (action, app_id, app_name, operator, detail) VALUES (?, ?, ?, ?, ?)",
		entry.Action, entry.AppID, entry.AppName, entry.Operator, detail,
	)
	return err
}
