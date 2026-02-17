from pkg.evaluator.runner import PolicyRunner

runner = PolicyRunner()

class TestRunner:
    def test_full_check(self, sample_policy, sample_trace):
        result = runner.check(sample_policy, sample_trace)
        assert result.passed is True
        assert result.total == 6
        assert result.failed_count == 0

    def test_batch_check(self, sample_policy, sample_trace):
        results = runner.check_batch(sample_policy, [sample_trace, sample_trace])
        assert len(results) == 2
        assert all(r.passed for r in results)

    def test_failing_trace(self, sample_policy):
        bad_trace = {"id": "bad", "total_tokens": 99999, "total_cost_usd": 999, "model": "unknown", "interactions": []}
        result = runner.check(sample_policy, bad_trace)
        assert result.passed is False
