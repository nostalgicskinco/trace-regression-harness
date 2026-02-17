from pkg.models.assertion import *

class TestModels:
    def test_assertion_create(self):
        a = Assertion(name="test", assertion_type=AssertionType.MAX_TOKENS, threshold=100)
        assert a.severity == Severity.ERROR

    def test_policy_spec(self):
        p = PolicySpec(name="test", assertions=[
            Assertion(name="a1", assertion_type=AssertionType.NO_ERRORS),
        ])
        assert len(p.assertions) == 1

    def test_check_result_tracking(self):
        r = CheckResult(policy_name="test")
        r.add_result(AssertionResult(assertion_name="a1", assertion_type=AssertionType.NO_ERRORS, status=AssertionStatus.PASS, severity=Severity.ERROR))
        r.add_result(AssertionResult(assertion_name="a2", assertion_type=AssertionType.MAX_COST, status=AssertionStatus.FAIL, severity=Severity.ERROR))
        assert r.passed is False
        assert r.passed_count == 1
        assert r.failed_count == 1

    def test_warning_doesnt_fail(self):
        r = CheckResult(policy_name="test")
        r.add_result(AssertionResult(assertion_name="a1", assertion_type=AssertionType.MAX_COST, status=AssertionStatus.FAIL, severity=Severity.WARNING))
        assert r.passed is True  # Warnings don't fail the check
