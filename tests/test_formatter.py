from pkg.models.assertion import *
from pkg.reports.formatter import ReportFormatter

formatter = ReportFormatter()

class TestFormatter:
    def test_summary(self):
        result = CheckResult(policy_name="test", passed=True, total=3, passed_count=3)
        text = formatter.to_summary(result)
        assert "PASS" in text
        assert "test" in text

    def test_detail(self):
        result = CheckResult(policy_name="test")
        result.add_result(AssertionResult(assertion_name="a1", assertion_type=AssertionType.NO_ERRORS, status=AssertionStatus.PASS, severity=Severity.ERROR, message="OK"))
        text = formatter.to_detail(result)
        assert "✅" in text

    def test_json_output(self):
        result = CheckResult(policy_name="test", passed=True)
        text = formatter.to_json(result)
        assert '"policy_name"' in text

    def test_batch_summary(self):
        results = [
            CheckResult(policy_name="p", passed=True, trace_id="t1"),
            CheckResult(policy_name="p", passed=False, trace_id="t2", failed_count=1),
        ]
        text = formatter.batch_summary(results)
        assert "Failed: 1" in text
