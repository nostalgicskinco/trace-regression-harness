// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

package policy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nostalgicskinco/trace-regression-harness/pkg/trace"
)

// ToolNeverCalled asserts that a specific tool span name never appears.
type ToolNeverCalled struct {
	ToolName string
	Sev      Severity
}

func (r *ToolNeverCalled) ID() string          { return "TRH-001" }
func (r *ToolNeverCalled) Name() string        { return "tool-never-called" }
func (r *ToolNeverCalled) Description() string { return "Tool must never be called: " + r.ToolName }
func (r *ToolNeverCalled) Severity() Severity  { return r.Sev }

func (r *ToolNeverCalled) Evaluate(td *trace.TraceData) []Violation {
	var violations []Violation
	for _, s := range td.Spans {
		if s.Name == r.ToolName || strings.Contains(s.Name, r.ToolName) {
			violations = append(violations, Violation{
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Severity: r.Sev,
				Message:  fmt.Sprintf("forbidden tool '%s' was called", r.ToolName),
				SpanName: s.Name,
				SpanID:   s.SpanID,
			})
		}
	}
	return violations
}

// MaxSpanCount asserts at most N spans with a given name pattern.
type MaxSpanCount struct {
	Pattern  string
	MaxCount int
	Sev      Severity
}

func (r *MaxSpanCount) ID() string          { return "TRH-002" }
func (r *MaxSpanCount) Name() string        { return "max-span-count" }
func (r *MaxSpanCount) Description() string { return fmt.Sprintf("Max %d spans matching '%s'", r.MaxCount, r.Pattern) }
func (r *MaxSpanCount) Severity() Severity  { return r.Sev }

func (r *MaxSpanCount) Evaluate(td *trace.TraceData) []Violation {
	count := 0
	for _, s := range td.Spans {
		if strings.Contains(s.Name, r.Pattern) {
			count++
		}
	}
	if count > r.MaxCount {
		return []Violation{{
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Severity: r.Sev,
			Message:  fmt.Sprintf("span pattern '%s' appeared %d times (max %d)", r.Pattern, count, r.MaxCount),
		}}
	}
	return nil
}

// TokenBudget asserts that total token usage stays below a threshold.
type TokenBudget struct {
	MaxTokens int64
	TokenKeys []string // attribute keys to check for token counts
	Sev       Severity
}

func (r *TokenBudget) ID() string         { return "TRH-003" }
func (r *TokenBudget) Name() string       { return "token-budget" }
func (r *TokenBudget) Description() string { return fmt.Sprintf("Total tokens must be < %d", r.MaxTokens) }
func (r *TokenBudget) Severity() Severity { return r.Sev }

func (r *TokenBudget) Evaluate(td *trace.TraceData) []Violation {
	keys := r.TokenKeys
	if len(keys) == 0 {
		keys = []string{
			"gen_ai.usage.prompt_tokens",
			"gen_ai.usage.completion_tokens",
			"llm.token_count.prompt",
			"llm.token_count.completion",
		}
	}
	var total int64
	for _, s := range td.Spans {
		for _, key := range keys {
			if v, ok := s.Attributes[key]; ok {
				total += toInt64(v)
			}
		}
	}
	if total > r.MaxTokens {
		return []Violation{{
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Severity: r.Sev,
			Message:  fmt.Sprintf("total tokens %d exceeds budget %d", total, r.MaxTokens),
		}}
	}
	return nil
}

// NoErrors asserts that no spans have error status.
type NoErrors struct {
	Sev Severity
}

func (r *NoErrors) ID() string          { return "TRH-004" }
func (r *NoErrors) Name() string        { return "no-errors" }
func (r *NoErrors) Description() string { return "No spans may have error status" }
func (r *NoErrors) Severity() Severity  { return r.Sev }

func (r *NoErrors) Evaluate(td *trace.TraceData) []Violation {
	var violations []Violation
	for _, s := range td.Spans {
		if s.StatusCode == 2 { // STATUS_CODE_ERROR
			violations = append(violations, Violation{
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Severity: r.Sev,
				Message:  fmt.Sprintf("span '%s' has error status: %s", s.Name, s.StatusMsg),
				SpanName: s.Name,
				SpanID:   s.SpanID,
			})
		}
	}
	return violations
}

// NoSensitiveAttributes asserts that no span carries attributes matching
// patterns that suggest PII or secrets.
type NoSensitiveAttributes struct {
	Patterns []string // regex patterns for attribute keys
	Sev      Severity
}

func (r *NoSensitiveAttributes) ID() string          { return "TRH-005" }
func (r *NoSensitiveAttributes) Name() string        { return "no-sensitive-attrs" }
func (r *NoSensitiveAttributes) Description() string { return "No sensitive attribute patterns in spans" }
func (r *NoSensitiveAttributes) Severity() Severity  { return r.Sev }

