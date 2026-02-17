"""Placeholder replacement utilities for API testing."""

import json
import re
import requests
from typing import Optional, Dict, Any, List

from .types import TestDefinition, PlaceholderContext, AuthState


def replace_auth_placeholders(
    test: TestDefinition,
    auth_state: AuthState
) -> None:
    """
    Replace $ACCESS_TOKEN and $REFRESH_TOKEN placeholders in test definition.
    
    Args:
        test: Test definition to modify in-place
        auth_state: Current authentication state
    """
    # Replace $ACCESS_TOKEN in headers
    if auth_state.access_token and test.headers:
        if "Authorization" in test.headers:
            test.headers["Authorization"] = test.headers["Authorization"].replace(
                "$ACCESS_TOKEN", auth_state.access_token
            )
    
    # Replace $REFRESH_TOKEN in data
    if auth_state.refresh_token and test.data:
        data_str = json.dumps(test.data)
        if "$REFRESH_TOKEN" in data_str:
            data_str = data_str.replace("$REFRESH_TOKEN", auth_state.refresh_token)
            test.data = json.loads(data_str)


def replace_record_placeholders(
    test: TestDefinition,
    context: PlaceholderContext
) -> Optional[str]:
    """
    Replace $ULID, $ULID1, $ULID2, etc., and $NEXT_CURSOR placeholders in test definition.
    
    Args:
        test: Test definition to modify in-place
        context: Placeholder context with captured record ID(s)
        
    Returns:
        The placeholder type that was used ('$ULID', '$NEXT_CURSOR', or None)
    """
    placeholder_used = None
    
    # Handle numbered ULID placeholders (e.g., $ULID1, $ULID2)
    if context.captured_record_ids:
        # Replace in endpoint
        if test.endpoint:
            test.endpoint = _replace_numbered_placeholders(
                test.endpoint, context.captured_record_ids
            )
        
        # Replace in data
        if test.data:
            data_str = json.dumps(test.data)
            data_str = _replace_numbered_placeholders(
                data_str, context.captured_record_ids
            )
            test.data = json.loads(data_str)
            if "$ULID" in json.dumps(context.captured_record_ids):
                placeholder_used = "$ULID"
    
    # Handle single record ID placeholders
    if not context.captured_record_id:
        return placeholder_used
    
    record_id = context.captured_record_id
    
    # Replace in endpoint
    if test.endpoint:
        if "$NEXT_CURSOR" in test.endpoint:
            test.endpoint = test.endpoint.replace("$NEXT_CURSOR", record_id)
            placeholder_used = "$NEXT_CURSOR"
        elif "$ULID" in test.endpoint:
            test.endpoint = test.endpoint.replace("$ULID", record_id)
            placeholder_used = "$ULID"
    
    # Replace in data (recursive)
    if test.data:
        data_str = json.dumps(test.data)
        if "$NEXT_CURSOR" in data_str:
            data_str = data_str.replace("$NEXT_CURSOR", record_id)
            placeholder_used = "$NEXT_CURSOR"
        elif "$ULID" in data_str:
            data_str = data_str.replace("$ULID", record_id)
            placeholder_used = "$ULID"
        test.data = json.loads(data_str)
    
    return placeholder_used


def _replace_numbered_placeholders(text: str, record_ids: List[str]) -> str:
    """
    Replace numbered placeholders like $ULID1, $ULID2 with actual record IDs.
    
    Args:
        text: Text containing placeholders
        record_ids: List of record IDs to use for replacement
        
    Returns:
        Text with placeholders replaced
    """
    # Find all numbered ULID placeholders
    pattern = r'\$ULID(\d+)'
    matches = re.findall(pattern, text)
    
    # Replace each numbered placeholder with corresponding record ID
    for match in matches:
        index = int(match) - 1  # Convert to 0-based index
        if 0 <= index < len(record_ids):
            placeholder = f"$ULID{match}"
            text = text.replace(placeholder, record_ids[index])
    
    return text


