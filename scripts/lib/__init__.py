"""API testing framework library."""

from .types import (
    TestDefinition,
    TestSuite,
    TestResponse,
    AuthState,
    PlaceholderContext,
    TestResult
)
from .http_client import execute_request, check_health
from .auth import perform_login
from .test_runner import run_test_suite
from .formatters import format_markdown_result, sanitize_curl_for_documentation

__all__ = [
    "TestDefinition",
    "TestSuite",
    "TestResponse",
    "AuthState",
    "PlaceholderContext",
    "TestResult",
    "execute_request",
    "check_health",
    "perform_login",
    "run_test_suite",
    "format_markdown_result",
    "sanitize_curl_for_documentation",
]
