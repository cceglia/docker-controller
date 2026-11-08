## CodeGraph — required before exploring the code

**CodeGraph** is the source of reference for understanding this project's code: it returns symbols, call paths, dependencies, and the impact of a change. There are two equivalent ways to use it:

- the MCP tool `mcp__codegraph__codegraph_explore`, when it is available in the session (it is deferred: load its schema with ToolSearch `select:mcp__codegraph__codegraph_explore` before the first call);
- the command `codegraph explore "<question or symbol names>"` through Bash, which always works.

**RULE: use CodeGraph before Read, Grep, Glob, or shell `grep`/`rg`/`find`** for code relationships, dependencies, callers, execution paths, behaviour that spans several files, and any other exploration of the source. You may skip it only when you already know the exact file path and line number you need. The rule also applies to every subagent you delegate to.

Query it with a natural-language question or with the relevant file and symbol names. For flow questions, include the symbols that span the flow, so CodeGraph can return the call path among them. Treat the returned source as already read and do not repeat the exploration with Grep or Read.

### Querying it well

The index covers the project's Go source. The `explore` output is limited (about 12 files by default). Therefore:

- Put symbol, file, or package names in the query (`internal/compose Manager Do Attività`, `internal/gui shell`), not generic words such as "handler" or "service": a vague query finds namesakes and misses the one you want.
- When the answer is truncated or shows only fragments (`... (gap) ...`), widen it with `codegraph explore --max-files <n> "…"` or narrow the query. Do not fall back to Grep.
- For a precise line range or a whole symbol, use the CLI instead of Read, `sed`, or `cat` (the MCP server exposes only `explore`): `codegraph node -f <repo-relative file> --offset <line> --limit <n>` (file mode, with line numbers) or `codegraph node <Symbol> [-f <file>]` (the whole symbol with callers and callees).

CodeGraph updates the index by itself through the MCP server. **Never run `codegraph update`, manual rebuilds, or other index updates** after changing the source. Without an active MCP server, `codegraph sync` is the only allowed command to align the index.

### Delegating to subagents

Every delegation prompt is written **entirely in English**: instructions, CodeGraph block, style rules, task, and report format. Do not mix Italian and English in the same prompt. The report is written in English unless the prompt asks for another language. The project documents stay in Italian.

Every delegation prompt that can explore source code starts with a **CodeGraph first** block:

- CodeGraph is the first exploration step and also every later step;
- Grep, Read, `sed`, or `cat` on source only serve a range CodeGraph could not return;
- the "Querying it well" rules above, copied in full, because a subagent does not apply this file reliably on its own;
- never `codegraph update`;
- the final report lists the CodeGraph queries run, so the delegating agent can check them.

Review delegations are narrow by default:

- a review subagent checks only the severity and the area named in its prompt; the default is blocking problems only;
- a review subagent does not re-check non-blocking findings from earlier reviews; a full review is a deliberate decision, not the default;
- after a fix, verify only the changed area, not the whole spec;
- do not pay for the same non-blocking check twice.

## Agent skills

### Issue tracker

Issues and specs are local markdown files in `docs/specs/`. See `docs/agents/issue-tracker.md`.

**MANDATORY RULE: every folder inside `docs/specs/` carries a numeric ID as a prefix, in the form `<id>-<name>`.** The ID is two digits and grows in creation order: `01-docker-controller-v1`, `02-restyling-interfaccia`, `03-wails-v3-stabile`, `04-gotk4`. A spec folder without a prefix is not allowed, not even for a newly created, temporary, or test spec. A used ID is never reused. If you create a spec, first choose the next number, then the folder.

### UI

The view uses Qt 6 Widgets through miqt. Every change to a window, a dialog, a form, a card, a menu, or the system icon first loads the local `.agents/skills/miqt-ui-design` skill and follows its rules: layout hierarchy, type scale, style sheet roles, focus, accessible names, tray, and Qt threads.

Layout and navigation reviews, and UX decisions, use `.agents/skills/qt-ui-design`. The rule also applies to every subagent you delegate UI work to.

Verifying a UI change goes through `make test-ui`: offscreen smoke test of the window and design token tests. The screenshot pass stays manual.

### Triage labels

The five canonical default labels (needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` + `docs/adr/` at the root. See `docs/agents/domain.md`.
