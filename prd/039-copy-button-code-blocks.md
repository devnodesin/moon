## Overview

The HTML documentation endpoint (`/doc/html`) currently displays code examples in `<pre><code>` blocks with syntax highlighting, but lacks an easy way for users to copy these commands to their clipboard. This creates friction when users want to try the API examples, as they must manually select and copy text.

This PRD introduces a **copy-to-clipboard button** for all code blocks in the generated HTML documentation. Users will be able to click a button to instantly copy the entire command from any code block to their clipboard, ready to paste into their terminal or code editor.

**Goals:**
- Improve developer experience by reducing friction when using documentation examples
- Provide visual feedback when code is successfully copied
- Maintain clean, accessible, and mobile-friendly documentation design
- Keep implementation lightweight with minimal JavaScript

## Requirements

### Functional Requirements

**FR1: Copy Button Placement**
- A "Copy" button must be added to every `<pre><code>` block in the HTML documentation
- Button must be positioned in the top-right corner of each code block
- Button must be visually distinct but not intrusive
- Button must remain visible even when code blocks have scrollable content

**FR2: Copy Functionality**
- Clicking the copy button must copy the **entire text content** of the corresponding `<code>` element to the system clipboard
- The copied text must be the raw, unformatted text (no HTML tags or styling)
- Copy operation must use the modern Clipboard API (`navigator.clipboard.writeText()`)
- The exact text copied must match what the user would manually select and copy

**FR3: Visual Feedback**
- Button text must change from "Copy" to "Copied!" immediately after successful copy
- "Copied!" state must persist for 1.2 seconds before reverting to "Copy"
- Button styling should subtly indicate the state change (e.g., color change)
- No error should be thrown if clipboard access fails (graceful degradation)

**FR4: Styling and UX**
- Copy button must have:
  - Clear, readable text ("Copy")
  - Comfortable click target size (minimum 32x32px recommended)
  - Visible but non-obtrusive appearance
  - Hover state to indicate interactivity
  - Responsive design that works on mobile devices
- Code blocks must have `position: relative` to support absolute positioning of buttons
- Button must not obscure code content

**FR5: Automatic Button Injection**
- Copy buttons must be automatically injected via JavaScript after page load
- All `<pre><code>` blocks must be detected using `document.querySelectorAll('pre > code')`
- Button injection must work regardless of code block content or language
- No server-side template changes required for individual code blocks

### Technical Requirements

**TR1: JavaScript Implementation**
- Use vanilla JavaScript (no external libraries required)
- Code must be injected into the HTML output in the `generateHTML()` function
- JavaScript must be placed just before the closing `</body>` tag
- Code must be compatible with modern browsers (Chrome, Firefox, Safari, Edge)

**TR2: HTML Structure**
```html
<pre style="position: relative;">
  <code class="language-bash">curl -s http://localhost:6006/collections:list | jq .</code>
  <button class="copy-btn" style="position:absolute;top:8px;right:8px;...">Copy</button>
</pre>
```

**TR3: Clipboard API**
- Use `navigator.clipboard.writeText(text)` for modern clipboard access
- Extract text content using `codeBlock.innerText` to get plain text without HTML
- Handle promise resolution to trigger visual feedback
- No fallback for older browsers required (Moon targets modern environments)

**TR4: CSS Styling**
- Button styling must be inline or in the existing `<style>` block
- Minimum styling requirements:
  - Absolute positioning (top: 8px, right: 8px)
  - Padding for comfortable click area (4px 10px)
  - Cursor: pointer
  - Readable font size (0.9em)
  - Background color distinct from code block
  - Border and border-radius for visual definition
  - Hover state with slight color change
  - Transition for smooth state changes

**TR5: Integration with Existing Code**
- Modify the `generateHTML()` function in `cmd/moon/internal/handlers/doc.go`
- Add JavaScript snippet before the closing `</body>` tag in the HTML output
- Add CSS for `.copy-btn` styling in the existing `<style>` block
- No changes required to the Markdown template (`doc.md.tmpl`)
- No changes required to the `generateMarkdown()` function

### Implementation Specification

**Modification Location:**
- File: `cmd/moon/internal/handlers/doc.go`
- Function: `generateHTML()`
- Section: HTML template string builder

**JavaScript Code (to be added before `</body>`):**
```html
<script>
document.addEventListener('DOMContentLoaded', function() {
    document.querySelectorAll('pre > code').forEach(function(codeBlock) {
        var pre = codeBlock.parentNode;
        var button = document.createElement('button');
        button.innerText = 'Copy';
        button.className = 'copy-btn';
        
        pre.style.position = 'relative';
        pre.appendChild(button);
        
        button.addEventListener('click', function() {
            var text = codeBlock.innerText;
            navigator.clipboard.writeText(text).then(function() {
                button.innerText = 'Copied!';
                button.classList.add('copied');
                setTimeout(function() {
                    button.innerText = 'Copy';
                    button.classList.remove('copied');
                }, 1200);
            }).catch(function(err) {
                console.error('Failed to copy text: ', err);
            });
        });
    });
});
</script>
```

**CSS Additions (to existing `<style>` block):**
```css
.copy-btn {
    position: absolute;
    top: 8px;
    right: 8px;
    padding: 6px 12px;
    font-size: 0.85em;
    font-weight: 500;
    background: #34495e;
    color: #ecf0f1;
    border: 1px solid #2c3e50;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s ease;
    z-index: 10;
}

.copy-btn:hover {
    background: #2c3e50;
    border-color: #1a252f;
}

.copy-btn.copied {
    background: #27ae60;
    border-color: #229954;
}

pre {
    position: relative;
    padding-right: 80px; /* Space for copy button */
}
```

