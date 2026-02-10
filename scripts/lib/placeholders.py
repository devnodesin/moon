"""Placeholder replacement utilities for API testing."""

import json
import requests
from typing import Optional, Dict, Any

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
    Replace $ULID and $NEXT_CURSOR placeholders in test definition.
    
    Args:
        test: Test definition to modify in-place
        context: Placeholder context with captured record ID
        
    Returns:
        The placeholder type that was used ('$ULID', '$NEXT_CURSOR', or None)
    """
    if not context.captured_record_id:
        return None
    
    placeholder_used = None
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


def extract_record_id_from_response(
    response_obj: Optional[Dict[str, Any]]
) -> Optional[str]:
    """
    Extract record ID from a create or list response.
    Checks common patterns like data.id, record.id, id, etc.
    
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
    
    # Try data.id
    if "data" in response_obj and isinstance(response_obj["data"], dict):
        if "id" in response_obj["data"]:
            return response_obj["data"]["id"]
    
    # Try record.id
    if "record" in response_obj and isinstance(response_obj["record"], dict):
        if "id" in response_obj["record"]:
            return response_obj["record"]["id"]
    
    # Try arrays in common field names
    for array_key in ["apikeys", "users", "data", "records", "items"]:
        if array_key in response_obj:
            items = response_obj[array_key]
            if isinstance(items, list) and len(items) > 0:
                # For users array, select second record if available, otherwise first
                if array_key == "users" and len(items) > 1:
                    selected_item = items[1]
                else:
                    selected_item = items[0]
                
                if isinstance(selected_item, dict) and "id" in selected_item:
                    return selected_item["id"]
    
    # Try other common ID field names
    for key in ["_id", "ulid", "uuid"]:
        if key in response_obj:
            return response_obj[key]
        if "data" in response_obj and isinstance(response_obj["data"], dict):
            if key in response_obj["data"]:
                return response_obj["data"][key]
    
    return None


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
            for array_key in ["data", "records", "items", "apikeys", "users"]:
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
