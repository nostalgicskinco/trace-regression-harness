from pkg.models.assertion import *
from pkg.rules.engine import RuleEngine

engine = RuleEngine()

class TestEngine:
    def test_max_tokens_pass(self, sample_trace):
        a = Assertion(name="tok", assertion_type=AssertionType.MAX_TOKENS, threshold=1000)
        r = engine.evaluate(a, sample_trace)
        assert r.status == AssertionStatus.PASS

    def test_max_tokens_fail(self, sample_trace):
        a = Assertion(name="tok", assertion_type=AssertionType.MAX_TOKENS, threshold=100)
        r = engine.evaluate(a, sample_trace)
        assert r.status == AssertionStatus.FAIL

    def test_max_cost_pass(self, sample_trace):
        a = Assertion(name="cost", assertion_type=AssertionType.MAX_COST, threshold=0.05)
        r = engine.evaluate(a, sample_trace)
        assert r.status == AssertionStatus.PASS

    def test_forbidden_tools_pass(self, sample_trace):
        a = Assertion(name="forbidden", assertion_type=AssertionType.FORBIDDEN_TOOLS, values=["delete_file"])
        r = engine.evaluate(a, sample_trace)
        assert r.status == AssertionStatus.PASS

    def test_forbidden_tools_fail(self, sample_trace):
        a = Assertion(name="forbidden", assertion_type=AssertionType.FORBIDDEN_TOOLS, values=["search"])
        r = engine.evaluate(a, sample_trace)
        assert r.status == AssertionStatus.FAIL

    def test_required_tools(self, sample_trace):
        a = Assertion(name="required", assertion_type=AssertionType.REQUIRED_TOOLS, values=["search"])
        r = engine.evaluate(a, sample_trace)
        assert r.status == AssertionStatus.PASS

    def test_no_errors(self, sample_trace):
        a = Assertion(name="errors", assertion_type=AssertionType.NO_ERRORS)
        r = engine.evaluate(a, sample_trace)
        assert r.status == AssertionStatus.PASS

    def test_model_match(self, sample_trace):
        a = Assertion(name="model", assertion_type=AssertionType.MODEL_MATCH, values=["gpt-4o"])
        r = engine.evaluate(a, sample_trace)
        assert r.status == AssertionStatus.PASS

    def test_custom_expression(self, sample_trace):
        a = Assertion(name="custom", assertion_type=AssertionType.CUSTOM, expression="trace.get('total_tokens', 0) < 1000")
        r = engine.evaluate(a, sample_trace)
        assert r.status == AssertionStatus.PASS
