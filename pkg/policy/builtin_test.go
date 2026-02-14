// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

package policy

import (
	"testing"

	"github.com/nostalgicskinco/trace-regression-harness/pkg/trace"
)

func makeTestTraceData() *trace.TraceData {
	return &trace.TraceData{
		Spans: []trace.Span{
			{
				Name:       "chat gpt-4",
				SpanID:     "span-1",
				StatusCode: 0,
				DurationMs: 500,
				Attributes: map[string]any{
					"gen_ai.request.model":            "gpt-4",
					"gen_ai.system":                   "openai",
					"gen_ai.usage.prompt_tokens":      float64(100),
					"gen_ai.usage.completion_tokens":   float64(50),
				},
			},
			{
				Name:       "tool exec_sql",
				SpanID:     "span-2",
				StatusCode: 0,
				DurationMs: 200,
				Attributes: map[string]any{
					"tool.name": "exec_sql",
				},
			},
			{
				Name:       "tool web_search",
				SpanID:     "span-3",
				StatusCode: 2,
				StatusMsg:  "timeout",
				DurationMs: 3000,
				Attributes: map[string]any{
					"tool.name": "web_search",
				},
			},
			{
				Name:       "chat gpt-4",
				SpanID:     "span-4",
				StatusCode: 0,
				DurationMs: 400,
				Attributes: map[string]any{
					"gen_ai.request.model":            "gpt-4",
					"gen_ai.usage.prompt_tokens":      float64(200),
					"gen_ai.usage.completion_tokens":   float64(100),
				},
			},
		},
	}
}

func TestToolNeverCalled(t *testing.T) {
	td := makeTestTraceData()

	// exec_sql IS present → should violate
	rule := &ToolNeverCalled{ToolName: "exec_sql", Sev: SeverityError}
	violations := rule.Evaluate(td)
	if len(violations) == 0 {
		t.Fatal("expected violation for exec_sql")
	}

	// delete_all is NOT present → no violation
	rule2 := &ToolNeverCalled{ToolName: "delete_all", Sev: SeverityError}
	violations2 := rule2.Evaluate(td)
	if len(violations2) != 0 {
		t.Fatalf("expected no violation for delete_all, got %d", len(violations2))
	}
}

func TestMaxSpanCount(t *testing.T) {
	td := makeTestTraceData()

	// "chat" appears 2 times, max 1 → violate
	rule := &MaxSpanCount{Pattern: "chat", MaxCount: 1, Sev: SeverityError}
	violations := rule.Evaluate(td)
	if len(violations) == 0 {
		t.Fatal("expected violation for chat count > 1")
	}

	// "chat" appears 2 times, max 5 → pass
	rule2 := &MaxSpanCount{Pattern: "chat", MaxCount: 5, Sev: SeverityError}
	violations2 := rule2.Evaluate(td)
	if len(violations2) != 0 {
		t.Fatal("expected no violation for chat count <= 5")
	}
}

func TestTokenBudget(t *testing.T) {
	td := makeTestTraceData()

	// Total tokens = 100+50+200+100 = 450; budget 400 → violate
	rule := &TokenBudget{MaxTokens: 400, Sev: SeverityError}
	violations := rule.Evaluate(td)
	if len(violations) == 0 {
		t.Fatal("expected violation for token budget")
	}

	// Budget 1000 → pass
	rule2 := &TokenBudget{MaxTokens: 1000, Sev: SeverityError}
	violations2 := rule2.Evaluate(td)
	if len(violations2) != 0 {
		t.Fatal("expected no violation with large budget")
	}
}

func TestNoErrors(t *testing.T) {
	td := makeTestTraceData()

	// span-3 has error status → should violate
	rule := &NoErrors{Sev: SeverityError}
	violations := rule.Evaluate(td)
	if len(violations) == 0 {
		t.Fatal("expected violation for error spans")
	}
	if violations[0].SpanName != "tool web_search" {
		t.Fatalf("expected web_search span, got %s", violations[0].SpanName)
	}
}

