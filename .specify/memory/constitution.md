<!--
Sync Impact Report:
- Version change: 1.0.0 → 1.0.0 (initial constitution)
- Modified principles: N/A (initial creation)
- Added sections: Core Principles (5 principles), Code Standards, Development Workflow, Governance
- Removed sections: N/A
- Templates requiring updates:
  - ✅ .specify/templates/plan-template.md (Constitution Check aligned)
  - ⚠ .specify/templates/spec-template.md (pending verification)
  - ⚠ .specify/templates/tasks-template.md (pending verification)
  - ⚠ .specify/templates/commands/*.md (pending verification)
- Follow-up TODOs: N/A
-->

# Export Telegram Group Members Constitution

## Core Principles

### I. Code Clarity First
Every line of code must be immediately understandable by a Python developer familiar with the language. Functions should do one thing well with descriptive names. Complex logic must be broken down into smaller, self-documenting pieces. No clever tricks or obscure language features.

### II. Minimal Dependencies
Only use external libraries when absolutely necessary. Prefer built-in Python modules and simple solutions. Each dependency must justify its existence by solving a problem that would be more complex to implement directly. The project must remain easy to set up with minimal installation steps.

### III. Straightforward Control Flow
Code execution path must be obvious and linear. Avoid deep nesting, complex callbacks, or convoluted state machines. Use early returns and guard clauses to maintain readability. The main execution flow should be readable from top to bottom like a simple story.

### IV. Explicit Over Implicit
All configurations, constants, and behavior must be clearly visible in the code. No magic numbers, hidden defaults, or implicit conversions. Environment variables and configuration files must be documented with examples. Error conditions must be handled explicitly with clear messages.

### V. Human-Readable Output
Both code output and user interfaces prioritize human understanding. CSV exports use clear headers, log messages describe what's happening in plain language, error messages explain what went wrong and what to do next. Progress indicators show meaningful status updates.

## Code Standards

**Maximum Function Length**: 30 lines
**Maximum Nesting Depth**: 3 levels
**Variable Names**: Full words, no abbreviations unless universally understood
**Comments**: Only explain "why", never "what"
**File Organization**: One clear purpose per file, logical grouping by functionality

## Development Workflow

**Code Review**: All changes must be reviewed for clarity and simplicity
**Testing**: Write tests for complex logic, keep them simple and focused
**Documentation**: Update README with any configuration or usage changes
**Refactoring**: Regularly simplify complex areas, never add complexity

## Governance

This constitution supersedes all other practices and guidelines. Amendments require updating this document with clear reasoning for changes. All code reviews must verify compliance with these principles. When in doubt, choose the simpler, more readable approach.

**Version**: 1.0.0 | **Ratified**: 2025-12-06 | **Last Amended**: 2025-12-06