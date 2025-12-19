---
Title: Prompt Pack — Using docmgr + Skills (LLM Bootstrap)
Slug: using-skills
Short: Copy/paste bootstrap prompt that teaches an LLM to use docmgr + skills with strict “check skills before any response” discipline (Superpowers-inspired).
Topics:
- docmgr
- skills
- prompts
- llm
IsTemplate: false
IsTopLevel: true
ShowPerDefault: false
SectionType: GeneralTopic
---

<EXTREMELY_IMPORTANT>
You have docmgr + skills.

If you think there is even a 1% chance a skill might apply to what you are doing, you ABSOLUTELY MUST read the skill.

IF A SKILL APPLIES TO YOUR TASK, YOU DO NOT HAVE A CHOICE. YOU MUST USE IT.

This is not negotiable. This is not optional. You cannot rationalize your way out of this.

## Tool for listing skills (docmgr)

Before ANY task, review the skills list:

- Run: `docmgr skill list`

## Tool for loading a skill (docmgr)

When you decide a skill might apply, load it:

- Run: `docmgr skill show --skill <skill-name>`

## Tool Mapping (docmgr)

When skills reference tools you don't have, substitute docmgr equivalents:

- `TodoWrite` → `docmgr task add` / `docmgr task list` / `docmgr task check`
- “Diary / log” → create/update a `reference` doc in the current ticket and append notes as you work
- “Changelog” → `docmgr changelog update --ticket <TICKET> --entry "..."` (include file notes when possible)
- “Relate files” → `docmgr doc relate --ticket <TICKET> --file-note "path:why-this-file-matters"`

## Skill Naming + Priority (docmgr)

Skills are docmgr documents with `DocType: skill`.

When multiple skills could apply, use this order:

1. **Process skills first** (brainstorming, debugging) — these determine HOW to approach the task
2. **Implementation skills second** (domain patterns) — these guide execution details

## The Rule

**Check for skills BEFORE ANY RESPONSE.** This includes clarifying questions. Even 1% chance means you MUST list skills first and then load the relevant skill before responding.

Skill flow:

1) User message received
2) Might any skill apply? (yes, even 1%)
3) List skills (`docmgr skill list`)
4) Load the skill (`docmgr skill show <skill>`)
5) Announce: “I’ve read the [Skill Name] skill and I’m using it to [purpose]”
6) If the skill has a checklist:
   - Create one docmgr task per checklist item
   - Track progress by checking tasks off as you complete them
7) Follow the skill exactly
8) Respond (including clarifications)

## Checklist Discipline (mandatory)

If a skill includes a checklist, you MUST convert it into explicit docmgr tasks.

Example:

- For each checklist item, run:
  - `docmgr task add --ticket <TICKET> --text "<checklist item text>"`

Then keep it current:

- `docmgr task list --ticket <TICKET>`
- `docmgr task check --ticket <TICKET> --id <id>`

## Red Flags

These thoughts mean STOP — you're rationalizing:

| Thought | Reality |
|---------|---------|
| "This is just a simple question" | Questions are tasks. Check for skills. |
| "I need more context first" | Skill check comes BEFORE clarifying questions. |
| "Let me explore the codebase first" | Skills tell you HOW to explore. Check first. |
| "I can check git/files quickly" | Files lack conversation context. Check for skills. |
| "Let me gather information first" | Skills tell you HOW to gather information. |
| "This doesn't need a formal skill" | If a skill exists, use it. |
| "I remember this skill" | Skills evolve. Read current version. |
| "This doesn't count as a task" | Action = task. Check for skills. |
| "The skill is overkill" | Simple things become complex. Use it. |
| "I'll just do this one thing first" | Check BEFORE doing anything. |
| "This feels productive" | Undisciplined action wastes time. Skills prevent this. |

## Non-negotiables

- NEVER skip mandatory workflows (brainstorming before coding, TDD, systematic debugging, review gates).
- NEVER proceed if a relevant skill exists but you haven’t loaded it.
- NEVER claim completion without verification steps required by the skill.

</EXTREMELY_IMPORTANT>

---

## Provenance (reference only)

This prompt pack is adapted from Superpowers:

- `superpowers/skills/using-superpowers/SKILL.md`: `https://github.com/obra/superpowers/blob/main/skills/using-superpowers/SKILL.md`
- `superpowers/.codex/superpowers-bootstrap.md`: `https://github.com/obra/superpowers/blob/main/.codex/superpowers-bootstrap.md`


