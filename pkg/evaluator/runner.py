"""Runner — executes a full policy spec against trace data."""

from __future__ import annotations

import json
from pathlib import Path

import yaml

from pkg.models.assertion import CheckResult, PolicySpec
from pkg.rules.engine import RuleEngine


class PolicyRunner:
    """Runs a complete policy check against trace data."""

    def __init__(self) -> None:
        self.engine = RuleEngine()

    def check(self, policy: PolicySpec, trace: dict) -> CheckResult:
        """Run all assertions in a policy against a trace."""
        result = CheckResult(
            policy_name=policy.name,
            trace_id=trace.get("id", ""),
        )
        for assertion in policy.assertions:
            assertion_result = self.engine.evaluate(assertion, trace)
            result.add_result(assertion_result)
        return result

    def check_batch(self, policy: PolicySpec, traces: list[dict]) -> list[CheckResult]:
        """Run policy against multiple traces."""
        return [self.check(policy, trace) for trace in traces]

    @staticmethod
    def load_policy_yaml(path: str) -> PolicySpec:
        """Load a policy spec from a YAML file."""
        content = Path(path).read_text()
        data = yaml.safe_load(content)
        return PolicySpec(**data)

    @staticmethod
    def load_policy_json(path: str) -> PolicySpec:
        """Load a policy spec from a JSON file."""
        content = Path(path).read_text()
        data = json.loads(content)
        return PolicySpec(**data)
