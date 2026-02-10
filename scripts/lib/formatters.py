"""Output formatting utilities for API test results."""

from typing import Optional, List

from .types import AuthState


def format_markdown_result(
    curl_cmd: str,
    status: str,
    body: str,
    test_name: Optional[str] = None,
    details: Optional[str] = None,
    notes: Optional[str] = None
) -> List[str]:
    """
    Format a single API test result as Markdown lines.
    
    Args:
        curl_cmd: Formatted curl command
        status: HTTP response status
        body: Response body (JSON or text)
        test_name: Optional test name for heading
        details: Optional details to include before curl command
        notes: Optional notes to include after response
        
    Returns:
        List of markdown lines
    """
    lines = []
    
    if test_name:
        lines.append(f"### {test_name}\n")
    
    if details:
        lines.append(f"{details}\n")
    
    if notes:
        lines.append(f"{notes}\n")
    
    lines.append(f"```bash\n{curl_cmd}\n```")
    lines.append(f"\n**Response ({status}):**\n")
    lines.append(f"```json\n{body}\n```\n")
    
    return lines


def sanitize_curl_for_documentation(
    curl_cmd: str,
    server_url: str,
    doc_url: str,
    auth_state: AuthState,
    record_id: Optional[str] = None,
    placeholder_type: Optional[str] = None
) -> str:
    """
    Sanitize curl command for documentation by replacing actual values with placeholders.
    
    Args:
        curl_cmd: Original curl command
        server_url: Actual server URL to replace
        doc_url: Documentation URL to use
        auth_state: Authentication state with tokens to replace
        record_id: Actual record ID to replace (if any)
        placeholder_type: Placeholder type to use for record ID
        
    Returns:
        Sanitized curl command suitable for documentation
    """
    sanitized = curl_cmd
    
    # Replace server URL with doc URL
    sanitized = sanitized.replace(server_url, doc_url)
    
    # Replace all access tokens with placeholder
    for token in auth_state.all_access_tokens:
        if token:
            sanitized = sanitized.replace(token, "$ACCESS_TOKEN")
    
    # Replace all refresh tokens with placeholder
    for token in auth_state.all_refresh_tokens:
        if token:
            sanitized = sanitized.replace(token, "$REFRESH_TOKEN")
    
    # Replace actual record ID with placeholder
    if record_id and placeholder_type:
        sanitized = sanitized.replace(record_id, placeholder_type)
    
    return sanitized


def write_markdown_output(
    output_file: str,
    markdown_lines: List[str]
) -> None:
    """
    Write markdown lines to output file.
    
    Args:
        output_file: Path to output file
        markdown_lines: List of markdown content lines
    """
    markdown = "\n".join(markdown_lines)
    with open(output_file, "w", encoding="utf-8") as f:
        f.write(markdown)