func (r *NoSensitiveAttributes) Evaluate(td *trace.TraceData) []Violation {
	patterns := r.Patterns
	if len(patterns) == 0 {
		patterns = []string{
			`(?i)password`,
			`(?i)secret`,
			`(?i)api[_.]?key`,
			`(?i)bearer`,
			`(?i)ssn`,
			`(?i)credit.?card`,
			`(?i)token(?!s$)`, // match "token" but not "tokens"
		}
	}

	var compiled []*regexp.Regexp
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			continue
		}
		compiled = append(compiled, re)
	}

	var violations []Violation
	for _, s := range td.Spans {
		for key := range s.Attributes {
			for _, re := range compiled {
				if re.MatchString(key) {
					violations = append(violations, Violation{
						RuleID:    r.ID(),
						RuleName:  r.Name(),
						Severity:  r.Sev,
						Message:   fmt.Sprintf("sensitive attribute '%s' found", key),
						SpanName:  s.Name,
						SpanID:    s.SpanID,
						Attribute: key,
					})
					break
				}
			}
		}
	}
	return violations
}

// MaxDuration asserts that total trace duration stays below a threshold.
type MaxDuration struct {
	MaxMs float64
	Sev   Severity
}

func (r *MaxDuration) ID() string          { return "TRH-006" }
func (r *MaxDuration) Name() string        { return "max-duration" }
func (r *MaxDuration) Description() string { return fmt.Sprintf("Total duration must be < %.0fms", r.MaxMs) }
func (r *MaxDuration) Severity() Severity  { return r.Sev }

func (r *MaxDuration) Evaluate(td *trace.TraceData) []Violation {
	var total float64
	for _, s := range td.Spans {
		total += s.DurationMs
	}
	if total > r.MaxMs {
		return []Violation{{
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Severity: r.Sev,
			Message:  fmt.Sprintf("total duration %.0fms exceeds max %.0fms", total, r.MaxMs),
		}}
	}
	return nil
}

// RequiredSpan asserts that at least one span with a given name exists.
type RequiredSpan struct {
	SpanName string
	Sev      Severity
}

func (r *RequiredSpan) ID() string          { return "TRH-007" }
func (r *RequiredSpan) Name() string        { return "required-span" }
func (r *RequiredSpan) Description() string { return "Span must exist: " + r.SpanName }
func (r *RequiredSpan) Severity() Severity  { return r.Sev }

func (r *RequiredSpan) Evaluate(td *trace.TraceData) []Violation {
	for _, s := range td.Spans {
		if s.Name == r.SpanName {
			return nil
		}
	}
	return []Violation{{
		RuleID:   r.ID(),
		RuleName: r.Name(),
		Severity: r.Sev,
		Message:  fmt.Sprintf("required span '%s' not found", r.SpanName),
	}}
}

// AttributeEquals asserts that a specific attribute on matching spans has
// an expected string value.
type AttributeEquals struct {
	SpanPattern string
	AttrKey     string
	Expected    string
	Sev         Severity
}

func (r *AttributeEquals) ID() string          { return "TRH-008" }
func (r *AttributeEquals) Name() string        { return "attribute-equals" }
func (r *AttributeEquals) Description() string { return fmt.Sprintf("%s == %s on spans matching '%s'", r.AttrKey, r.Expected, r.SpanPattern) }
func (r *AttributeEquals) Severity() Severity  { return r.Sev }

func (r *AttributeEquals) Evaluate(td *trace.TraceData) []Violation {
	var violations []Violation
	for _, s := range td.Spans {
		if !strings.Contains(s.Name, r.SpanPattern) {
			continue
		}
		v, ok := s.Attributes[r.AttrKey]
		if !ok {
			violations = append(violations, Violation{
				RuleID:    r.ID(),
				RuleName:  r.Name(),
				Severity:  r.Sev,
				Message:   fmt.Sprintf("attribute '%s' missing on span '%s'", r.AttrKey, s.Name),
				SpanName:  s.Name,
				SpanID:    s.SpanID,
				Attribute: r.AttrKey,
			})
			continue
		}
		if fmt.Sprintf("%v", v) != r.Expected {
			violations = append(violations, Violation{
				RuleID:    r.ID(),
				RuleName:  r.Name(),
				Severity:  r.Sev,
				Message:   fmt.Sprintf("attribute '%s' = '%v', expected '%s'", r.AttrKey, v, r.Expected),
				SpanName:  s.Name,
				SpanID:    s.SpanID,
				Attribute: r.AttrKey,
			})
		}
	}
	return violations
}

// helper: convert any numeric-ish value to int64
func toInt64(v any) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case string:
		var n int64
		fmt.Sscanf(val, "%d", &n)
		return n
	default:
		return 0
	}
}
