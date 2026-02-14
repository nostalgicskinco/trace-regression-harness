// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

package policy

import (
	"encoding/json"
	"fmt"
	"os"
)

// PolicyFile is the JSON representation of a policy file.
type PolicyFile struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Rules       []RuleDef  `json:"rules"`
}

// RuleDef is a single rule in a policy file.
type RuleDef struct {
	Type     string            `json:"type"`
	Severity string            `json:"severity,omitempty"`
	Params   map[string]any    `json:"params,omitempty"`
}

// LoadPolicyFile reads a JSON policy file and returns a Policy.
func LoadPolicyFile(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy file: %w", err)
	}
	return ParsePolicyJSON(data)
}

// ParsePolicyJSON parses JSON bytes into a Policy.
func ParsePolicyJSON(data []byte) (*Policy, error) {
	var pf PolicyFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("parse policy JSON: %w", err)
	}

	p := &Policy{
		Name:        pf.Name,
		Description: pf.Description,
	}

	for _, rd := range pf.Rules {
		sev := SeverityError
		if rd.Severity != "" {
			sev = Severity(rd.Severity)
		}

		rule, err := buildRule(rd.Type, sev, rd.Params)
		if err != nil {
			return nil, fmt.Errorf("build rule %s: %w", rd.Type, err)
		}
		p.Rules = append(p.Rules, rule)
	}
	return p, nil
}

func buildRule(typ string, sev Severity, params map[string]any) (Rule, error) {
	switch typ {
	case "tool-never-called":
		name, _ := params["tool_name"].(string)
		if name == "" {
			return nil, fmt.Errorf("tool_name required")
		}
		return &ToolNeverCalled{ToolName: name, Sev: sev}, nil

	case "max-span-count":
		pattern, _ := params["pattern"].(string)
		maxF, _ := params["max_count"].(float64)
		return &MaxSpanCount{Pattern: pattern, MaxCount: int(maxF), Sev: sev}, nil

	case "token-budget":
		maxF, _ := params["max_tokens"].(float64)
		var keys []string
		if ks, ok := params["token_keys"].([]any); ok {
			for _, k := range ks {
				if s, ok := k.(string); ok {
					keys = append(keys, s)
				}
			}
		}
		return &TokenBudget{MaxTokens: int64(maxF), TokenKeys: keys, Sev: sev}, nil

	case "no-errors":
		return &NoErrors{Sev: sev}, nil

	case "no-sensitive-attrs":
		var patterns []string
		if ps, ok := params["patterns"].([]any); ok {
			for _, p := range ps {
				if s, ok := p.(string); ok {
					patterns = append(patterns, s)
				}
			}
		}
		return &NoSensitiveAttributes{Patterns: patterns, Sev: sev}, nil

	case "max-duration":
		maxF, _ := params["max_ms"].(float64)
		return &MaxDuration{MaxMs: maxF, Sev: sev}, nil

	case "required-span":
		name, _ := params["span_name"].(string)
		return &RequiredSpan{SpanName: name, Sev: sev}, nil

	case "attribute-equals":
		return &AttributeEquals{
			SpanPattern: strParam(params, "span_pattern"),
			AttrKey:     strParam(params, "attr_key"),
			Expected:    strParam(params, "expected"),
			Sev:         sev,
		}, nil

	default:
		return nil, fmt.Errorf("unknown rule type: %s", typ)
	}
}

func strParam(params map[string]any, key string) string {
	if v, ok := params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
