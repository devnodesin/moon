"""Automated API test runner and result saver."""

import json
import argparse
import os
from typing import List, Optional

from lib.types import TestDefinition, TestSuite, AuthState
from lib.http_client import check_health
from lib.auth import perform_login
from lib.test_runner import run_test_suite


def parse_args() -> argparse.Namespace:
    """Parse command-line arguments for output directory and test files."""
    parser = argparse.ArgumentParser(description="Automated API test runner")
    parser.add_argument(
        '-o', '--outdir',
        default='./out',
        help='Output directory for result files (default: ./out)'
    )
    parser.add_argument(
        '-i', '--input',
        default=None,
        help='Test JSON file to run (default: all in tests dir)'
    )
    parser.add_argument(
        '-t', '--testdir',
        default='./tests',
        help='Directory containing test JSON files (default: ./tests)'
    )
    parser.add_argument(
        '-s', '--server',
        default=None,
        help='Server URL to use for all tests (overrides serverURL in JSON files)'
    )
    return parser.parse_args()


def setup_outdir(outdir: str) -> None:
    """Ensure the output directory exists (creates if missing)."""
    os.makedirs(outdir, exist_ok=True)


def load_test_suite(test_file: str) -> TestSuite:
    """Load test suite from JSON file."""
    with open(test_file, 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    # Convert test dictionaries to TestDefinition objects
    tests = [
        TestDefinition(
            name=t.get("name", ""),
            cmd=t.get("cmd", "GET"),
            endpoint=t.get("endpoint", "/"),
            headers=t.get("headers"),
            data=t.get("data"),
            details=t.get("details"),
            notes=t.get("notes")
        )
        for t in data.get("tests", [])
    ]
    
    return TestSuite(
        docURL=data["docURL"],
        serverURL=data["serverURL"],
        prefix=data.get("prefix", ""),
        username=data.get("username", "admin"),
        password=data.get("password", "moonadmin12#"),
        health=data.get("health", "/health"),
        tests=tests
    )


def check_if_auth_needed(test_suite: TestSuite) -> bool:
    """Check if any test requires authentication token."""
    return any(
        test.headers
        and "Authorization" in test.headers
        and "$ACCESS_TOKEN" in test.headers["Authorization"]
        for test in test_suite.tests
    )


def get_test_files(args: argparse.Namespace) -> List[str]:
    """Get list of test files to process."""
    if args.input:
        return [args.input]
    
    # Find all .json files in testdir, sorted alphabetically for deterministic order
    return sorted([
        os.path.join(args.testdir, f)
        for f in os.listdir(args.testdir)
        if f.endswith('.json')
    ])


def process_test_file(test_file: str, args: argparse.Namespace) -> None:
    """Process a single test file."""
    # Load test suite
    test_suite = load_test_suite(test_file)
    
    # Override serverURL if --server parameter is provided
    if args.server:
        test_suite.serverURL = args.server
    
    # Perform health check
    health_url = f"{test_suite.serverURL}{test_suite.prefix}{test_suite.health}"
    is_healthy, error_msg = check_health(health_url)
    
    if not is_healthy:
        print(f"Skipping {test_file} [server unhealthy: {error_msg}]")
        return
    
    # Check if authentication is needed
    auth_state: Optional[AuthState] = None
    if check_if_auth_needed(test_suite):
        auth_state = perform_login(
            test_suite.serverURL,
            test_suite.username,
            test_suite.password
        )
        if not auth_state or not auth_state.access_token:
            print(f"Login failed for {test_file}")
            return
    
    # Prepare output file
    base = os.path.splitext(os.path.basename(test_file))[0]
    outfilename = os.path.join(args.outdir, f"{base}.md")
    
    # Run test suite
    result = run_test_suite(test_suite, auth_state, outfilename)
    
    # Print results
    print("\n==============================================")
    print(f"Executed {test_file} [{result.status}]")
    print("==============================================\n")
    print(result.markdown)


def main() -> None:
    """Entry point: parse arguments, ensure output directory, and run all tests."""
    args = parse_args()
    setup_outdir(args.outdir)
    
    test_files = get_test_files(args)
    
    for test_file in test_files:
        try:
            process_test_file(test_file, args)
        except Exception as e:
            print(f"Error processing {test_file}: {e}")


if __name__ == "__main__":
    main()