### Error Handling

**EH1: Clipboard API Unavailable**
- If `navigator.clipboard` is undefined (older browsers or insecure contexts), log error to console
- Button should still appear but fail silently with console error message
- No user-facing error message required

**EH2: Copy Operation Failure**
- If `writeText()` promise rejects, log error to console
- Do not change button text to "Copied!"
- No user-facing error message required

### Constraints and Assumptions

**Assumptions:**
- Users are using modern browsers with Clipboard API support (Chrome 63+, Firefox 53+, Safari 13.1+, Edge 79+)
- Users are accessing the documentation over HTTPS or localhost (required for Clipboard API)
- JavaScript is enabled in the browser
- Code blocks contain text content suitable for clipboard copying

**Out of Scope:**
- Fallback implementation for Internet Explorer or very old browsers
- Copy button for inline `<code>` elements (only `<pre><code>` blocks)
- Customizable copy button text or styling via configuration
- Copy with syntax highlighting (only plain text)
- Multiple copy format options (e.g., with/without comments)

## Acceptance Criteria

**AC1: Button Presence**
- [ ] All `<pre><code>` blocks in the HTML documentation have a "Copy" button in the top-right corner
- [ ] Buttons are visually consistent across all code blocks
- [ ] Buttons do not obscure code content

**AC2: Copy Functionality**
- [ ] Clicking any "Copy" button copies the exact text content of the corresponding code block to clipboard
- [ ] Copied text can be pasted into a terminal and executes correctly (for bash commands)
- [ ] No HTML tags or extra whitespace are included in the copied text
- [ ] Manual testing: Copy a curl command from the docs and paste it into a terminal - it should execute without modification

**AC3: Visual Feedback**
- [ ] Button text changes from "Copy" to "Copied!" immediately after click
- [ ] Button background color changes to green when in "Copied!" state
- [ ] "Copied!" state persists for 1.2 seconds
- [ ] Button automatically reverts to "Copy" state after timeout

**AC4: User Experience**
- [ ] Hover effect is visible when cursor is over the copy button
- [ ] Button is easily clickable on both desktop and mobile devices
- [ ] Copy button styling matches the overall documentation theme
- [ ] Code blocks remain readable with the copy button present

**AC5: Browser Compatibility**
- [ ] Feature works in Chrome (latest)
- [ ] Feature works in Firefox (latest)
- [ ] Feature works in Safari (latest)
- [ ] Feature works in Edge (latest)

**AC6: Error Handling**
- [ ] If Clipboard API fails, an error is logged to the browser console
- [ ] Page remains functional even if clipboard access is denied

**AC7: Code Integration**
- [ ] No changes to the Markdown template
- [ ] JavaScript is properly embedded in the HTML output
- [ ] CSS styling is in the existing `<style>` block
- [ ] Documentation cache refresh still works correctly

**AC8: Testing Scenarios**

**Scenario 1: Copy Simple Command**
1. Open `/doc/html` in a browser
2. Locate a code block with a curl command
3. Click the "Copy" button
4. Verify button changes to "Copied!" with green background
5. Paste into a text editor
6. Expected: Exact command text is pasted

**Scenario 2: Copy Multi-line Code**
1. Find a code block with multiple lines (e.g., JSON payload)
2. Click "Copy" button
3. Paste into a text editor
4. Expected: All lines are copied with correct line breaks

**Scenario 3: Multiple Copies**
1. Copy from one code block
2. Immediately copy from a different code block
3. Expected: Each copy operation works independently, each button shows its own state

**Scenario 4: Cache Refresh**
1. Generate HTML documentation
2. Call `/doc/refresh` endpoint
3. Access `/doc/html` again
4. Expected: Copy buttons still function correctly after cache refresh

**AC9: Documentation Updates**
- [ ] Ensure all documentation, scripts, and samples (`SPEC.md`, `INSTALL.md`, `README.md`, `USAGE.md`, `install.sh`, and all files in `samples/*`) are updated and remain consistent with the implemented code changes.

## Test Plan

### Unit Tests

No unit tests required for this feature as it is pure client-side JavaScript.

### Manual Testing

**Test 1: Basic Copy Functionality**
- Access `/doc/html`
- Click copy button on first code block
- Paste into terminal
- Verify exact command is pasted

**Test 2: Visual State Changes**
- Click copy button
- Observe button changes to "Copied!" with green background
- Wait 1.5 seconds
- Verify button reverts to "Copy"

**Test 3: Multiple Code Blocks**
- Count all code blocks on the page
- Verify each has a copy button
- Click copy on 3 different blocks
- Verify each works independently

**Test 4: Long Code Blocks**
- Find a code block with horizontal scroll
- Verify copy button remains visible when scrolling code
- Copy and paste
- Verify all text is copied

**Test 5: Mobile Responsiveness**
- Open documentation on mobile device or responsive view
- Verify copy buttons are visible and clickable
- Test copy functionality on mobile

### Edge Cases

- Empty code blocks (should still have button, copy empty string)
- Code blocks with special characters (quotes, backslashes)
- Code blocks with Unicode characters
- Very long single-line commands

## Implementation Notes

- The feature is entirely client-side and requires no server-side changes beyond modifying the HTML output
- Cache behavior is unchanged - cached HTML will include the JavaScript
- No performance impact on Markdown generation or API response times
- The Clipboard API requires a secure context (HTTPS or localhost), which is satisfied by typical Moon deployments
- This feature improves documentation usability without any breaking changes