// Package server implements the MCP server for exposing skills as tools.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/portertech/skills-mcp-server/internal/registry"
	"github.com/portertech/skills-mcp-server/pkg/skill"
)

// Server wraps an MCP server that exposes skills as tools.
type Server struct {
	mcp      *mcp.Server
	registry *registry.Registry
	logger   *slog.Logger
}

// New creates a new skills MCP server.
func New(reg *registry.Registry, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "skills",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{
			Instructions: "This server provides Claude-compatible skills as tools. " +
				"Call a skill tool to receive expert instructions for that task.",
			Logger: logger,
		},
	)

	s := &Server{
		mcp:      mcpServer,
		registry: reg,
		logger:   logger,
	}

	s.registerSkillTools()

	return s
}

// registerSkillTools registers each skill as an MCP tool.
func (s *Server) registerSkillTools() {
	for _, sk := range s.registry.List() {
		s.registerSkillTool(sk)
	}
}

// SkillInput is the input type for skill tools (empty, no arguments needed).
type SkillInput struct{}

// SkillOutput is the output type for skill tools.
type SkillOutput struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Instructions string `json:"instructions"`
	Path         string `json:"path"`
}

// registerSkillTool registers a single skill as an MCP tool.
func (s *Server) registerSkillTool(sk *skill.Skill) {
	toolName := registry.ToolNameForSkill(sk.Name)

	tool := &mcp.Tool{
		Name:        toolName,
		Description: sk.Description,
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest, input SkillInput) (*mcp.CallToolResult, SkillOutput, error) {
		output := SkillOutput{
			Name:         sk.Name,
			Description:  sk.Description,
			Instructions: sk.Instructions,
			Path:         sk.Path,
		}

		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: formatSkillResponse(sk),
				},
			},
		}

		return result, output, nil
	}

	mcp.AddTool(s.mcp, tool, handler)
	s.logger.Debug("registered skill tool", "name", toolName, "skill", sk.Name)
}

// formatSkillResponse formats a skill as a text response.
func formatSkillResponse(sk *skill.Skill) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Skill: %s\n\n", sk.Name))
	sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", sk.Description))
	sb.WriteString("---\n\n")
	sb.WriteString("## Instructions\n\n")
	sb.WriteString(sk.Instructions)

	return sb.String()
}

// Run starts the MCP server with stdio transport.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("starting skills MCP server",
		"skills_count", s.registry.Count(),
		"skills_root", s.registry.Root(),
	)
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

// RunWithTransport starts the MCP server with a custom transport.
// This is primarily useful for testing.
func (s *Server) RunWithTransport(ctx context.Context, transport mcp.Transport) error {
	return s.mcp.Run(ctx, transport)
}
