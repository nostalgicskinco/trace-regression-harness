// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package report formats policy evaluation results for CLI and CI output.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/nostalgicskinco/trace-regression-harness/pkg/policy"
)

// WriteText writes human-readable results to w.
func WriteText(w io.Writer, results []*policy.Result) {
	for _, r := range results {
		fmt.Fprintf(w, "Policy: %s\n", r.PolicyName)
		if len(r.Violations) == 0 {
			fmt.Fprintf(w, "  ✓ All checks passed\n\n")
			continue
		}
		for _, v := range r.Violations {
			icon := "✗"
			if v.Severity == policy.SeverityWarning {
				icon = "⚠"
			} else if v.Severity == policy.SeverityInfo {
				icon = "ℹ"
			}
			fmt.Fprintf(w, "  %s [%s] %s: %s\n", icon, strings.ToUpper(string(v.Severity)), v.RuleID, v.Message)
			if v.SpanName != "" {
				fmt.Fprintf(w, "    span: %s\n", v.SpanName)
			}
			if v.Attribute != "" {
				fmt.Fprintf(w, "    attribute: %s\n", v.Attribute)
			}
		}
		fmt.Fprintf(w, "  Summary: %d errors, %d warnings\n\n", r.ErrorCount, r.WarnCount)
	}
}

// WriteJSON writes JSON results to w.
func WriteJSON(w io.Writer, results []*policy.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

// Summary returns a one-line summary of all results.
func Summary(results []*policy.Result) string {
	totalErrors := 0
	totalWarns := 0
	for _, r := range results {
		totalErrors += r.ErrorCount
		totalWarns += r.WarnCount
	}
	if totalErrors == 0 && totalWarns == 0 {
		return fmt.Sprintf("All %d policies passed", len(results))
	}
	return fmt.Sprintf("%d policies checked: %d errors, %d warnings", len(results), totalErrors, totalWarns)
}
