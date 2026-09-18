package hook

import (
	"path"
	"regexp"
	"strings"

	"github.com/perplexityai/numbat/internal/model"
)

var terminalUUID = regexp.MustCompile(
	`(?i)([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`,
)

// SessionLinks projects only explicit source lineage from a hook payload.
// A transcript path is a source-provided reference to the durable artifact;
// timestamps, cwd, command text, and model names are never used.
func SessionLinks(
	agent string,
	sourceAgent string,
	payload map[string]any,
) []model.SessionLink {
	r := newResolver(agent, payload)
	hookSessionID := r.sessionID()
	artifactPath := r.envStr(
		"transcript_path",
		"transcriptPath",
		"rollout_path",
		"rolloutPath",
		"session_path",
		"sessionPath",
	)
	artifactSessionID := artifactSessionIDFromPath(agent, artifactPath)
	if hookSessionID == "" || artifactSessionID == "" ||
		hookSessionID == artifactSessionID {
		return nil
	}
	link, err := model.NewSessionLink(
		sourceAgent,
		model.SessionIdentity{
			Namespace: "hook",
			SessionID: hookSessionID,
		},
		model.SessionIdentity{
			Namespace: "artifact",
			SessionID: artifactSessionID,
		},
		model.SessionLinkHookArtifactAlias,
		[]string{"hook_payload:artifact_path"},
	)
	if err != nil {
		return nil
	}
	return []model.SessionLink{link}
}

func artifactSessionIDFromPath(agent string, value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), `\`, "/")
	if value == "" {
		return ""
	}
	clean := path.Clean(value)
	switch agent {
	case AgentClaude:
		parts := strings.Split(clean, "/")
		for index, part := range parts {
			if part == "subagents" && index > 0 {
				return parts[index-1]
			}
		}
		return trimSessionArtifactSuffix(path.Base(clean))
	case AgentCodex:
		stem := trimSessionArtifactSuffix(path.Base(clean))
		match := terminalUUID.FindStringSubmatch(stem)
		if len(match) == 2 {
			return strings.ToLower(match[1])
		}
	}
	return ""
}

func trimSessionArtifactSuffix(value string) string {
	value = strings.TrimSuffix(value, ".zst")
	value = strings.TrimSuffix(value, ".jsonl")
	return strings.TrimSpace(value)
}
