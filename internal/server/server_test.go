package server

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/portertech/skills-mcp-server/internal/registry"
	pkgskill "github.com/portertech/skills-mcp-server/pkg/skill"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	content := `---
name: test-skill
description: A test skill
---

Test instructions.
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write skill: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := registry.NewRegistry(tmpDir, logger)
	if err := reg.Scan(); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	srv := New(reg, logger)
	if srv == nil {
		t.Fatal("New() returned nil")
	}
	if srv.mcp == nil {
		t.Error("Server.mcp is nil")
	}
	if srv.registry != reg {
		t.Error("Server.registry not set correctly")
	}
}

func TestNewWithNilLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := registry.NewRegistry(tmpDir, logger)

	srv := New(reg, nil)
	if srv == nil {
		t.Fatal("New() returned nil")
	}
	if srv.logger == nil {
		t.Error("Server.logger should default when nil")
	}
}

func TestFormatSkillResponse(t *testing.T) {
	sk := &pkgskill.Skill{
		Name:         "test-skill",
		Description:  "A test skill",
		Instructions: "Do the thing.",
		Path:         "/path/to/skill",
	}

	response := formatSkillResponse(sk)

	if !strings.Contains(response, "# Skill: test-skill") {
		t.Error("response missing skill name header")
	}
	if !strings.Contains(response, "**Description:** A test skill") {
		t.Error("response missing description")
	}
	if !strings.Contains(response, "Do the thing.") {
		t.Error("response missing instruction content")
	}
}

func TestIntegration(t *testing.T) {
	// Create test skills
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "greet")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	content := `---
name: greet
description: Greeting instructions
---

Say hello politely.
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write skill: %v", err)
	}

	// Set up registry and server
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := registry.NewRegistry(tmpDir, logger)
	if err := reg.Scan(); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	srv := New(reg, logger)

	// Create in-memory transports for testing
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.RunWithTransport(ctx, serverTransport)
	}()

	// Connect client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error: %v", err)
	}
	defer session.Close()

	// List tools
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error: %v", err)
	}

	if len(tools.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools.Tools))
	}

	if tools.Tools[0].Name != "greet" {
		t.Errorf("expected tool name 'greet', got %q", tools.Tools[0].Name)
	}

	// Call the tool
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "greet",
	})
	if err != nil {
		t.Fatalf("CallTool() error: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "# Skill: greet") {
		t.Error("response missing skill header")
	}
	if !strings.Contains(textContent.Text, "Say hello politely.") {
		t.Error("response missing instructions")
	}

	cancel()
}

func TestIntegrationMultipleSkills(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple skills
	skills := []struct {
		name        string
		description string
		content     string
	}{
		{"alpha", "First skill", "Alpha instructions."},
		{"beta", "Second skill", "Beta instructions."},
		{"gamma", "Third skill", "Gamma instructions."},
	}

	for _, s := range skills {
		dir := filepath.Join(tmpDir, s.name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		content := "---\nname: " + s.name + "\ndescription: " + s.description + "\n---\n\n" + s.content + "\n"
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write skill: %v", err)
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	reg := registry.NewRegistry(tmpDir, logger)
	if err := reg.Scan(); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	srv := New(reg, logger)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		srv.RunWithTransport(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer session.Close()

	// Verify all tools are listed
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error: %v", err)
	}

	if len(tools.Tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools.Tools))
	}

	// Verify each tool can be called
	for _, s := range skills {
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: s.name})
		if err != nil {
			t.Errorf("CallTool(%s) error: %v", s.name, err)
			continue
		}

		textContent, ok := result.Content[0].(*mcp.TextContent)
		if !ok {
			t.Errorf("expected TextContent for %s", s.name)
			continue
		}

		if !strings.Contains(textContent.Text, s.content) {
			t.Errorf("tool %s response missing expected content", s.name)
		}
	}

	cancel()
}
