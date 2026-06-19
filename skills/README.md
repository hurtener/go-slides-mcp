# Deckard skills (for the agents that use Deckard)

These are **agent-facing skills**: drop them into the agent you connect to the
Deckard MCP server, and it will know how to drive Deckard's tools to produce
genuinely good decks — not just valid ones. They teach the *contract* (which tool
to call when) and the *taste* (how to make slides that look designed).

| Skill | Teaches |
|---|---|
| [`building-a-deck`](building-a-deck/SKILL.md) | The end-to-end loop: create → style → fill → validate → export. Start here. |
| [`composing-a-slide`](composing-a-slide/SKILL.md) | The slide node vocabulary and when to use each (hero, list, callout, two-column, grid, chart…). |
| [`design-principles`](design-principles/SKILL.md) | Hierarchy, typography, color, spacing, contrast — making output look designed. |
| [`styling-with-souls`](styling-with-souls/SKILL.md) | Themes: the built-in Deckard White, bootstrapping a soul from a brand, refining tokens. |
| [`charts-and-code`](charts-and-code/SKILL.md) | Adding data charts (`compile_chart`) and code blocks (`compile_code`). |
| [`validating-and-exporting`](validating-and-exporting/SKILL.md) | The StyleScore (contrast/overflow/structure) and delivering the `.pptx`. |

## Installing them

The skills are plain `SKILL.md` files with standard frontmatter (`name`,
`description`). Install them wherever your agent loads skills from:

- **Claude Code / Agent SDK:** copy each folder into your skills directory
  (e.g. `~/.claude/skills/` or your project's `.claude/skills/`).
- **Other agents:** point your skill/instruction loader at this `skills/`
  directory, or paste the relevant `SKILL.md` into your system prompt.

They reference Deckard's real tool names, so they stay accurate as long as you're
on a matching server version.
