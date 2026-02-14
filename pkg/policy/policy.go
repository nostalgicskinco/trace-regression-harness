// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package policy defines trace-based assertion rules and an evaluation engine
// that checks OTel spans against declarative policies.
package policy

import (
	"fmt"
	"strings"

	"github.com/nostalgicskinco/trace-regression-harness/pkg/trace"
)

// Severity indicates how a violation should be treated.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Rule is a single assertion that can be evaluated against trace data.
type Rule interface {
	ID() string
	Name() string
	Description() string
	Severity() Severity
	Evaluate(td *trace.TraceData) []Violation
}

// Violation is a single policy violation found in trace data.
type Violation struct {
	RuleID      string   `json:"ruleId"`
	RuleName    string   `json:"ruleName"`
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	SpanName    string   `json:"spanName,omitempty"`
	SpanID      string   `json:"spanId,omitempty"`
	Attribute   string   `json:"attribute,omitempty"`
}

func (v Violation) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s] %s: %s", strings.ToUpper(string(v.Severity)), v.RuleID, v.Message))
	if v.SpanName != "" {
		parts = append(parts, fmt.Sprintf("  span: %s", v.SpanName))
	}
	return strings.Join(parts, "\n")
}

// Policy is a named collection of rules.
type Policy struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Rules       []Rule `json:"rules"`
}

// Result holds all violations from evaluating a policy.
type Result struct {
	PolicyName  string      `json:"policyName"`
	Violations  []Violation `json:"violations"`
	ErrorCount  int         `json:"errorCount"`
	WarnCount   int         `json:"warnCount"`
}

// Passed returns true if there are no error-severity violations.
func (r *Result) Passed() bool {
	return r.ErrorCount == 0
}

// Evaluate runs all rules in the policy against trace data.
func (p *Policy) Evaluate(td *trace.TraceData) *Result {
	r := &Result{PolicyName: p.Name}
	for _, rule := range p.Rules {
		violations := rule.Evaluate(td)
		for _, v := range violations {
			r.Violations = append(r.Violations, v)
			switch v.Severity {
			case SeverityError:
				r.ErrorCount++
			case SeverityWarning:
				r.WarnCount++
			}
		}
	}
	return r
}

// Engine manages multiple policies and evaluates them against trace data.
type Engine struct {
	policies []*Policy
}

// NewEngine creates an empty policy engine.
func NewEngine() *Engine {
	return &Engine{}
}

// AddPolicy adds a policy to the engine.
func (e *Engine) AddPolicy(p *Policy) {
	e.policies = append(e.policies, p)
}

// EvaluateAll runs all policies and returns combined results.
func (e *Engine) EvaluateAll(td *trace.TraceData) []*Result {
	var results []*Result
	for _, p := range e.policies {
		results = append(results, p.Evaluate(td))
	}
	return results
}

// HasErrors returns true if any result has error-severity violations.
func HasErrors(results []*Result) bool {
	for _, r := range results {
		if !r.Passed() {
			return true
		}
	}
	return false
}
