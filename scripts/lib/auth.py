"""Authentication and token management for API testing."""

import requests
from typing import Optional, Dict, Any

from .types import AuthState, TestDefinition


def perform_login(
    server_url: str,
    username: str,
    password: str,
    prefix: str = "",
    timeout: int = 10
) -> Optional[AuthState]:
    """
    Perform login via POST /auth:session (op=login) and return authentication state.

    Args:
        server_url: Base server URL
        username: Username for login
        password: Password for login
        prefix: URL prefix (e.g. "/api")
        timeout: Request timeout in seconds

    Returns:
        AuthState with tokens, or None if login failed
    """
    login_url = f"{server_url}{prefix}/auth:session"
    login_data = {
        "op": "login",
        "data": {
            "username": username,
            "password": password,
        },
    }

    try:
        resp = requests.post(
            login_url,
            json=login_data,
            headers={"Content-Type": "application/json"},
            timeout=timeout,
        )
        resp.raise_for_status()
        token_json = resp.json()

        auth_state = AuthState(
            current_username=username,
            current_password=password,
        )

        # Response shape: {"message": "...", "data": [{"access_token": ..., "refresh_token": ..., ...}]}
        data_list = token_json.get("data", [])
        data_obj: Dict[str, Any] = {}
        if isinstance(data_list, list) and len(data_list) > 0:
            data_obj = data_list[0] if isinstance(data_list[0], dict) else {}
        elif isinstance(data_list, dict):
            data_obj = data_list

        access_token = data_obj.get("access_token")
        if access_token:
            auth_state.update_access_token(access_token)

        refresh_token = data_obj.get("refresh_token")
        if refresh_token:
            auth_state.update_refresh_token(refresh_token)

        return auth_state
    except Exception:
        return None


def detect_password_change(test: TestDefinition) -> Optional[str]:
    """
    Detect if a test changes the user's password and return the new password.

    Args:
        test: Test definition to check

    Returns:
        New password if detected, None otherwise
    """
    method = test.cmd.upper()
    endpoint = test.endpoint
    data = test.data

    # POST /auth:me with both old_password and password fields is a password change.
    if method == "POST" and "/auth:me" in endpoint:
        if isinstance(data, dict):
            if "old_password" in data and "password" in data:
                return data["password"]

    return None


def detect_token_refresh(test: TestDefinition) -> bool:
    """
    Detect if a test performs a token refresh operation via POST /auth:session op=refresh.

    Args:
        test: Test definition to check

    Returns:
        True if this is a token refresh operation
    """
    method = test.cmd.upper()
    endpoint = test.endpoint
    data = test.data

    if method != "POST" or "/auth:session" not in endpoint:
        return False

    if isinstance(data, dict):
        return data.get("op") == "refresh"

    return False


def detect_login(test: TestDefinition) -> bool:
    """
    Detect if a test performs a login operation via POST /auth:session op=login.

    Args:
        test: Test definition to check

    Returns:
        True if this is a login operation
    """
    method = test.cmd.upper()
    endpoint = test.endpoint
    data = test.data

    if method != "POST" or "/auth:session" not in endpoint:
        return False

    if isinstance(data, dict):
        return data.get("op") == "login"

    return False


def extract_tokens_from_response(
    response_obj: Optional[Dict[str, Any]]
) -> tuple[Optional[str], Optional[str]]:
    """
    Extract access and refresh tokens from a Moon API response object.

    Response shape: {"message": "...", "data": [{"access_token": ..., "refresh_token": ..., ...}]}

    Args:
        response_obj: Response JSON object

    Returns:
        Tuple of (access_token, refresh_token), either can be None
    """
    if not response_obj or not isinstance(response_obj, dict):
        return None, None

    data_raw = response_obj.get("data")
    data_obj: Dict[str, Any] = {}

    if isinstance(data_raw, list) and len(data_raw) > 0:
        data_obj = data_raw[0] if isinstance(data_raw[0], dict) else {}
    elif isinstance(data_raw, dict):
        data_obj = data_raw
    else:
        data_obj = response_obj

    if not isinstance(data_obj, dict):
        return None, None

    access_token = data_obj.get("access_token")
    refresh_token = data_obj.get("refresh_token")

    return access_token, refresh_token


def extract_api_key_from_response(response_obj: Optional[Dict[str, Any]]) -> Optional[str]:
    """
    Extract a raw API key from a Moon API response object.

    API keys are returned only by API key create and rotate responses.

    Args:
        response_obj: Response JSON object

    Returns:
        Raw API key if present, otherwise None
    """
    if not response_obj or not isinstance(response_obj, dict):
        return None

    data_raw = response_obj.get("data")
    data_obj: Dict[str, Any] = {}

    if isinstance(data_raw, list) and len(data_raw) > 0:
        data_obj = data_raw[0] if isinstance(data_raw[0], dict) else {}
    elif isinstance(data_raw, dict):
        data_obj = data_raw
    else:
        data_obj = response_obj

    if not isinstance(data_obj, dict):
        return None

    api_key = data_obj.get("key")
    return api_key if isinstance(api_key, str) else None


def detect_api_key_create_or_rotate(test: TestDefinition) -> bool:
    """
    Detect if a test creates or rotates an API key.

    Args:
        test: Test definition to check

    Returns:
        True if the request should return raw API key material
    """
    method = test.cmd.upper()
    endpoint = test.endpoint
    data = test.data

    if method != "POST" or "/data/apikeys:mutate" not in endpoint:
        return False

    if not isinstance(data, dict):
        return False

    op = data.get("op")
    if op == "create":
        return True

    return op == "action" and data.get("action") == "rotate"


def should_relogin_after_test(
    test: TestDefinition,
    response_status: str,
    new_password: Optional[str],
) -> bool:
    """
    Determine if we should re-login after a test completes.

    Args:
        test: Test that was executed
        response_status: HTTP response status string (e.g. "200 OK")
        new_password: New password if password was changed

    Returns:
        True if re-login is required
    """
    return new_password is not None and response_status.startswith("2")


def relogin_with_new_password(
    server_url: str,
    username: str,
    new_password: str,
    prefix: str = "",
    timeout: int = 10,
) -> Optional[AuthState]:
    """
    Re-login with new password after password change.

    Args:
        server_url: Base server URL
        username: Username for login
        new_password: New password to use
        prefix: URL prefix (e.g. "/api")
        timeout: Request timeout in seconds

    Returns:
        New AuthState with updated tokens, or None if login failed
    """
    return perform_login(server_url, username, new_password, prefix, timeout)
