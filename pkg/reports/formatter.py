"""Report formatter — renders check results for display."""

from __future__ import annotations

import json

from pkg.models.assertion import AssertionStatus, CheckResult


class ReportFormatter:
    """Formats check results into readable reports."""

    def to_summary(self, result: CheckResult) -> str:
        status = "PASS" if result.passed else "FAIL"
        lines = [
            f"Policy: {result.policy_name}",
            f"Status: {status}",
            f"Total: {result.total} | Passed: {result.passed_count} | Failed: {result.failed_count} | Warned: {result.warned_count}",
        ]
        return "\n".join(lines)

    def to_detail(self, result: CheckResult) -> str:
        lines = [self.to_summary(result), "---"]
        for r in result.results:
            icon = {"pass": "✅", "fail": "❌", "warn": "⚠️", "skip": "⏭️"}.get(r.status.value, "?")
            lines.append(f"{icon} [{r.severity.value.upper()}] {r.assertion_name}: {r.message}")
        return "\n".join(lines)

    def to_json(self, result: CheckResult) -> str:
        return json.dumps(result.model_dump(mode="json"), indent=2)

    def batch_summary(self, results: list[CheckResult]) -> str:
        total = len(results)
        passed = sum(1 for r in results if r.passed)
        failed = total - passed
        lines = [
            f"Batch: {total} traces checked",
            f"Passed: {passed} | Failed: {failed}",
        ]
        if failed > 0:
            lines.append("Failed traces:")
            for r in results:
                if not r.passed:
                    lines.append(f"  - {r.trace_id}: {r.failed_count} failures")
        return "\n".join(lines)
