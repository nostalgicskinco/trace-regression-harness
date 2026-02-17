"""Test fixtures for trace regression harness."""

import pytest
from pkg.models.assertion import (
    Assertion, AssertionType, PolicySpec, Severity,
)


@pytest.fixture
def sample_trace():
    return {
        "id": "trace-001",
        "agent_id": "test-agent",
        "task": "test-task",
        "model": "gpt-4o",
        "total_tokens": 500,
        "total_cost_usd": 0.01,
        "total_duration_ms": 2000,
        "interactions": [
            {"interaction_type": "llm_request", "model": "gpt-4o"},
            {"interaction_type": "llm_response", "model": "gpt-4o", "tokens_used": 200},
            {"interaction_type": "tool_call", "tool_name": "search", "tool_input": {"q": "test"}},
            {"interaction_type": "tool_result", "tool_name": "search", "tool_output": "results"},
            {"interaction_type": "llm_request", "model": "gpt-4o"},
            {"interaction_type": "llm_response", "model": "gpt-4o", "tokens_used": 300},
        ],
    }


@pytest.fixture
def sample_policy():
    return PolicySpec(
        name="test-policy",
        version="1.0",
        assertions=[
            Assertion(name="token-limit", assertion_type=AssertionType.MAX_TOKENS, threshold=1000),
            Assertion(name="cost-limit", assertion_type=AssertionType.MAX_COST, threshold=0.05),
            Assertion(name="no-errors", assertion_type=AssertionType.NO_ERRORS),
            Assertion(name="required-search", assertion_type=AssertionType.REQUIRED_TOOLS, values=["search"]),
            Assertion(name="no-delete", assertion_type=AssertionType.FORBIDDEN_TOOLS, values=["delete_file"], severity=Severity.ERROR),
            Assertion(name="model-check", assertion_type=AssertionType.MODEL_MATCH, values=["gpt-4o", "gpt-4o-mini"]),
        ],
    )
