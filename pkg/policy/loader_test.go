// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

package policy

import (
	"testing"

	"github.com/nostalgicskinco/trace-regression-harness/pkg/trace"
)

func TestParsePolicyJSON(t *testing.T) {
	data := []byte(`{
		"name": "safety-check",
		"description": "Basic safety policy",
		"rules": [
			{"type": "tool-never-called", "severity": "error", "params": {"tool_name": "exec_sql"}},
			{"type": "no-errors", "severity": "error"},
			{"type": "token-budget", "severity": "warning", "params": {"max_tokens": 5000}},
			{"type": "max-span-count", "params": {"pattern": "retry", "max_count": 3}},
			{"type": "max-duration", "params": {"max_ms": 10000}},
			{"type": "required-span", "params": {"span_name": "chat gpt-4"}},
			{"type": "attribute-equals", "params": {"span_pattern": "chat", "attr_key": "gen_ai.system", "expected": "openai"}}
		]
	}`)

	pol, err := ParsePolicyJSON(data)
	if err != nil {
		t.Fatalf("ParsePolicyJSON: %v", err)
	}
	if pol.Name != "safety-check" {
		t.Fatalf("expected name 'safety-check', got '%s'", pol.Name)
	}
	if len(pol.Rules) != 7 {
		t.Fatalf("expected 7 rules, got %d", len(pol.Rules))
	}
}

func TestParsePolicyJSONEvaluate(t *testing.T) {
	data := []byte(`{
		"name": "integration",
		"rules": [
			{"type": "tool-never-called", "severity": "error", "params": {"tool_name": "dangerous_tool"}},
			{"type": "no-errors"},
			{"type": "token-budget", "severity": "warning", "params": {"max_tokens": 10000}}
		]
	}`)

	pol, err := ParsePolicyJSON(data)
	if err != nil {
		t.Fatalf("ParsePolicyJSON: %v", err)
	}

	td := &trace.TraceData{
		Spans: []trace.Span{
			{
				Name:       "chat gpt-4",
				SpanID:     "s1",
				StatusCode: 0,
				Attributes: map[string]any{
					"gen_ai.usage.prompt_tokens":    float64(50),
					"gen_ai.usage.completion_tokens": float64(30),
				},
			},
		},
	}

	result := pol.Evaluate(td)
	if !result.Passed() {
		t.Fatalf("expected policy to pass, got %d errors", result.ErrorCount)
	}
}

func TestUnknownRuleType(t *testing.T) {
	data := []byte(`{
		"name": "bad",
		"rules": [{"type": "nonexistent-rule"}]
	}`)
	_, err := ParsePolicyJSON(data)
	if err == nil {
		t.Fatal("expected error for unknown rule type")
	}
}
