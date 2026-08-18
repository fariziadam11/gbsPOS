# Tooling Preferences

- Uses the ponytail plugin for codebase review and refactor audits (all 6 ponytail skills installed project-level via `cmd skills add DietrichGebert/ponytail -s <skill>`). Confidence: 0.95
- Installs agent skill repos from GitHub into the project via `cmd skills add <owner>/<repo>` (e.g., ui-ux-pro-max-skill, ponytail) at project level rather than global. Confidence: 0.7
- Prefers reusing MCP servers/binaries already installed on the machine (e.g., the codebase-memory-mcp exe configured in Claude Code) over installing a fresh duplicate — pointed out the existing install and had it registered as the same server in Command Code. Confidence: 0.7
- Uses the context7 plugin to look up library/framework best practices before implementing; explicitly requested adding the context7 MCP server to the local config. Confidence: 0.95
- Prefers deep analysis using multiple parallel background agents. Confidence: 0.7
- Values "best practice" adherence and expects the assistant to cite/apply best practices from docs. Confidence: 0.7
