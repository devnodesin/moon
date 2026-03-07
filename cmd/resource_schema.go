package main

import (
	"net/http"
)

// ResourceSchemaHandler implements GET /data/{resource}:schema.
type ResourceSchemaHandler struct {
	registry *SchemaRegistry
	prefix   string
}

// NewResourceSchemaHandler creates a ResourceSchemaHandler with the given dependencies.
func NewResourceSchemaHandler(registry *SchemaRegistry, prefix string) *ResourceSchemaHandler {
	return &ResourceSchemaHandler{
		registry: registry,
		prefix:   prefix,
	}
}

// fieldDescriptor is the JSON representation of a single field in a schema response.
type fieldDescriptor struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Unique   bool   `json:"unique"`
	ReadOnly bool   `json:"readonly"`
}

// schemaObject is the JSON representation of a collection schema.
type schemaObject struct {
	Name   string            `json:"name"`
	Fields []fieldDescriptor `json:"fields"`
}

// HandleSchema handles GET /data/{resource}:schema requests.
func (h *ResourceSchemaHandler) HandleSchema(w http.ResponseWriter, r *http.Request) {
	resource := extractResource(r.URL.Path)
	if resource == "" {
		WriteError(w, http.StatusBadRequest, "Missing resource name")
		return
	}

	col, ok := h.registry.Get(resource)
	if !ok {
		WriteError(w, http.StatusNotFound, "Collection not found")
		return
	}

	apiFields := col.APIFields()
	descriptors := make([]fieldDescriptor, len(apiFields))
	for i, f := range apiFields {
		descriptors[i] = fieldDescriptor{
			Name:     f.Name,
			Type:     f.Type,
			Nullable: f.Nullable,
			Unique:   f.Unique,
			ReadOnly: f.ReadOnly,
		}
	}

	schema := schemaObject{
		Name:   col.Name,
		Fields: descriptors,
	}

	WriteSuccess(w, http.StatusOK, "Schema retrieved successfully", []any{schema})
}
