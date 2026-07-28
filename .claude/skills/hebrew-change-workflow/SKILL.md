---
name: hebrew-change-workflow
description: Use BEFORE implementing any change in the comtec-devteam project — a new feature/story, a bug fix, or any code change. Enforces the team rule: first open a Hebrew GitHub issue (Story or bug) whose body contains the Hebrew description or steps-to-reproduce PLUS a Hebrew Cucumber acceptance test, and write ALL tests as Hebrew Cucumber features. Trigger whenever the user asks to add a feature, fix a bug, or make a change here, or mentions writing tests for this project.
---

# Hebrew change workflow (comtec-devteam)

Every change in this project starts with a **Hebrew GitHub issue** and is verified with **Hebrew Cucumber tests**. This extends the English workflow in `~/projects/CLAUDE.md` with a Hebrew-first rule for this repo.

## Rule

1. **Before any change**, create a GitHub issue **in Hebrew**:
   - **Story** (feature): title + description in Hebrew, with acceptance criteria.
   - **Bug**: title + **steps to reproduce** (expected vs. actual) in Hebrew.
2. The issue body **must embed a Cucumber acceptance test in Hebrew** — a fenced ```gherkin block starting with `# language: he`.
3. Then branch and implement per `~/projects/CLAUDE.md` (branch `story/<n>-…` or `bug/<n>-…`).
4. **All tests are Cucumber, in Hebrew** — the `.feature` files under `features/` and their step definitions.

## Creating the issue

Use `gh` (authenticated in WSL). Write the body to a file to avoid quoting issues, then:

```bash
gh label create "Story" 2>/dev/null || true   # or "bug"
gh issue create --label "Story" --title "<כותרת בעברית>" --body-file /tmp/issue_body.md
```

Issue body template (Story):

```markdown
## תיאור
<מה ולמה, בעברית>

## קריטריוני קבלה
- [ ] ...

## בדיקת קבלה (Cucumber)
```gherkin
# language: he
תכונה: <שם התכונה>
  תרחיש: <שם התרחיש>
    בהינתן <מצב התחלתי>
    כאשר <פעולה>
    אז <תוצאה צפויה>
```
```

Bug body uses **## שלבים לשחזור**, **## התנהגות צפויה**, **## התנהגות בפועל** plus the same Hebrew Gherkin regression scenario.

## Writing the Cucumber tests

- Each `.feature` begins with `# language: he`.
- Hebrew keywords: `תכונה` (Feature), `רקע` (Background), `תרחיש` (Scenario), `תבנית תרחיש` (Scenario Outline), `דוגמאות` (Examples), `בהינתן` (Given), `כאשר` (When), `אז` (Then), `וגם` (And), `אבל` (But).
- Step definitions live in `features/step_definitions/` with **Hebrew matcher strings**; the Ruby DSL stays `Given/When/Then`. Reuse existing shared Hebrew steps in `steps.rb` rather than duplicating.
- **Keep in English:** UI-facing literals inside quotes (button labels, asserted page text) and data-table headers — they are matched against the real (English) UI or read by key, so translating them breaks the tests. Example: `וגם אני לוחץ "New Ticket"`.
- Possessive prefix attaches directly to the quote, no space: `ל"Beta Client" יש כרטיס פתוח...`.

## Running the suite (WSL + mise)

The Ruby toolchain runs through mise with a clean PATH (Windows PATH leaks into WSL and breaks `mise activate`):

```bash
wsl -d ubuntu -- bash -lc 'export PATH="/usr/local/bin:/usr/bin:/bin:$HOME/.local/bin"; cd ~/projects/comtec-devteam; ~/.local/bin/mise exec -- bundle exec cucumber'
```

Keep new features in the CI-run set; verify green before marking a story completed and relabeling the issue `Completed`.
