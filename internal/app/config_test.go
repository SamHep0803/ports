package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigRejectsDuplicateProfileNames(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, `
profiles:
  - name: dup
    host: a.example.com
    user: sam
    keyPath: ~/.ssh/id_rsa
    forwards:
      - localPort: 5432
        remoteHost: 127.0.0.1
        remotePort: 5432
  - name: dup
    host: b.example.com
    user: sam
    keyPath: ~/.ssh/id_rsa
`)

	_, err := LoadConfig(path)
	if err == nil || !strings.Contains(err.Error(), "duplicate profile name") {
		t.Fatalf("expected duplicate profile validation error, got: %v", err)
	}
}

func TestLoadConfigRejectsInvalidForwardPort(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, `
profiles:
  - name: db
    host: db.example.com
    user: sam
    keyPath: ~/.ssh/id_rsa
    forwards:
      - localPort: 0
        remoteHost: 127.0.0.1
        remotePort: 5432
`)

	_, err := LoadConfig(path)
	if err == nil || !strings.Contains(err.Error(), "localPort must be between 1 and 65535") {
		t.Fatalf("expected localPort validation error, got: %v", err)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}
