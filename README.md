# Trace Regression Harness

[![CI](https://github.com/airblackbox/trace-regression-harness/actions/workflows/ci.yml/badge.svg)](https://github.com/airblackbox/trace-regression-harness/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/airblackbox/trace-regression-harness/blob/main/LICENSE)
[![Python 3.10+](https://img.shields.io/badge/python-3.10+-3776AB.svg?logo=python&logoColor=white)](https://python.org)


**Policy assertions for agent traces.** Define rules your AI agent runs must satisfy, then check traces against them. The quality gate between "the agent ran" and "the agent ran correctly."

## Quick Start

```bash
pip install -e ".[dev]"

# Check a trace against a policy
trace-check check policy.yaml trace.json

# Validate a policy file
trace-check validate policy.yaml
```

## Policy Format (YAML)

```yaml
name: production-safety
version: "1.0"
assertions:
  - name: token-budget
    assertion_type: max_tokens
    threshold: 5000

  - name: cost-cap
    assertion_type: max_cost
    threshold: 0.10

  - name: no-delete
    assertion_type: forbidden_tools
    values: [delete_file, drop_table]

  - name: must-search
    assertion_type: required_tools
    values: [search]

  - name: clean-run
    assertion_type: no_errors
```

## Assertion Types

| Type | Checks | Threshold/Values |
|---|---|---|
| `max_tokens` | Total tokens ≤ threshold | `threshold: 5000` |
| `max_cost` | Total cost ≤ threshold | `threshold: 0.10` |
| `max_duration` | Duration ≤ threshold (ms) | `threshold: 30000` |
| `max_tool_calls` | Tool call count ≤ threshold | `threshold: 20` |
| `max_llm_calls` | LLM call count ≤ threshold | `threshold: 10` |
| `required_tools` | All listed tools were used | `values: [search]` |
| `forbidden_tools` | None of listed tools were used | `values: [delete]` |
| `no_errors` | No error interactions | — |
| `model_match` | Model is in allowed list | `values: [gpt-4o]` |
| `custom` | Python expression evaluates truthy | `expression: "..."` |


## Part of the AIR Platform

This is one component of the [AIR Blackbox Gateway](https://github.com/airblackbox/gateway) ecosystem.

## License

Apache-2.0
