"""Placeholder replacement utilities for API testing."""

import json
import re
import requests
from typing import Optional, Dict, Any, List, Tuple

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
    Replace $PREV_CURSOR, $NEXT_CURSOR, $ULID, $ULID1, $ULID2, etc. in test definition.

    Replacement order (all independent — multiple placeholders can coexist):
      1. $PREV_CURSOR  — meta.prev from the most recent list response
      2. $NEXT_CURSOR  — meta.next from the most recent list response;
                         falls back to captured_record_id when cursors have
                         not yet been initialised by a list response.
      3. $ULID1/$ULID2 — numbered placeholders from a fresh collection fetch
      4. $ULID         — single captured record ID

    Args:
        test: Test definition to modify in-place
        context: Placeholder context with captured record ID(s) and cursors

    Returns:
        The last placeholder type that was substituted, or None
    """
    placeholder_used = None

    def _sub(text: str, old: str, new: str) -> str:
        return text.replace(old, new) if old in text else text

    # 1. $PREV_CURSOR
    if context.prev_cursor:
        if test.endpoint and "$PREV_CURSOR" in test.endpoint:
            test.endpoint = _sub(test.endpoint, "$PREV_CURSOR", context.prev_cursor)
            placeholder_used = "$PREV_CURSOR"
        if test.data:
            data_str = json.dumps(test.data)
            if "$PREV_CURSOR" in data_str:
                test.data = json.loads(_sub(data_str, "$PREV_CURSOR", context.prev_cursor))
                placeholder_used = "$PREV_CURSOR"

    # 2. $NEXT_CURSOR
    # When cursors have been initialised by a list response use next_cursor
    # (may be None on the last page).  Before any list response has run,
    # fall back to the captured_record_id so legacy tests still work.
    if context.cursors_initialized:
        next_cursor_value = context.next_cursor
    else:
        next_cursor_value = context.captured_record_id

    if next_cursor_value:
        if test.endpoint and "$NEXT_CURSOR" in test.endpoint:
            test.endpoint = _sub(test.endpoint, "$NEXT_CURSOR", next_cursor_value)
            placeholder_used = "$NEXT_CURSOR"
        if test.data:
            data_str = json.dumps(test.data)
            if "$NEXT_CURSOR" in data_str:
                test.data = json.loads(_sub(data_str, "$NEXT_CURSOR", next_cursor_value))
                placeholder_used = "$NEXT_CURSOR"

    # 3. Numbered $ULID1, $ULID2, … placeholders
    if context.captured_record_ids:
        if test.endpoint:
            test.endpoint = _replace_numbered_placeholders(
                test.endpoint, context.captured_record_ids
            )
        if test.data:
            data_str = _replace_numbered_placeholders(
                json.dumps(test.data), context.captured_record_ids
            )
            test.data = json.loads(data_str)
        placeholder_used = placeholder_used or "$ULID"

    # 4. Single $ULID
    if context.captured_record_id:
        if test.endpoint and "$ULID" in test.endpoint:
            test.endpoint = _sub(test.endpoint, "$ULID", context.captured_record_id)
            placeholder_used = placeholder_used or "$ULID"
        if test.data:
            data_str = json.dumps(test.data)
            if "$ULID" in data_str:
                test.data = json.loads(_sub(data_str, "$ULID", context.captured_record_id))
                placeholder_used = placeholder_used or "$ULID"

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


def extract_cursors_from_response(
    response_obj: Optional[Dict[str, Any]]
) -> Tuple[Optional[str], Optional[str]]:
    """
    Extract pagination cursors from the meta field of a list response.

    Args:
        response_obj: Response JSON object

    Returns:
        Tuple of (next_cursor, prev_cursor); either value may be None
    """
    if not response_obj or not isinstance(response_obj, dict):
        return None, None
    meta = response_obj.get("meta")
    if not isinstance(meta, dict):
        return None, None
    return meta.get("next"), meta.get("prev")


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
