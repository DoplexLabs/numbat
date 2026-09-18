package hook

import (
	"testing"

	"github.com/perplexityai/numbat/internal/model"
)

func TestSessionLinksUsesExplicitArtifactPath(t *testing.T) {
	links := SessionLinks(
		AgentClaude,
		model.AgentClaudeCode,
		map[string]any{
			"session_id":      "hook-session",
			"transcript_path": "/home/dev/.claude/projects/p/artifact-session.jsonl",
		},
	)
	if len(links) != 1 {
		t.Fatalf("links = %+v, want one", links)
	}
	link := links[0]
	if link.Relationship != model.SessionLinkHookArtifactAlias ||
		link.Left.SessionID != "hook-session" ||
		link.Right.SessionID != "artifact-session" {
		t.Fatalf("link = %+v", link)
	}

	if links := SessionLinks(
		AgentClaude,
		model.AgentClaudeCode,
		map[string]any{"session_id": "same", "transcript_path": "/p/same.jsonl"},
	); len(links) != 0 {
		t.Fatalf("redundant links = %+v", links)
	}
}
