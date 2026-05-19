package store

import (
	"database/sql"
	"time"
)

type VocabularyProfile struct {
	ID        int64     `json:"id"`
	AppID     string    `json:"app_id"`
	SubjectID string    `json:"subject_id"`
	ScopeType string    `json:"scope_type"`
	ScopeID   string    `json:"scope_id"`
	Content   string    `json:"content"`
	UpdatedBy string    `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var validScopeTypes = map[string]bool{
	"global":  true,
	"space":   true,
	"org":     true,
	"project": true,
}

func IsValidScopeType(s string) bool {
	return validScopeTypes[s]
}

type VocabularyStore struct {
	db *sql.DB
}

func NewVocabularyStore(db *sql.DB) *VocabularyStore {
	return &VocabularyStore{db: db}
}

func (s *VocabularyStore) Upsert(appID, subjectID, scopeType, scopeID, content, updatedBy string) error {
	_, err := s.db.Exec(
		`INSERT INTO vocabulary_profile (app_id, subject_id, scope_type, scope_id, content, updated_by)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE content=VALUES(content), updated_by=VALUES(updated_by)`,
		appID, subjectID, scopeType, scopeID, content, updatedBy,
	)
	return err
}

func (s *VocabularyStore) QueryWithPriority(appID, subjectID, scopeType, scopeID string) (*VocabularyProfile, error) {
	var v VocabularyProfile
	err := s.db.QueryRow(
		`SELECT id, app_id, subject_id, scope_type, scope_id, content, updated_by, created_at, updated_at
		 FROM vocabulary_profile
		 WHERE app_id = ? AND subject_id = ? AND scope_type = ? AND scope_id = ?`,
		appID, subjectID, scopeType, scopeID,
	).Scan(&v.ID, &v.AppID, &v.SubjectID, &v.ScopeType, &v.ScopeID, &v.Content, &v.UpdatedBy, &v.CreatedAt, &v.UpdatedAt)
	if err == nil {
		return &v, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	err = s.db.QueryRow(
		`SELECT id, app_id, subject_id, scope_type, scope_id, content, updated_by, created_at, updated_at
		 FROM vocabulary_profile
		 WHERE app_id = ? AND subject_id = ? AND scope_type = 'global' AND scope_id = 'default'`,
		appID, subjectID,
	).Scan(&v.ID, &v.AppID, &v.SubjectID, &v.ScopeType, &v.ScopeID, &v.Content, &v.UpdatedBy, &v.CreatedAt, &v.UpdatedAt)
	if err == nil {
		return &v, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	err = s.db.QueryRow(
		`SELECT id, app_id, subject_id, scope_type, scope_id, content, updated_by, created_at, updated_at
		 FROM vocabulary_profile
		 WHERE app_id = '*' AND subject_id = ? AND scope_type = 'global' AND scope_id = 'default'`,
		subjectID,
	).Scan(&v.ID, &v.AppID, &v.SubjectID, &v.ScopeType, &v.ScopeID, &v.Content, &v.UpdatedBy, &v.CreatedAt, &v.UpdatedAt)
	if err == nil {
		return &v, nil
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return nil, err
}

func (s *VocabularyStore) Delete(appID, subjectID, scopeType, scopeID string) error {
	_, err := s.db.Exec(
		`DELETE FROM vocabulary_profile
		 WHERE app_id = ? AND subject_id = ? AND scope_type = ? AND scope_id = ?`,
		appID, subjectID, scopeType, scopeID,
	)
	return err
}
