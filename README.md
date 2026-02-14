# trace-regression-harness

**Trace-based regression testing for GenAI agents** — assert what your agent *actually did*, not just what it said.

Policy-as-code assertions over OpenTelemetry spans: tool guards, retry limits, token budgets, PII detection, required behaviors. Drop into `go test` or CI with a single binary.

> Part of the **GenAI Infrastructure Standard** — a composable suite of open-source tools for enterprise-grade GenAI observability, security, and governance.
>
> | Layer | Component | Repo |
> |-------|-----------|------|
> | Privacy | Prompt Vault Processor | [prompt-vault-processor](https://github.com/nostalgicskinco/prompt-vault-processor) |
> | Normalization | Semantic Normalizer | [genai-semantic-normalizer](https://github.com/nostalgicskinco/genai-semantic-normalizer) |
> | Metrics | Cost & SLO Pack | [genai-cost-slo](https://github.com/nostalgicskinco/genai-cost-slo) |
> | Replay | Agent VCR | [agent-vcr](https://github.com/nostalgicskinco/agent-vcr) |
> | **Testing** | **Regression Harness** | **this repo** |
> | Security | MCP Scanner | [mcp-security-scanner](https://github.com/nostalgicskinco/mcp-security-scanner) |
> | Safety | GenAI Safe Processor | [otel-processor-genai](https://github.com/nostalgicskinco/opentelemetry-collector-processor-genai) |

## Problem

Prompt-level evals tell you if the output *looks right*. They don't tell you if the agent:
- Called a forbidden tool (`exec_sql`, `delete_file`)
- Retried 47 times burning your token budget
- Leaked PII in span attributes
- Took 30 seconds when the SLO is 5 seconds

**Trace-based assertions** operate on what the agent *actually did* — the spans, tool calls, token counts, and error states recorded by OpenTelemetry.

## Quick Start

```bash
# Build the CLI
go build -o tracecheck ./cmd/tracecheck

# Run assertions against a trace export
./tracecheck -trace traces.json -policy safety.json
```

## Built-in Rules

| Rule | ID | Description |
|------|----|-------------|
| `tool-never-called` | TRH-001 | Assert a specific tool was never invoked |
| `max-span-count` | TRH-002 | Limit how many times a span pattern appears |
| `token-budget` | TRH-003 | Total token usage must stay under threshold |
| `no-errors` | TRH-004 | No spans may have error status |
| `no-sensitive-attrs` | TRH-005 | No PII/secret patterns in attribute keys |
| `max-duration` | TRH-006 | Total trace duration budget |
| `required-span` | TRH-007 | Assert a specific span must exist |
| `attribute-equals` | TRH-008 | Assert attribute values on matching spans |

## Policy Files

Policies are JSON files containing a list of rules:

```json
{
  "name": "agent-safety",
  "rules": [
    {"type": "tool-never-called", "severity": "error", "params": {"tool_name": "exec_sql"}},
    {"type": "no-errors", "severity": "error"},
    {"type": "token-budget", "severity": "warning", "params": {"max_tokens": 5000}},
    {"type": "no-sensitive-attrs", "severity": "error"}
  ]
}
```

## Go Test Integration

```go
func TestAgentTrace(t *testing.T) {
    td, _ := trace.LoadFile("golden-trace.json")
    pol, _ := policy.LoadPolicyFile("safety.json")
    result := pol.Evaluate(td)
    if !result.Passed() {
        for _, v := range result.Violations {
            t.Errorf("[%s] %s: %s", v.RuleID, v.RuleName, v.Message)
        }
    }
}
```

## Output Formats

- **text** — human-readable with icons (default)
- **json** — structured for CI pipelines

Exit code `2` on policy violations, making it CI-friendly.

## License

AGPL-3.0-or-later — see [LICENSE](LICENSE). Commercial licenses available — see [COMMERCIAL_LICENSE.md](COMMERCIAL_LICENSE.md).
