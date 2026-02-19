"""Test execution orchestration for API testing framework."""

import re
from typing import List, Optional

from .types import (
    TestDefinition,
    TestSuite,
    TestResult,
    AuthState,
    PlaceholderContext
)
from .http_client import execute_request
from .auth import (
    detect_password_change,
    detect_token_refresh,
    detect_login,
    extract_tokens_from_response,
    should_relogin_after_test,
    relogin_with_new_password
)
from .placeholders import (
    replace_auth_placeholders,
    replace_record_placeholders,
    extract_record_id_from_response,
    extract_record_ids_from_response
)
from .formatters import (
    format_markdown_result,
    sanitize_curl_for_documentation,
    write_markdown_output
)


def run_test_suite(
    test_suite: TestSuite,
    auth_state: Optional[AuthState],
    output_file: Optional[str]
) -> TestResult:
    """
    Execute all tests in a test suite and generate markdown output.
    
    Args:
        test_suite: Test suite definition
        auth_state: Initial authentication state (can be None)
        output_file: Path to write markdown output
        
    Returns:
        TestResult with execution status and output
    """
    markdown_lines: List[str] = []
    all_tests_passed = True
    placeholder_context = PlaceholderContext()
    
    # Initialize auth state if not provided
    if auth_state is None:
        auth_state = AuthState()
    
    for test in test_suite.tests:
        # Make a copy to avoid modifying the original
        test_copy = _copy_test(test)
        
        # Replace authentication placeholders first (needed for API calls)
        replace_auth_placeholders(test_copy, auth_state)
        
        # Check if this test needs fresh numbered placeholders
        if _test_needs_numbered_placeholders(test_copy):
            # Extract collection name from the test endpoint
            collection_name = _extract_collection_name_from_test(test_copy)
            if collection_name:
                # Fetch fresh IDs from the collection (headers now have real tokens)
                fresh_ids = _fetch_fresh_record_ids(
                    test_suite.serverURL,
                    test_suite.prefix,
                    collection_name,
                    test_copy.headers
                )
                if fresh_ids:
                    # Create a temporary context with fresh IDs
                    temp_context = PlaceholderContext()
                    temp_context.set_record_ids(fresh_ids)
                    replace_record_placeholders(test_copy, temp_context)
        
        # Replace single record placeholders if we have a captured ID
        elif placeholder_context.captured_record_id:
            placeholder_type = replace_record_placeholders(test_copy, placeholder_context)
            if placeholder_type:
                placeholder_context.placeholder_type = placeholder_type
        
        # Detect password change before execution
        new_password = detect_password_change(test_copy)
        
        # Execute the test
        response = execute_request(
            test_suite.serverURL,
            test_suite.prefix,
            test_copy
        )
        
        # Check for token refresh or login and update tokens
        is_successful = response.status.startswith("2")
        if is_successful and response.response_obj:
            is_login = detect_login(test_copy)
            is_refresh = detect_token_refresh(test_copy)
            
            if is_login or is_refresh:
                access_token, refresh_token = extract_tokens_from_response(
                    response.response_obj
                )
                if access_token:
                    auth_state.update_access_token(access_token)
                if refresh_token:
                    auth_state.update_refresh_token(refresh_token)
        
        # Handle password change: re-login with new password
        if should_relogin_after_test(test_copy, response.status, new_password):
            if new_password and auth_state.current_username:
                new_auth_state = relogin_with_new_password(
                    test_suite.serverURL,
                    auth_state.current_username,
                    new_password
                )
                if new_auth_state:
                    # Update credentials and tokens
                    auth_state.update_credentials(
                        auth_state.current_username,
                        new_password
                    )
                    if new_auth_state.access_token:
                        auth_state.update_access_token(new_auth_state.access_token)
                    if new_auth_state.refresh_token:
                        auth_state.update_refresh_token(new_auth_state.refresh_token)
        
        # Try to capture single record ID from successful create/list responses
        if is_successful and response.response_obj:
            endpoint = test_copy.endpoint
            if ":create" in endpoint or ":list" in endpoint:
                if not placeholder_context.captured_record_id:
                    record_id = extract_record_id_from_response(response.response_obj)
                    if record_id:
                        placeholder_context.set_record_id(record_id, "$ULID")
        
        # Sanitize curl command for documentation
        sanitized_curl = sanitize_curl_for_documentation(
            response.curl_command,
            test_suite.serverURL,
            test_suite.docURL,
            auth_state,
            placeholder_context.captured_record_id,
            placeholder_context.placeholder_type
        )
        
        # Only add to output if test has a name
        if test.name:
            markdown_lines.extend(format_markdown_result(
                sanitized_curl,
                response.status,
                response.body,
                test.name,
                test.details,
                test.notes
            ))
        
        # Track overall success — only named tests count toward pass/fail;
        # unnamed tests are cleanup/preparation steps that may fail intentionally.
        if test.name and not is_successful:
            all_tests_passed = False
    
    # Write output to file if specified
    if output_file:
        write_markdown_output(output_file, markdown_lines)
    
    return TestResult(
        status="success" if all_tests_passed else "failure",
        output_file=output_file,
        markdown="\n".join(markdown_lines),
        all_tests_passed=all_tests_passed
    )


