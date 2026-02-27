Analyze the gap between the source code and the specification files.

Compare the implementation in the source code with the following specification files to ensure complete compliance:

- SPEC.md
- SPEC_API.md
- SPEC_AUTH.md
- SPEC_API/*.md

For each file listed in `SPEC_API/*.md`, verify that the source code is fully compliant with the corresponding API specification:

Document your findings in a new file named GAP_ANALYSIS.md. This document should include:

- Features present in the code but not specified in the spec.
- Features specified but not implemented in the code.
- A clear, structured analysis of the gaps.
- Issues or inconsistencies between the current spec and code that need to be addressed.
- Any other relevant observations.

Do not modify any other source code; only create and write the GAP_ANALYSIS.md file.

- Run the moon server locally and verify correct behavior againts the local moon server by running the Python API test script:  
  `cd scripts && python api-check.py --server=http://localhost:6000`
