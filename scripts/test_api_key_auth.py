"""Unit tests for API key handling in api-check."""

import importlib.util
import json
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT))

from lib.auth import extract_api_key_from_response
from lib.formatters import (
    sanitize_credentials_for_documentation,
    sanitize_curl_for_documentation,
)
from lib.placeholders import replace_auth_placeholders
from lib.types import AuthState, TestDefinition


def load_api_check_module():
    """Load the api-check module from its file path."""
    spec = importlib.util.spec_from_file_location("api_check", ROOT / "api-check.py")
    module = importlib.util.module_from_spec(spec)
    if spec is None or spec.loader is None:
        raise RuntimeError("failed to load api-check module")
    spec.loader.exec_module(module)
    return module


class APIKeyAuthTests(unittest.TestCase):
    """Covers API key parsing, replacement, and documentation redaction."""

    def test_load_test_suite_reads_api_key(self):
        api_check = load_api_check_module()
        payload = {
            "docURL": "http://localhost:6000",
            "serverURL": "http://localhost:6006",
            "apiKey": "moon_live_test",
            "tests": [],
        }

        with tempfile.NamedTemporaryFile("w", suffix=".json", delete=False) as tmp:
            json.dump(payload, tmp)
            tmp_path = tmp.name

        self.addCleanup(lambda: Path(tmp_path).unlink(missing_ok=True))

        suite = api_check.load_test_suite(tmp_path)

        self.assertEqual(suite.api_key, "moon_live_test")

    def test_extract_api_key_from_response_returns_key(self):
        response = {
            "message": "Mutation completed successfully",
            "data": [
                {
                    "id": "01TEST",
                    "key": "moon_live_abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ",
                }
            ],
        }

        api_key = extract_api_key_from_response(response)

        self.assertTrue(api_key.startswith("moon_live_"))

    def test_replace_auth_placeholders_replaces_api_key(self):
        test = TestDefinition(
            name="Use API key",
            cmd="GET",
            endpoint="/collections:query",
            headers={"Authorization": "Bearer $API_KEY"},
            data={"token": "$API_KEY"},
        )
        auth_state = AuthState()
        auth_state.update_api_key("moon_live_example")

        replace_auth_placeholders(test, auth_state)

        self.assertEqual(test.headers["Authorization"], "Bearer moon_live_example")
        self.assertEqual(test.data["token"], "moon_live_example")

    def test_documentation_sanitizers_redact_api_key(self):
        auth_state = AuthState()
        auth_state.update_access_token("jwt-token")
        auth_state.update_refresh_token("refresh-token")
        auth_state.update_api_key("moon_live_example")

        curl = sanitize_curl_for_documentation(
            'curl -H "Authorization: Bearer moon_live_example" http://localhost:6006',
            "http://localhost:6006",
            "http://localhost:6000",
            auth_state,
        )
        body = sanitize_credentials_for_documentation(
            '{"access_token":"jwt-token","refresh_token":"refresh-token","key":"moon_live_example"}',
            auth_state,
        )

        self.assertIn("$API_KEY", curl)
        self.assertNotIn("moon_live_example", curl)
        self.assertIn("$ACCESS_TOKEN", body)
        self.assertIn("$REFRESH_TOKEN", body)
        self.assertIn("$API_KEY", body)


if __name__ == "__main__":
    unittest.main()