def extract_record_id_from_response(
    response_obj: Optional[Dict[str, Any]]
) -> Optional[str]:
    """
    Extract record ID from a create or list response.
    Checks common patterns like data.id, record.id, id, etc.
    Also extracts meta.next cursor for pagination.
    
    Args:
        response_obj: Response JSON object
        
    Returns:
        Extracted record ID or None
    """
    if not response_obj or not isinstance(response_obj, dict):
        return None
    
    # Try direct id field
    if "id" in response_obj:
        return response_obj["id"]
    
    # Try data.id (single object response)
    if "data" in response_obj and isinstance(response_obj["data"], dict):
        if "id" in response_obj["data"]:
            return response_obj["data"]["id"]
    
    # Try record.id
    if "record" in response_obj and isinstance(response_obj["record"], dict):
        if "id" in response_obj["record"]:
            return response_obj["record"]["id"]
    
    # Try arrays in common field names
    for array_key in ["data", "records", "items"]:
        if array_key in response_obj:
            items = response_obj[array_key]
            if isinstance(items, list) and len(items) > 0:
                # For list responses, select second record if available (for user management tests)
                if len(items) > 1:
                    selected_item = items[1]
                else:
                    selected_item = items[0]
                
                if isinstance(selected_item, dict) and "id" in selected_item:
                    return selected_item["id"]
    
    # Try meta.next for pagination cursor
    if "meta" in response_obj and isinstance(response_obj["meta"], dict):
        next_cursor = response_obj["meta"].get("next")
        if next_cursor:
            return next_cursor
    
    # Try other common ID field names
    for key in ["_id", "ulid", "uuid"]:
        if key in response_obj:
            return response_obj[key]
        if "data" in response_obj and isinstance(response_obj["data"], dict):
            if key in response_obj["data"]:
                return response_obj["data"][key]
    
    return None


def extract_record_ids_from_response(
    response_obj: Optional[Dict[str, Any]],
    max_count: int = 10
) -> List[str]:
    """
    Extract multiple record IDs from a list response.
    
    Args:
        response_obj: Response JSON object
        max_count: Maximum number of IDs to extract
        
    Returns:
        List of extracted record IDs (may be empty)
    """
    if not response_obj or not isinstance(response_obj, dict):
        return []
    
    record_ids = []
    
    # Try arrays in common field names
    for array_key in ["data", "records", "items"]:
        if array_key in response_obj:
            items = response_obj[array_key]
            if isinstance(items, list):
                for item in items[:max_count]:
                    if isinstance(item, dict):
                        # Try common ID field names
                        for id_field in ["id", "_id", "ulid", "uuid"]:
                            if id_field in item:
                                record_ids.append(item[id_field])
                                break
                        if len(record_ids) >= max_count:
                            break
                break
    
    return record_ids


def fetch_record_id_from_collection(
    base_url: str,
    prefix: str,
    collection_name: str,
    headers: Optional[Dict[str, str]],
    timeout: int = 10
) -> Optional[str]:
    """
    Fetch the first record ID from a collection by calling /{collection}:list.
    
    Args:
        base_url: Base server URL
        prefix: URL prefix
        collection_name: Name of the collection
        headers: Request headers
        timeout: Request timeout in seconds
        
    Returns:
        First record ID or None
    """
    try:
        list_endpoint = f"/{collection_name}:list"
        url = f"{base_url}{prefix}{list_endpoint}"
        resp = requests.get(url, headers=headers, timeout=timeout)
        
        if resp.status_code == 200:
            data = resp.json()
            # Try to find records in common response structures
            for array_key in ["data", "records", "items"]:
                records = data.get(array_key, [])
                if records and len(records) > 0:
                    # Try common ID field names
                    first_record = records[0]
                    for id_field in ["id", "_id", "ulid"]:
                        if id_field in first_record:
                            return first_record[id_field]
    except Exception:
        pass
    
    return None


def extract_collection_name(endpoint: str) -> Optional[str]:
    """
    Extract the collection name from an endpoint.
    E.g., "/products:get" -> "products"
    
    Args:
        endpoint: API endpoint path
        
    Returns:
        Collection name or None
    """
    if '/' in endpoint:
        endpoint = endpoint.split('?')[0]  # Remove query params
        parts = endpoint.split('/')
        for part in parts:
            if ':' in part:
                return part.split(':')[0]
    return None
