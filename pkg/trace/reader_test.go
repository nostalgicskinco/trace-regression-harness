// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

package trace

import (
	"testing"
)

const sampleOTLP = `{
	"resourceSpans": [{
		"scopeSpans": [{
			"spans": [
				{
					"traceId": "abc123",
					"spanId": "span-1",
					"name": "chat gpt-4",
					"kind": 3,
					"startTimeUnixNano": "1700000000000000000",
					"endTimeUnixNano": "1700000000500000000",
					"status": {"code": 0},
					"attributes": [
						{"key": "gen_ai.request.model", "value": {"stringValue": "gpt-4"}},
						{"key": "gen_ai.usage.prompt_tokens", "value": {"intValue": "100"}}
					],
					"events": [
						{
							"name": "gen_ai.content.prompt",
							"attributes": [
								{"key": "gen_ai.prompt", "value": {"stringValue": "hello"}}
							]
						}
					]
				},
				{
					"traceId": "abc123",
					"spanId": "span-2",
					"parentSpanId": "span-1",
					"name": "tool web_search",
					"kind": 1,
					"startTimeUnixNano": "1700000000100000000",
					"endTimeUnixNano": "1700000000200000000",
					"status": {"code": 2, "message": "not found"}
				}
			]
		}]
	}]
}`

func TestParse(t *testing.T) {
	td, err := Parse([]byte(sampleOTLP))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(td.Spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(td.Spans))
	}

	s := td.Spans[0]
	if s.Name != "chat gpt-4" {
		t.Fatalf("expected 'chat gpt-4', got '%s'", s.Name)
	}
	if s.TraceID != "abc123" {
		t.Fatalf("expected traceId 'abc123', got '%s'", s.TraceID)
	}
	if s.DurationMs != 500 {
		t.Fatalf("expected 500ms, got %.0fms", s.DurationMs)
	}
	model, ok := s.Attributes["gen_ai.request.model"]
	if !ok || model != "gpt-4" {
		t.Fatalf("expected model gpt-4, got %v", model)
	}
	if len(s.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(s.Events))
	}
}

func TestParseErrorSpan(t *testing.T) {
	td, err := Parse([]byte(sampleOTLP))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	s := td.Spans[1]
	if s.StatusCode != 2 {
		t.Fatalf("expected status code 2, got %d", s.StatusCode)
	}
	if s.StatusMsg != "not found" {
		t.Fatalf("expected 'not found', got '%s'", s.StatusMsg)
	}
	if s.ParentID != "span-1" {
		t.Fatalf("expected parent span-1, got '%s'", s.ParentID)
	}
}

func TestSpansWithName(t *testing.T) {
	td, _ := Parse([]byte(sampleOTLP))
	matches := td.SpansWithName("tool web_search")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestSpansWithAttribute(t *testing.T) {
	td, _ := Parse([]byte(sampleOTLP))
	matches := td.SpansWithAttribute("gen_ai.request.model")
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}
