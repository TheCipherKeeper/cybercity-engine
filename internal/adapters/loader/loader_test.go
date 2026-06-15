package loader

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func makeEngineZip(t *testing.T) []byte {
	t.Helper()
	runtime := map[string]any{
		"schema_version": 1,
		"tick_ms":        1000,
		"services": []any{
			map[string]any{
				"id":                  "bank-web",
				"org_id":              "bank",
				"name":                "Bank portal",
				"host":                "portal.bank.corp",
				"ip":                  "10.1.0.10",
				"kind":                "web",
				"exposure":            "public",
				"network_id":          "bank-dmz",
				"auth":                "sso",
				"data_classification": "public",
				"criticality":         "high",
				"ports":               []string{"tcp/443"},
			},
			map[string]any{
				"id":                  "bank-db",
				"org_id":              "bank",
				"name":                "Bank DB",
				"host":                "db.bank.corp",
				"ip":                  "10.1.1.10",
				"kind":                "db",
				"exposure":            "intranet",
				"network_id":          "bank-lan",
				"auth":                "certificate",
				"data_classification": "pci",
				"criticality":         "critical",
				"ports":               []string{"tcp/1521"},
			},
		},
		"links": []any{
			map[string]any{
				"from":       "bank-web",
				"to":         "bank-db",
				"kind":       "db-read",
				"protocol":   "tcp/1521",
				"encryption": "tls",
			},
		},
	}

	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	w, err := zw.Create("runtime/engine.json")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	data, err := json.Marshal(runtime)
	if err != nil {
		t.Fatalf("marshal runtime: %v", err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func TestLoadTopologyFromZip(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "engine.zip")
	if err := os.WriteFile(zipPath, makeEngineZip(t), 0o644); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	l := NewLoader()
	graph, err := l.Load(context.Background(), zipPath)
	if err != nil {
		t.Fatalf("load topology: %v", err)
	}
	if graph.SchemaVersion != "1" {
		t.Errorf("expected schema version 1, got %q", graph.SchemaVersion)
	}
	if _, ok := graph.Services["bank-web"]; !ok {
		t.Fatal("expected bank-web service")
	}
	if graph.Services["bank-web"].Kind != "web" {
		t.Errorf("expected kind web, got %q", graph.Services["bank-web"].Kind)
	}
	if graph.Services["bank-web"].BindIP != "10.1.0.10" {
		t.Errorf("expected bind ip 10.1.0.10, got %q", graph.Services["bank-web"].BindIP)
	}
	if len(graph.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(graph.Edges))
	}
	if graph.Edges[0].Kind != "db-read" {
		t.Errorf("expected edge kind db-read, got %q", graph.Edges[0].Kind)
	}
}

func TestLoadTopologyEmpty(t *testing.T) {
	l := NewLoader()
	graph, err := l.Load(context.Background(), "")
	if err != nil {
		t.Fatalf("empty topology should not error: %v", err)
	}
	if len(graph.Services) != 0 {
		t.Errorf("expected empty services, got %d", len(graph.Services))
	}
}
