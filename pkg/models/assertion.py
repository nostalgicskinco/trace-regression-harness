"""Assertion models — policy rules that traces must satisfy."""

from __future__ import annotations

from enum import Enum
from typing import Any

from pydantic import BaseModel, Field


class AssertionType(str, Enum):
    MAX_TOKENS = "max_tokens"
    MAX_COST = "max_cost"
    MAX_DURATION = "max_duration"
    MAX_TOOL_CALLS = "max_tool_calls"
    MAX_LLM_CALLS = "max_llm_calls"
    REQUIRED_TOOLS = "required_tools"
    FORBIDDEN_TOOLS = "forbidden_tools"
    NO_ERRORS = "no_errors"
    MODEL_MATCH = "model_match"
    CUSTOM = "custom"


class Severity(str, Enum):
    ERROR = "error"        # Fails the check
    WARNING = "warning"    # Warns but passes
    INFO = "info"          # Informational only


class AssertionStatus(str, Enum):
    PASS = "pass"
    FAIL = "fail"
    WARN = "warn"
    SKIP = "skip"


class Assertion(BaseModel):
    """A single policy assertion rule."""
    name: str
    assertion_type: AssertionType
    severity: Severity = Severity.ERROR
    description: str = ""
    
    # Threshold values (used by max_* assertions)
    threshold: float | None = None
    
    # List values (used by required/forbidden tools, model_match)
    values: list[str] = Field(default_factory=list)
    
    # Custom expression (used by CUSTOM type)
    expression: str | None = None
    
    metadata: dict[str, Any] = Field(default_factory=dict)


class AssertionResult(BaseModel):
    """Result of evaluating a single assertion."""
    assertion_name: str
    assertion_type: AssertionType
    status: AssertionStatus
    severity: Severity
    message: str = ""
    actual_value: Any = None
    expected_value: Any = None


class PolicySpec(BaseModel):
    """A collection of assertions that define a policy."""
    name: str
    version: str = "1.0"
    description: str = ""
    assertions: list[Assertion] = Field(default_factory=list)


class CheckResult(BaseModel):
    """Result of running all assertions against a trace."""
    policy_name: str
    trace_id: str = ""
    passed: bool = True
    total: int = 0
    passed_count: int = 0
    failed_count: int = 0
    warned_count: int = 0
    skipped_count: int = 0
    results: list[AssertionResult] = Field(default_factory=list)

    def add_result(self, result: AssertionResult) -> None:
        self.results.append(result)
        self.total += 1
        if result.status == AssertionStatus.PASS:
            self.passed_count += 1
        elif result.status == AssertionStatus.FAIL:
            self.failed_count += 1
            if result.severity == Severity.ERROR:
                self.passed = False
        elif result.status == AssertionStatus.WARN:
            self.warned_count += 1
        elif result.status == AssertionStatus.SKIP:
            self.skipped_count += 1
