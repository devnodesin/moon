"""Authentication and token management for API testing."""

import requests
import json
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
    Perform login and return authentication state.
    
    Args:
        server_url: Base server URL
        username: Username for login
        password: Password for login
        prefix: URL prefix (e.g. "/api")
        timeout: Request timeout in seconds
        
    Returns:
        AuthState with tokens, or None if login failed
    """
    login_url = f"{server_url}{prefix}/auth:login"
    login_data = {
        "username": username,
        "password": password
    }
    
    try:
        resp = requests.post(
            login_url,
            json=login_data,
            headers={"Content-Type": "application/json"},
            timeout=timeout
        )
        resp.raise_for_status()
        token_json = resp.json()
        
        auth_state = AuthState(
            current_username=username,
            current_password=password
        )
        
        # Tokens are inside the "data" wrapper per SPEC_API.md
        data_obj = token_json.get("data", token_json)
        
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
    endpoint = test.endpoint
    method = test.cmd.upper()
    data = test.data
    
    # Check if this is a password change operation
    if method == "POST" and "/auth:me" in endpoint:
        if isinstance(data, dict):
            # Password change requires both old_password and password fields
            if "old_password" in data and "password" in data:
                return data["password"]
    
    return None


def detect_token_refresh(test: TestDefinition) -> bool:
    """
    Detect if a test performs a token refresh operation.
    
    Args:
        test: Test definition to check
        
    Returns:
        True if this is a token refresh operation
    """
    endpoint = test.endpoint
    method = test.cmd.upper()
    
    return method == "POST" and "/auth:refresh" in endpoint


def detect_login(test: TestDefinition) -> bool:
    """
    Detect if a test performs a login operation.
    
    Args:
        test: Test definition to check
        
    Returns:
        True if this is a login operation
    """
    endpoint = test.endpoint
    method = test.cmd.upper()
    
    return method == "POST" and "/auth:login" in endpoint


def extract_tokens_from_response(
    response_obj: Optional[Dict[str, Any]]
) -> tuple[Optional[str], Optional[str]]:
    """
    Extract access and refresh tokens from a response object.
    
    Args:
        response_obj: Response JSON object
        
    Returns:
        Tuple of (access_token, refresh_token), either can be None
    """
    if not response_obj or not isinstance(response_obj, dict):
        return None, None
    
    # Tokens may be inside "data" wrapper per SPEC_API.md
    data_obj = response_obj.get("data", response_obj)
    if not isinstance(data_obj, dict):
        data_obj = response_obj
    
    access_token = data_obj.get("access_token")
    refresh_token = data_obj.get("refresh_token")
    
    return access_token, refresh_token


def should_relogin_after_test(
    test: TestDefinition,
    response_status: str,
    new_password: Optional[str]
) -> bool:
    """
    Determine if we should re-login after a test completes.
    
    Args:
        test: Test that was executed
        response_status: HTTP response status
        new_password: New password if password was changed
        
    Returns:
        True if re-login is required
    """
    # Re-login only if password was changed and request succeeded
    return (
        new_password is not None
        and response_status.startswith("2")
    )


def relogin_with_new_password(
    server_url: str,
    username: str,
    new_password: str,
    prefix: str = "",
    timeout: int = 10
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
