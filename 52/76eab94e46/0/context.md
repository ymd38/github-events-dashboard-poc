# Session Context

## User Prompts

### Prompt 1

frontend-ui-uxなどUI/UX、デザインのスキルで @frontend/ のデザインを評価し改善して下さい

### Prompt 2

Base directory for this skill: /Users/hirokazuyamada/.claude/skills/frontend-ui-ux

# Frontend UI/UX Command

Routes to the designer agent or Gemini MCP for frontend work.

## Usage

```
/oh-my-claudecode:frontend-ui-ux <design task>
```

## Routing

### Preferred: MCP Direct
Before first MCP tool use, call `ToolSearch("mcp")` to discover deferred MCP tools.
Use `mcp__g__ask_gemini` with `agent_role: "designer"` for design tasks.
If ToolSearch finds no MCP tools, use the Claude agent fallback be...

### Prompt 3

commitしてpushしてください

