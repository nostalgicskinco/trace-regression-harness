// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command tracecheck evaluates trace-based regression policies against
// OTLP JSON trace exports.
//
// Usage:
//
//	tracecheck -trace traces.json -policy safety.json [-format text|json]
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nostalgicskinco/trace-regression-harness/pkg/policy"
	"github.com/nostalgicskinco/trace-regression-harness/pkg/report"
	"github.com/nostalgicskinco/trace-regression-harness/pkg/trace"
)

func main() {
	traceFile := flag.String("trace", "", "Path to OTLP JSON trace export")
	policyFile := flag.String("policy", "", "Path to policy JSON file")
	format := flag.String("format", "text", "Output format: text or json")
	flag.Parse()

	if *traceFile == "" || *policyFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: tracecheck -trace <traces.json> -policy <policy.json>\n")
		os.Exit(1)
	}

	td, err := trace.LoadFile(*traceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading trace: %v\n", err)
		os.Exit(1)
	}

	pol, err := policy.LoadPolicyFile(*policyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading policy: %v\n", err)
		os.Exit(1)
	}

	engine := policy.NewEngine()
	engine.AddPolicy(pol)
	results := engine.EvaluateAll(td)

	switch *format {
	case "json":
		if err := report.WriteJSON(os.Stdout, results); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON: %v\n", err)
			os.Exit(1)
		}
	default:
		report.WriteText(os.Stdout, results)
	}

	if policy.HasErrors(results) {
		os.Exit(2)
	}
}
