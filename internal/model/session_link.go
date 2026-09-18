package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

const (
	SessionLinkHookArtifactAlias = "hook_artifact_alias"
	SessionLinkRotatedArtifact   = "rotated_artifact"
	SessionLinkParentSubagent    = "parent_subagent"
)

// SessionIdentity names one source-defined session identity. Namespace
// distinguishes identities that may use the same opaque value.
type SessionIdentity struct {
	Namespace string `json:"namespace"`
	SessionID string `json:"session_id"`
}

// SessionLink records an explicit relationship reported by an agent source.
// It never represents a timestamp- or content-inferred join.
type SessionLink struct {
	SchemaVersion string          `json:"schema_version"`
	LinkID        string          `json:"link_id"`
	SourceAgent   string          `json:"source_agent"`
	Left          SessionIdentity `json:"left"`
	Right         SessionIdentity `json:"right"`
	Relationship  string          `json:"relationship"`
	Confidence    string          `json:"confidence"`
	SourceRefs    []string        `json:"source_refs"`
}

// NewSessionLink validates and assigns a replay-stable logical ID.
func NewSessionLink(
	sourceAgent string,
	left SessionIdentity,
	right SessionIdentity,
	relationship string,
	sourceRefs []string,
) (SessionLink, error) {
	value := SessionLink{
		SchemaVersion: SchemaVersion,
		SourceAgent:   strings.TrimSpace(sourceAgent),
		Left: SessionIdentity{
			Namespace: strings.TrimSpace(left.Namespace),
			SessionID: strings.TrimSpace(left.SessionID),
		},
		Right: SessionIdentity{
			Namespace: strings.TrimSpace(right.Namespace),
			SessionID: strings.TrimSpace(right.SessionID),
		},
		Relationship: strings.TrimSpace(relationship),
		Confidence:   ConfidenceHigh,
	}
	seen := make(map[string]bool)
	for _, ref := range sourceRefs {
		ref = strings.TrimSpace(ref)
		if ref == "" || seen[ref] {
			continue
		}
		seen[ref] = true
		value.SourceRefs = append(value.SourceRefs, ref)
	}
	sort.Strings(value.SourceRefs)
	digest := sha256.Sum256([]byte(strings.Join([]string{
		value.SchemaVersion,
		value.SourceAgent,
		value.Left.Namespace,
		value.Left.SessionID,
		value.Right.Namespace,
		value.Right.SessionID,
		value.Relationship,
	}, "\x00")))
	value.LinkID = "sl-" + hex.EncodeToString(digest[:16])
	if err := value.Validate(); err != nil {
		return SessionLink{}, err
	}
	return value, nil
}

func (s SessionLink) Validate() error {
	if s.SchemaVersion != SchemaVersion {
		return fmt.Errorf("session link %s: invalid schema_version %q", s.LinkID, s.SchemaVersion)
	}
	if !IsValidSourceAgent(s.SourceAgent) {
		return fmt.Errorf("session link %s: invalid source_agent %q", s.LinkID, s.SourceAgent)
	}
	if s.LinkID == "" || s.Left.Namespace == "" || s.Left.SessionID == "" ||
		s.Right.Namespace == "" || s.Right.SessionID == "" {
		return fmt.Errorf("session link: incomplete identity")
	}
	if s.Left == s.Right {
		return fmt.Errorf("session link %s: identical endpoints", s.LinkID)
	}
	switch s.Relationship {
	case SessionLinkHookArtifactAlias,
		SessionLinkRotatedArtifact,
		SessionLinkParentSubagent:
	default:
		return fmt.Errorf("session link %s: invalid relationship %q", s.LinkID, s.Relationship)
	}
	if s.Confidence != ConfidenceHigh {
		return fmt.Errorf("session link %s: confidence must be high", s.LinkID)
	}
	if len(s.SourceRefs) == 0 {
		return fmt.Errorf("session link %s: source_refs required", s.LinkID)
	}
	for _, ref := range s.SourceRefs {
		if strings.TrimSpace(ref) == "" {
			return fmt.Errorf("session link %s: empty source_ref", s.LinkID)
		}
	}
	return nil
}
