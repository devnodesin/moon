"""HTTP client for executing API requests."""

import requests
import json
from typing import Dict, Optional, Any, Tuple

from .types import TestDefinition, TestResponse


def execute_request(
    base_url: str,
    prefix: str,
    test: TestDefinition,
    timeout: int = 30
) -> TestResponse:
    """
    Execute a single HTTP request based on test definition.
    
    Args:
        base_url: Base URL of the API server
        prefix: URL prefix to prepend to endpoints
        test: Test definition containing method, endpoint, headers, data
        timeout: Request timeout in seconds
        
    Returns:
        TestResponse containing curl command, status, body, and response object
    """
    method = test.cmd.upper()
    url = f"{base_url}{prefix}{test.endpoint}"
    headers = test.headers or {}
    data = test.data
    
    # Build request kwargs
    req_kwargs: Dict[str, Any] = {"timeout": timeout}
    if headers:
        req_kwargs["headers"] = headers
    if data is not None:
        if isinstance(data, (dict, list)):
            req_kwargs["json"] = data
        else:
            req_kwargs["data"] = data
    
    # Build curl command for documentation
    curl_cmd = _build_curl_command(url, method, headers, data)
    
    # Execute request
    response_obj: Optional[Dict[str, Any]] = None
    try:
        resp = requests.request(method, url, **req_kwargs)
        status = f"{resp.status_code} {resp.reason}"
        try:
            body = resp.json()
            response_obj = body
            body_str = json.dumps(body, indent=2)
        except Exception:
            body_str = resp.text
    except Exception as e:
        status = "ERROR"
        body_str = str(e)
    
    return TestResponse(
        curl_command=curl_cmd,
        status=status,
        body=body_str,
        response_obj=response_obj
    )


def check_health(url: str, timeout: int = 5) -> Tuple[bool, Optional[str]]:
    """
    Check if the API server is healthy.
    
    Args:
        url: Full health check URL
        timeout: Request timeout in seconds
        
    Returns:
        Tuple of (is_healthy, error_message)
    """
    try:
        resp = requests.get(url, timeout=timeout)
        if resp.status_code == 200:
            return True, None
        return False, f"Status {resp.status_code}"
    except Exception as e:
        return False, str(e)


def _build_curl_command(
    url: str,
    method: str,
    headers: Optional[Dict[str, str]],
    data: Optional[Any]
) -> str:
    """
    Build a formatted curl command with proper line continuations and indentation.
    
    Args:
        url: Full request URL
        method: HTTP method
        headers: Request headers
        data: Request body data
        
    Returns:
        Formatted curl command string
    """
    curl_lines = [f'curl -s -X {method} "{url}"']
    
    if headers:
        for k, v in headers.items():
            curl_lines.append(f'    -H "{k}: {v}"')
    
    if data is not None:
        if isinstance(data, (dict, list)):
            # Format JSON with proper indentation (4 spaces base, 2 spaces for structure)
            pretty_data = json.dumps(data, indent=2)
            # Indent each line of the JSON by 6 spaces (4 base + 2 for -d flag alignment)
            indented_lines = ['      ' + line for line in pretty_data.split('\n')]
            indented_data = '\n'.join(indented_lines)
            curl_lines.append(f"    -d '\n{indented_data}\n    '")
        else:
            curl_lines.append(f"    -d '{data}'")
    
    # Add trailing backslash for all but last line
    for i in range(len(curl_lines) - 1):
        curl_lines[i] += ' \\'
    
    curl_cmd = "\n".join(curl_lines) + " | jq ."
    return curl_cmd
