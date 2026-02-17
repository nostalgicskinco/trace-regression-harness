"""Trace regression harness CLI."""

import json

import click
import yaml
from rich.console import Console
from rich.table import Table

from pkg.evaluator.runner import PolicyRunner
from pkg.reports.formatter import ReportFormatter

console = Console()
runner = PolicyRunner()
formatter = ReportFormatter()


@click.group()
def cli():
    """trace-check — Policy assertions for agent traces."""
    pass


@cli.command()
@click.argument("policy_path")
@click.argument("trace_path")
@click.option("--format", "fmt", type=click.Choice(["summary", "detail", "json"]), default="detail")
def check(policy_path: str, trace_path: str, fmt: str):
    """Check a trace against a policy file."""
    if policy_path.endswith(".yaml") or policy_path.endswith(".yml"):
        policy = runner.load_policy_yaml(policy_path)
    else:
        policy = runner.load_policy_json(policy_path)

    with open(trace_path) as f:
        trace = json.load(f)

    result = runner.check(policy, trace)

    if fmt == "summary":
        console.print(formatter.to_summary(result))
    elif fmt == "json":
        console.print(formatter.to_json(result))
    else:
        console.print(formatter.to_detail(result))

    raise SystemExit(0 if result.passed else 1)


@cli.command()
@click.argument("policy_path")
def validate(policy_path: str):
    """Validate a policy file."""
    try:
        if policy_path.endswith(".yaml") or policy_path.endswith(".yml"):
            policy = runner.load_policy_yaml(policy_path)
        else:
            policy = runner.load_policy_json(policy_path)
        console.print(f"[green]Valid policy: {policy.name} ({len(policy.assertions)} assertions)[/green]")
    except Exception as e:
        console.print(f"[red]Invalid policy: {e}[/red]")
        raise SystemExit(1)


if __name__ == "__main__":
    cli()
