Short requirements plan — move /doc/md → /doc/llms-full.txt (no backward compatibility)

## Objective

- Replace the existing Markdown docs endpoint by removing GET /doc/md and exposing the exact same documentation at GET /doc/llms-full.txt. No aliasing or redirects; /doc/md is removed.

## Scope

- Add new public endpoint: GET {prefix}/doc/llms-full.txt
- Remove endpoint: GET {prefix}/doc/md
- Keep content, caching, headers, and public/no-auth semantics identical to current behavior.

## Functional requirements

- GET {prefix}/doc/llms-full.txt returns the same markdown bytes previously served at /doc/md.
- Content-Type: text/markdown; preserve ETag, Last-Modified and Cache-Control behavior.
- CORS: Access-Control-Allow-Origin: *; OPTIONS preflight supported.
- No authentication required.

## Non-functional requirements

- No change to template rendering or cache implementation.
- Startup logs should list new endpoint and note removal of old endpoint.
- Minimal performance impact.

## Tests (must pass)

- Verify GET /doc/llms-full.txt returns HTTP 200 and correct Content-Type.
- Verify ETag and/or Last-Modified headers are present and behave as before.
- Verify Access-Control-Allow-Origin: * present.
- Verify GET /doc/md returns 404 (or equivalent removed behavior).

## Acceptance criteria

- /doc/llms-full.txt serves identical markdown content and headers as the former /doc/md.
- /doc/md is no longer routable (404).
- README and PRD references updated to point to /doc/llms-full.txt.
- 100% tests pass.

# Minimal PR checklist

- [ ] Register new route /doc/llms-full.txt and remove /doc/md route
- [ ] Update docs and SPEC to reference /doc/llms-full.txt
- [ ] Add/adjust tests verifying new endpoint and removal of old


/doc/llms-full.txt