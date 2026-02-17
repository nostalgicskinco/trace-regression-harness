"""Rule engine — evaluates assertions against trace data."""

from __future__ import annotations

from typing import Any

from pkg.models.assertion import (
    Assertion,
    AssertionResult,
    AssertionStatus,
    AssertionType,
    Severity,
)


class RuleEngine:
    """Evaluates individual assertions against trace data."""

    def evaluate(self, assertion: Assertion, trace: dict) -> AssertionResult:
        """Evaluate a single assertion against trace data."""
        evaluators = {
            AssertionType.MAX_TOKENS: self._check_max_tokens,
            AssertionType.MAX_COST: self._check_max_cost,
            AssertionType.MAX_DURATION: self._check_max_duration,
            AssertionType.MAX_TOOL_CALLS: self._check_max_tool_calls,
            AssertionType.MAX_LLM_CALLS: self._check_max_llm_calls,
            AssertionType.REQUIRED_TOOLS: self._check_required_tools,
            AssertionType.FORBIDDEN_TOOLS: self._check_forbidden_tools,
            AssertionType.NO_ERRORS: self._check_no_errors,
            AssertionType.MODEL_MATCH: self._check_model_match,
            AssertionType.CUSTOM: self._check_custom,
        }

        evaluator = evaluators.get(assertion.assertion_type)
        if not evaluator:
            return AssertionResult(
                assertion_name=assertion.name,
                assertion_type=assertion.assertion_type,
                status=AssertionStatus.SKIP,
                severity=assertion.severity,
                message=f"Unknown assertion type: {assertion.assertion_type}",
            )

        return evaluator(assertion, trace)

    def _check_max_tokens(self, a: Assertion, trace: dict) -> AssertionResult:
        actual = trace.get("total_tokens", 0)
        ok = actual <= (a.threshold or 0)
        return AssertionResult(
            assertion_name=a.name, assertion_type=a.assertion_type,
            status=AssertionStatus.PASS if ok else AssertionStatus.FAIL,
            severity=a.severity,
            message=f"Tokens: {actual} {'<=' if ok else '>'} {a.threshold}",
            actual_value=actual, expected_value=a.threshold,
        )

    def _check_max_cost(self, a: Assertion, trace: dict) -> AssertionResult:
        actual = trace.get("total_cost_usd", 0.0)
        ok = actual <= (a.threshold or 0)
        return AssertionResult(
            assertion_name=a.name, assertion_type=a.assertion_type,
            status=AssertionStatus.PASS if ok else AssertionStatus.FAIL,
            severity=a.severity,
            message=f"Cost: ${actual:.4f} {'<=' if ok else '>'} ${a.threshold:.4f}",
            actual_value=actual, expected_value=a.threshold,
        )

    def _check_max_duration(self, a: Assertion, trace: dict) -> AssertionResult:
        actual = trace.get("total_duration_ms", 0.0)
        ok = actual <= (a.threshold or 0)
        return AssertionResult(
            assertion_name=a.name, assertion_type=a.assertion_type,
            status=AssertionStatus.PASS if ok else AssertionStatus.FAIL,
            severity=a.severity,
            message=f"Duration: {actual}ms {'<=' if ok else '>'} {a.threshold}ms",
            actual_value=actual, expected_value=a.threshold,
        )

    def _check_max_tool_calls(self, a: Assertion, trace: dict) -> AssertionResult:
        interactions = trace.get("interactions", [])
        actual = sum(1 for i in interactions if i.get("interaction_type") == "tool_call")
        ok = actual <= (a.threshold or 0)
        return AssertionResult(
            assertion_name=a.name, assertion_type=a.assertion_type,
            status=AssertionStatus.PASS if ok else AssertionStatus.FAIL,
            severity=a.severity,
            message=f"Tool calls: {actual} {'<=' if ok else '>'} {int(a.threshold or 0)}",
            actual_value=actual, expected_value=a.threshold,
        )

    def _check_max_llm_calls(self, a: Assertion, trace: dict) -> AssertionResult:
        interactions = trace.get("interactions", [])
        actual = sum(1 for i in interactions if i.get("interaction_type") == "llm_request")
        ok = actual <= (a.threshold or 0)
        return AssertionResult(
            assertion_name=a.name, assertion_type=a.assertion_type,
            status=AssertionStatus.PASS if ok else AssertionStatus.FAIL,
            severity=a.severity,
            message=f"LLM calls: {actual} {'<=' if ok else '>'} {int(a.threshold or 0)}",
            actual_value=actual, expected_value=a.threshold,
        )

    def _check_required_tools(self, a: Assertion, trace: dict) -> AssertionResult:
        interactions = trace.get("interactions", [])
        used_tools = {i.get("tool_name") for i in interactions if i.get("interaction_type") == "tool_call"}
        missing = [t for t in a.values if t not in used_tools]
        ok = len(missing) == 0
        return AssertionResult(
            assertion_name=a.name, assertion_type=a.assertion_type,
            status=AssertionStatus.PASS if ok else AssertionStatus.FAIL,
            severity=a.severity,
            message=f"Required tools: {'all present' if ok else f'missing {missing}'}",
            actual_value=list(used_tools), expected_value=a.values,
        )

    def _check_forbidden_tools(self, a: Assertion, trace: dict) -> AssertionResult:
        interactions = trace.get("interactions", [])
        used_tools = {i.get("tool_name") for i in interactions if i.get("interaction_type") == "tool_call"}
        violations = [t for t in a.values if t in used_tools]
        ok = len(violations) == 0
        return AssertionResult(
            assertion_name=a.name, assertion_type=a.assertion_type,
            status=AssertionStatus.PASS if ok else AssertionStatus.FAIL,
            severity=a.severity,
            message=f"Forbidden tools: {'none used' if ok else f'violations {violations}'}",
            actual_value=list(used_tools), expected_value=a.values,
        )

    def _check_no_errors(self, a: Assertion, trace: dict) -> AssertionResult:
        interactions = trace.get("interactions", [])
        errors = [i for i in interactions if i.get("interaction_type") == "error"]
        ok = len(errors) == 0
        return AssertionResult(
            assertion_name=a.name, assertion_type=a.assertion_type,
            status=AssertionStatus.PASS if ok else AssertionStatus.FAIL,
            severity=a.severity,
            message=f"Errors: {len(errors)} found" if not ok else "No errors",
            actual_value=len(errors), expected_value=0,
        )

    def _check_model_match(self, a: Assertion, trace: dict) -> AssertionResult:
        actual = trace.get("model", "")
        ok = actual in a.values
        return AssertionResult(
            assertion_name=a.name, assertion_type=a.assertion_type,
            status=AssertionStatus.PASS if ok else AssertionStatus.FAIL,
            severity=a.severity,
            message=f"Model: {actual} {'in' if ok else 'not in'} {a.values}",
            actual_value=actual, expected_value=a.values,
        )

    def _check_custom(self, a: Assertion, trace: dict) -> AssertionResult:
        if not a.expression:
            return AssertionResult(
                assertion_name=a.name, assertion_type=a.assertion_type,
                status=AssertionStatus.SKIP, severity=a.severity,
                message="No expression provided",
            )
        try:
            result = eval(a.expression, {"trace": trace, "__builtins__": {}})
            ok = bool(result)
        except Exception as e:
            return AssertionResult(
                assertion_name=a.name, assertion_type=a.assertion_type,
                status=AssertionStatus.FAIL, severity=a.severity,
                message=f"Expression error: {e}",
            )
        return AssertionResult(
            assertion_name=a.name, assertion_type=a.assertion_type,
            status=AssertionStatus.PASS if ok else AssertionStatus.FAIL,
            severity=a.severity,
            message=f"Custom: {'passed' if ok else 'failed'} — {a.expression}",
            actual_value=result,
        )