def _test_needs_numbered_placeholders(test: TestDefinition) -> bool:
    """
    Check if a single test uses numbered placeholders like $ULID1, $ULID2.
    
    Args:
        test: Test definition
        
    Returns:
        True if numbered placeholders are found
    """
    pattern = r'\$ULID\d+'
    
    # Check endpoint
    if test.endpoint and re.search(pattern, test.endpoint):
        return True
    
    # Check data
    if test.data:
        import json
        data_str = json.dumps(test.data)
        if re.search(pattern, data_str):
            return True
    
    return False


def _extract_collection_name_from_test(test: TestDefinition) -> Optional[str]:
    """
    Extract collection name from a test endpoint.
    E.g., "/products:destroy" -> "products"
    
    Args:
        test: Test definition
        
    Returns:
        Collection name or None
    """
    if not test.endpoint:
        return None
    
    endpoint = test.endpoint.split('?')[0]  # Remove query params
    parts = endpoint.split('/')
    
    for part in parts:
        if ':' in part:
            return part.split(':')[0]
    
    return None


def _fetch_fresh_record_ids(
    base_url: str,
    prefix: str,
    collection_name: str,
    headers: Optional[dict],
    max_count: int = 10
) -> List[str]:
    """
    Fetch fresh record IDs from a collection by calling /{collection}:list.
    
    Args:
        base_url: Base server URL
        prefix: URL prefix
        collection_name: Name of the collection
        headers: Request headers (may contain auth token)
        max_count: Maximum number of IDs to fetch
        
    Returns:
        List of fresh record IDs (may be empty)
    """
    import requests
    
    try:
        list_endpoint = f"/{collection_name}:list"
        url = f"{base_url}{prefix}{list_endpoint}"
        
        # Make request with timeout
        resp = requests.get(url, headers=headers, timeout=10)
        
        if resp.status_code == 200:
            data = resp.json()
            
            # Try to find records in common response structures
            for array_key in ["data", "records", "items"]:
                records = data.get(array_key, [])
                if records and isinstance(records, list):
                    record_ids = []
                    for record in records[:max_count]:
                        if isinstance(record, dict):
                            # Try common ID field names
                            for id_field in ["id", "_id", "ulid", "uuid"]:
                                if id_field in record:
                                    record_ids.append(record[id_field])
                                    break
                            if len(record_ids) >= max_count:
                                break
                    return record_ids
    except Exception:
        pass
    
    return []


def _copy_test(test: TestDefinition) -> TestDefinition:
    """
    Create a deep copy of a test definition to avoid modifying the original.
    
    Args:
        test: Original test definition
        
    Returns:
        Copy of the test definition
    """
    import copy
    return TestDefinition(
        name=test.name,
        cmd=test.cmd,
        endpoint=test.endpoint,
        headers=copy.deepcopy(test.headers) if test.headers else None,
        data=copy.deepcopy(test.data) if test.data else None,
        details=test.details,
        notes=test.notes
    )
