## Recreate a Comprehensive, Spec-Compliant `samples/moon.conf`

**Goal:**  
Remove `samples/moon-full.conf` and replace all sample configuration with a single, comprehensive `samples/moon.conf` that is minimal by default but includes all advanced options (commented out), strictly following `SPEC.md` and `SPEC_AUTH.md`.

### Steps

1. **Review the Specs:**  
   - Carefully review `SPEC.md` and `SPEC_AUTH.md` for all required and optional configuration options, including structure, defaults, and documentation requirements.

2. **Audit Existing Configs:**  
   - Compare `samples/moon.conf` and `samples/moon-full.conf` to identify all available options, defaults, and documentation.
   - Preserve all config values currently in `samples/moon.conf` unless the spec requires otherwise.
   - Note any options present in the configs but missing from the specs (flag these for review).

3. **Check Code for Extra Configs:**  
   - Search the codebase for any config options used in code but not present in `SPEC.md`.
   - List these options separately for review.

4. **Recreate `samples/moon.conf`:**  
   - Create a new `samples/moon.conf`:
     - Enable only essential configuration options by default (server, database, minimal auth).
     - Include all advanced and optional options, but comment them out.
     - Ensure all options and inline documentation are up-to-date and spec-compliant.
     - Add a header comment explaining the file is minimal by default, and users should uncomment/configure as needed.

5. **Remove Redundant Config:**  
   - Delete `samples/moon-full.conf` from the repository.

6. **Update Documentation and Scripts:**  
   - Search and update all documentation, scripts, and specs to reference only `samples/moon.conf` as the sample config file.
   - Ensure clarity for new users about the single config file approach.

7. **Spec Compliance and Review:**  
   - Double-check the new config for strict compliance with `SPEC.md` and `SPEC_AUTH.md`.
   - Ensure clarity, completeness, and helpful inline documentation for users.

### Checklist

- [ ] All config values in `samples/moon.conf` are preserved unless the spec requires changes.
- [ ] Any config option found in code but not in the spec is listed for review.
- [ ] All advanced options are present but commented out.
- [ ] Inline documentation is clear and up-to-date.
- [ ] Only `samples/moon.conf` is referenced in docs/scripts.
- [ ] Header comment explains minimal-by-default, spec-compliant nature.

**If you notice anything missing or unclear, please highlight it for review.**