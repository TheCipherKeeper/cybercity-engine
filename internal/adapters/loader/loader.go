// Package loader адаптирует форматы cybercity-data к domain TopologyGraph.
package loader

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/models"
	"github.com/TheCipherKeeper/cybercity-engine/internal/domain/ports"
)

// Loader реализует ports.TopologyLoader.
type Loader struct{}

// NewLoader создаёт новый loader.
func NewLoader() *Loader {
	return &Loader{}
}

// Load загружает TopologyGraph из engine.zip, engine.json или возвращает
// пустой граф, если source == "".
func (l *Loader) Load(ctx context.Context, source string) (*models.TopologyGraph, error) {
	if source == "" {
		return &models.TopologyGraph{
			SchemaVersion: "3.0.0",
			SourceVersion: "0.0.0",
			Services:      make(map[string]*models.TopologyNode),
			Edges:         []models.TopologyEdge{},
		}, nil
	}

	raw, err := loadRaw(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("load topology from %q: %w", source, err)
	}

	return parseTopology(raw)
}

// Compile-time check.
var _ ports.TopologyLoader = (*Loader)(nil)

func loadRaw(ctx context.Context, source string) (map[string]any, error) {
	var data []byte
	var err error

	if isURL(source) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
	} else {
		data, err = os.ReadFile(source)
		if err != nil {
			return nil, err
		}
	}

	if filepath.Ext(source) == ".zip" {
		return loadFromZip(data)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func isURL(source string) bool {
	u, err := url.Parse(source)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func loadFromZip(data []byte) (map[string]any, error) {
	r := bytes.NewReader(data)
	zr, err := zip.NewReader(r, int64(len(data)))
	if err != nil {
		return nil, err
	}

	for _, f := range zr.File {
		if f.Name == "runtime/engine.json" || f.Name == "model/network.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			var raw map[string]any
			if err := json.Unmarshal(content, &raw); err != nil {
				return nil, err
			}
			return raw, nil
		}
	}
	return nil, fmt.Errorf("engine.zip has no runtime/engine.json or model/network.json")
}

func parseTopology(raw map[string]any) (*models.TopologyGraph, error) {
	servicesRaw := getSlice(raw, "services")
	if len(servicesRaw) == 0 {
		servicesRaw = getSlicePath(raw, "model", "services")
	}

	linksRaw := getSlice(raw, "links")
	if len(linksRaw) == 0 {
		linksRaw = getSlicePath(raw, "model", "links")
	}

	services := make(map[string]*models.TopologyNode)
	for _, svcAny := range servicesRaw {
		svc, ok := svcAny.(map[string]any)
		if !ok {
			continue
		}
		node := parseNode(svc)
		services[node.ID] = node
	}

	edges := make([]models.TopologyEdge, 0, len(linksRaw))
	for _, linkAny := range linksRaw {
		link, ok := linkAny.(map[string]any)
		if !ok {
			continue
		}
		edges = append(edges, parseEdge(link))
	}

	return &models.TopologyGraph{
		SchemaVersion: stringField(raw, "version", "schema_version", "3.0.0"),
		SourceVersion: stringField(raw, "source_version", "", "unknown"),
		Services:      services,
		Edges:         edges,
	}, nil
}

func parseNode(svc map[string]any) *models.TopologyNode {
	decoy := svc["decoy"]
	if decoy == nil {
		decoy = svc["decoy_profile"]
	}
	decoyMap, _ := decoy.(map[string]any)
	decoyKind := ""
	if decoyMap != nil {
		if k, ok := decoyMap["kind"].(string); ok {
			decoyKind = k
		}
	}

	bindIP := stringOr(svc["ip"], svc["bind_ip"])

	node := &models.TopologyNode{
		ID:                 stringOr(svc["id"], ""),
		OrgID:              stringOr(svc["org_id"], ""),
		Name:               stringOr(svc["name"], stringOr(svc["id"], "")),
		Kind:               stringOr(svc["kind"], ""),
		Exposure:           models.Exposure(stringOr(svc["exposure"], "intranet")),
		Host:               stringOr(svc["host"], ""),
		NetworkID:          stringOr(svc["network_id"], ""),
		BindIP:             bindIP,
		Auth:               stringOr(svc["auth"], "local"),
		DataClassification: stringOr(svc["data_classification"], "internal"),
		Criticality:        models.Criticality(stringOr(svc["criticality"], "medium")),
		Ports:              stringSlice(svc["ports"]),
		IsDecoy:            decoy != nil,
		DecoyKind:          decoyKind,
		Software:           mapAny(svc["software"]),
		OSHint:             stringOr(svc["os_hint"], ""),
	}
	return node
}

func parseEdge(link map[string]any) models.TopologyEdge {
	source := stringOr(link["from"], link["from_service"])
	target := stringOr(link["to"], link["to_service"])
	return models.TopologyEdge{
		Source:     source,
		Target:     target,
		Kind:       stringOr(link["kind"], ""),
		Protocol:   stringOr(link["protocol"], ""),
		Encryption: stringOr(link["encryption"], ""),
	}
}

func stringOr(values ...any) string {
	for _, v := range values {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func stringSlice(v any) []string {
	if s, ok := v.([]string); ok {
		return s
	}
	if arr, ok := v.([]any); ok {
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func mapAny(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func getSlice(raw map[string]any, key string) []any {
	v, ok := raw[key].([]any)
	if !ok {
		return nil
	}
	return v
}

func getSlicePath(raw map[string]any, first, second string) []any {
	m, ok := raw[first].(map[string]any)
	if !ok {
		return nil
	}
	return getSlice(m, second)
}

func stringField(raw map[string]any, keys ...string) string {
	defaultValue := keys[len(keys)-1]
	for _, key := range keys[:len(keys)-1] {
		v := raw[key]
		if v == nil {
			continue
		}
		switch s := v.(type) {
		case string:
			return s
		case int:
			return strconv.Itoa(s)
		case float64:
			return strconv.FormatFloat(s, 'f', -1, 64)
		}
	}
	return defaultValue
}
