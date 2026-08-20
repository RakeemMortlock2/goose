package goose

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestStatusMissingLocal(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE goose_db_version (id INTEGER PRIMARY KEY AUTOINCREMENT, version_id INTEGER NOT NULL, is_applied BOOLEAN NOT NULL, tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO goose_db_version (version_id, is_applied) VALUES (1, 1)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO goose_db_version (version_id, is_applied) VALUES (2, 1)")
	if err != nil {
		t.Fatal(err)
	}

	tmpDir, err := os.MkdirTemp("", "goose_status_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	f1, err := os.Create(filepath.Join(tmpDir, "00001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	f1.Close()

	err = SetDialect("sqlite3")
	if err != nil {
		t.Fatal(err)
	}

	oldStdout := stdout
	var buf bytes.Buffer
	stdout = &buf
	defer func() {
		stdout = oldStdout
	}()

	err = Status(db, tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "00001_init.sql") {
		t.Errorf("expected output to contain 00001_init.sql, got:\n%s", output)
	}
	if !strings.Contains(output, "00002 (missing local file)") {
		t.Errorf("expected output to contain 00002 (missing local file), got:\n%s", output)
	}
}
