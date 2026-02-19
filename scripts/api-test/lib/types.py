"""Type definitions and data classes for API testing framework."""

from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any


@dataclass
class TestDefinition:
    """Represents a single API test case."""
    name: str
    cmd: str
    endpoint: str
    headers: Optional[Dict[str, str]] = None
    data: Optional[Any] = None
    details: Optional[str] = None
    notes: Optional[str] = None
    expected_status: Optional[int] = None


@dataclass
class TestSuite:
    """Represents a complete test suite from a JSON file."""
    docURL: str
    serverURL: str
    prefix: str = ""
    username: str = "admin"
    password: str = "moonadmin12#"
    health: str = "/health"
    tests: List[TestDefinition] = field(default_factory=list)


@dataclass
class TestResponse:
    """Represents the response from a single test execution."""
    curl_command: str
    status: str
    body: str
    response_obj: Optional[Dict[str, Any]] = None


@dataclass
class AuthState:
    """Tracks authentication state throughout test execution."""
    access_token: Optional[str] = None
    refresh_token: Optional[str] = None
    current_username: Optional[str] = None
    current_password: Optional[str] = None
    all_access_tokens: List[str] = field(default_factory=list)
    all_refresh_tokens: List[str] = field(default_factory=list)
    
    def update_access_token(self, token: str) -> None:
        """Update access token and track it for documentation replacement."""
        self.access_token = token
        if token and token not in self.all_access_tokens:
            self.all_access_tokens.append(token)
    
    def update_refresh_token(self, token: str) -> None:
        """Update refresh token and track it for documentation replacement."""
        self.refresh_token = token
        if token and token not in self.all_refresh_tokens:
            self.all_refresh_tokens.append(token)
    
    def update_credentials(self, username: str, password: str) -> None:
        """Update current credentials after password change."""
        self.current_username = username
        self.current_password = password


@dataclass
class PlaceholderContext:
    """Context for placeholder replacements during test execution."""
    captured_record_id: Optional[str] = None
    placeholder_type: Optional[str] = None  # '$ULID' or '$NEXT_CURSOR'
    captured_record_ids: List[str] = field(default_factory=list)
    
    def set_record_id(self, record_id: str, placeholder_type: str) -> None:
        """Set the captured record ID and its placeholder type."""
        self.captured_record_id = record_id
        self.placeholder_type = placeholder_type
    
    def set_record_ids(self, record_ids: List[str]) -> None:
        """Set multiple captured record IDs for numbered placeholders."""
        self.captured_record_ids = record_ids


@dataclass
class TestResult:
    """Aggregated result from test suite execution."""
    status: str  # 'success' or 'failure'
    output_file: Optional[str]
    markdown: str
    all_tests_passed: bool