func TestNoSensitiveAttributes(t *testing.T) {
	td := &trace.TraceData{
		Spans: []trace.Span{
			{
				Name:   "operation",
				SpanID: "span-x",
				Attributes: map[string]any{
					"user.password":   "secret123",
					"gen_ai.model":    "gpt-4",
				},
			},
		},
	}

	rule := &NoSensitiveAttributes{Sev: SeverityError}
	violations := rule.Evaluate(td)
	if len(violations) == 0 {
		t.Fatal("expected violation for password attribute")
	}
}

func TestMaxDuration(t *testing.T) {
	td := makeTestTraceData()

	// Total duration = 500+200+3000+400 = 4100ms; max 3000 → violate
	rule := &MaxDuration{MaxMs: 3000, Sev: SeverityError}
	violations := rule.Evaluate(td)
	if len(violations) == 0 {
		t.Fatal("expected violation for duration")
	}

	// Max 10000 → pass
	rule2 := &MaxDuration{MaxMs: 10000, Sev: SeverityWarning}
	violations2 := rule2.Evaluate(td)
	if len(violations2) != 0 {
		t.Fatal("expected no violation for large max")
	}
}

func TestRequiredSpan(t *testing.T) {
	td := makeTestTraceData()

	// "chat gpt-4" exists → pass
	rule := &RequiredSpan{SpanName: "chat gpt-4", Sev: SeverityError}
	violations := rule.Evaluate(td)
	if len(violations) != 0 {
		t.Fatal("expected no violation, span exists")
	}

	// "summarize" does NOT exist → violate
	rule2 := &RequiredSpan{SpanName: "summarize", Sev: SeverityError}
	violations2 := rule2.Evaluate(td)
	if len(violations2) == 0 {
		t.Fatal("expected violation for missing span")
	}
}

func TestAttributeEquals(t *testing.T) {
	td := makeTestTraceData()

	// gpt-4 model on chat spans → pass
	rule := &AttributeEquals{
		SpanPattern: "chat",
		AttrKey:     "gen_ai.request.model",
		Expected:    "gpt-4",
		Sev:         SeverityError,
	}
	violations := rule.Evaluate(td)
	if len(violations) != 0 {
		t.Fatalf("expected no violation, got %d", len(violations))
	}

	// wrong expected value → violate
	rule2 := &AttributeEquals{
		SpanPattern: "chat",
		AttrKey:     "gen_ai.request.model",
		Expected:    "gpt-3.5-turbo",
		Sev:         SeverityError,
	}
	violations2 := rule2.Evaluate(td)
	if len(violations2) == 0 {
		t.Fatal("expected violation for wrong model")
	}
}

func TestPolicyEvaluation(t *testing.T) {
	td := makeTestTraceData()
	p := &Policy{
		Name: "test-policy",
		Rules: []Rule{
			&NoErrors{Sev: SeverityError},
			&TokenBudget{MaxTokens: 1000, Sev: SeverityWarning},
		},
	}
	result := p.Evaluate(td)
	if result.ErrorCount == 0 {
		t.Fatal("expected errors from NoErrors rule")
	}
	if result.Passed() {
		t.Fatal("expected policy to fail")
	}
}

func TestEngineEvaluateAll(t *testing.T) {
	td := makeTestTraceData()
	engine := NewEngine()
	engine.AddPolicy(&Policy{
		Name:  "safety",
		Rules: []Rule{&ToolNeverCalled{ToolName: "rm_rf", Sev: SeverityError}},
	})
	engine.AddPolicy(&Policy{
		Name:  "budget",
		Rules: []Rule{&TokenBudget{MaxTokens: 10000, Sev: SeverityError}},
	})

	results := engine.EvaluateAll(td)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if HasErrors(results) {
		t.Fatal("expected no errors")
	}
}
