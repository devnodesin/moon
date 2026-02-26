---
agent: agent
---

## Striclty follow Test-Driven Development (TDD)

- For every feature, bugfix, or refactor:
  - Write one or more unit tests that define the expected behavior before implementation.
  - Update or add tests first, covering all relevant logic and edge cases.
  - Ensure tests are clear, isolated, and directly related to the change.
  - Validate that tests fail before implementation and pass after.
  - Maintain corresponding `*_test.go` files for all major logic modules.
  - Achieve at least 90% test coverage for new and updated code.
  - Run all tests after each change; fix failures immediately even the failures are not related to your implmentation.
  - Review tests for completeness, clarity, and SPEC compliance.
- Review all code for proper style, robust error handling, and strict compliance with `SPEC_API.md` and `SPEC_API\090-errors.md`.
- After making changes, start the moon server locally and verify correct behavior by running the Python API test script:  
  `cd scripts && python api-check.py --server=http://localhost:6000`
