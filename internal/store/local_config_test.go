package store

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/Mininglamp-OSS/octo-speech/internal/config"
)

func TestUpsert_AllFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	s := &LocalConfigStore{db: db}

	timeout := 3000
	probe := "http://probe"
	transcribe := "http://transcribe"

	mock.ExpectExec(
		`INSERT INTO local_asr_config \(app_id, subject_id, scope_type, scope_id, enabled, timeout_ms, probe_url, transcribe_url\)`+
			` VALUES \(\?, \?, \?, \?, \?, \?, \?, \?\)`+
			` ON DUPLICATE KEY UPDATE enabled=VALUES\(enabled\), timeout_ms=VALUES\(timeout_ms\), probe_url=VALUES\(probe_url\), transcribe_url=VALUES\(transcribe_url\)`,
	).WithArgs("app1", "sub1", "org", "scope1", true, 3000, "http://probe", "http://transcribe").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.Upsert("app1", "sub1", "org", "scope1", true, &timeout, &probe, &transcribe)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpsert_EnabledOnly_NoOptionalFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	s := &LocalConfigStore{db: db}

	mock.ExpectExec(
		`INSERT INTO local_asr_config \(app_id, subject_id, scope_type, scope_id, enabled\)`+
			` VALUES \(\?, \?, \?, \?, \?\)`+
			` ON DUPLICATE KEY UPDATE enabled=VALUES\(enabled\)$`,
	).WithArgs("app1", "sub1", "org", "scope1", false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.Upsert("app1", "sub1", "org", "scope1", false, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpsert_PartialFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	s := &LocalConfigStore{db: db}

	timeout := 5000

	mock.ExpectExec(
		`INSERT INTO local_asr_config \(app_id, subject_id, scope_type, scope_id, enabled, timeout_ms\)`+
			` VALUES \(\?, \?, \?, \?, \?, \?\)`+
			` ON DUPLICATE KEY UPDATE enabled=VALUES\(enabled\), timeout_ms=VALUES\(timeout_ms\)$`,
	).WithArgs("app1", "sub1", "org", "scope1", true, 5000).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.Upsert("app1", "sub1", "org", "scope1", true, &timeout, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDelete_RowsAffected(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	s := &LocalConfigStore{db: db}

	mock.ExpectExec(`DELETE FROM local_asr_config`).
		WithArgs("app1", "sub1", "org", "scope1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows, err := s.Delete("app1", "sub1", "org", "scope1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDelete_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	s := &LocalConfigStore{db: db}

	mock.ExpectExec(`DELETE FROM local_asr_config`).
		WithArgs("app1", "sub1", "org", "scope1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	rows, err := s.Delete("app1", "sub1", "org", "scope1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows != 0 {
		t.Errorf("expected 0 rows affected, got %d", rows)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestResetToDefault_EnabledTrue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cfg := &config.Config{
		LocalTimeoutMs:     10000,
		LocalProbeURL:      "http://localhost:8787/",
		LocalTranscribeURL: "http://localhost:8787/v1/voice/transcribe",
	}
	s := NewLocalConfigStore(db, cfg)

	mock.ExpectExec(
		`INSERT INTO local_asr_config \(app_id, subject_id, scope_type, scope_id, enabled, timeout_ms, probe_url, transcribe_url\)`+
			` VALUES \(\?, \?, \?, \?, \?, \?, \?, \?\)`+
			` ON DUPLICATE KEY UPDATE`+
			`\s+enabled=VALUES\(enabled\),`+
			`\s+timeout_ms=VALUES\(timeout_ms\),`+
			`\s+probe_url=VALUES\(probe_url\),`+
			`\s+transcribe_url=VALUES\(transcribe_url\)`,
	).WithArgs("app1", "sub1", "space", "scope1", true, 10000, "http://localhost:8787/", "http://localhost:8787/v1/voice/transcribe").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.ResetToDefault("app1", "sub1", "space", "scope1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestResetToDefault_EnabledFalse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cfg := &config.Config{
		LocalTimeoutMs:     5000,
		LocalProbeURL:      "http://example.com/probe",
		LocalTranscribeURL: "http://example.com/transcribe",
	}
	s := NewLocalConfigStore(db, cfg)

	mock.ExpectExec(
		`INSERT INTO local_asr_config \(app_id, subject_id, scope_type, scope_id, enabled, timeout_ms, probe_url, transcribe_url\)`+
			` VALUES \(\?, \?, \?, \?, \?, \?, \?, \?\)`+
			` ON DUPLICATE KEY UPDATE`+
			`\s+enabled=VALUES\(enabled\),`+
			`\s+timeout_ms=VALUES\(timeout_ms\),`+
			`\s+probe_url=VALUES\(probe_url\),`+
			`\s+transcribe_url=VALUES\(transcribe_url\)`,
	).WithArgs("app1", "sub1", "org", "scope1", false, 5000, "http://example.com/probe", "http://example.com/transcribe").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.ResetToDefault("app1", "sub1", "org", "scope1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
