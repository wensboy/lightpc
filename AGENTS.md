# AGENTS.md - Engineering Constraints for This Project

> Core purpose: Make this repository's development conventions clear to AI agents so they do not drift off course.
>
> **AI agents must not edit this file directly. For more details, see the content under `docs/ref/`.**

## Project Overview

lightpc is a cross-platform PC application launcher skeleton focused on thin-layer startup handling, built on `wails + web` technology.

### Features

- Exposes multiple interfaces, including `cli`, `sdk`, `app`, and `web`.
- Unified interface abstraction.
- CLI and PC startup.
- Configuration parsing and transparent invocation.

### Technology Stack

- Frontend: `node 24.20.0, pnpm 12.16.1, react 19, tailwindcss v4, vite, typescript, tanstack`, etc.
- Backend: `go 1.26.0`, etc.
- Desktop: `wails v3`

## Principles

- Readability first: Prioritize readability in every generation task. Avoid vague wording, disconnected context, gratuitous jargon, and overly compressed terminology; present everything plainly.
- Strong semantics: When a term can be expressed in full, do not abbreviate or compress it. This applies to code identifiers, file names, field names, and similar cases.
- Adaptive handling: Do not let the same task fail repeatedly (>= 3 times), and never make pointless trial-and-error attempts. After several approaches fail, switch strategies.
- Precise targeting: Never load all related content at once. Always expand the scope through call-chain tracing so the problem can be pinpointed and fixed.

## Conventions

- Composition first: If a required feature can be achieved by composing existing libraries and fully meets expectations, do not implement it yourself.
- Strict comments: Comments are allowed only at key decisions, in doc comments, and at file tops; clear code is far better than comments.
- CamelCase naming: Unless otherwise specified, use camelCase consistently — lowerCamelCase for fields and UpperCamelCase for types.
- **Less is more: Do not introduce entities unless necessary.**
- Complete information: When an operation cannot be performed and the retry limit is reached, offer the user a manual action that includes all key details, the reason for the action, and the relevant commands.
- Snowball effect: Assume every decision has future consequences. Favor decisions with strong compatibility, avoid overthinking, and keep a short-term view of the future.
- Effective reuse: When there is a sound reason to extract a piece of logic (currently referenced in >= 2 places, or needed for future compatibility), extract it into independent logic and reuse it.
- Explicitness first: For both types and operations, **explicit >> implicit** behavior.
- Clean up after yourself: Remove all temporary test artifacts once testing is done, keep the state consistent before and after tests, and keep the file tree clean.
- Data sanitization: Sanitize sensitive data such as access keys and tokens through environment variables to avoid exposing private information publicly.
- Commit conventions: Follow Git commit message best practices, adding domain information with `()` when necessary.
- Compressed information: Since content depends on a dynamic environment, do not record actual data; fetch it live through commands. When the command list grows large enough, merge small command sets into a script to reduce calls.

## Commands

```sh
# This section will be filled in during development. AI agents may only record entries here with the user's permission, strictly following this format:
# Format:
# <id - number>. <description>
# <command>
# ids increment
# Command pairs are separated by \n
# Example:
# 1. Build the project
task build
###########################
```

## Must Do

- Commit completed milestones to local version control; generate a semantic summary from a diff analysis and use it as the commit message.
- Verify the environment thoroughly so that environment issues do not cause failures that cannot be worked around.
- Check code style to ensure it conforms to the conventions (both format and logic), with no dead, redundant, or ineffective code and no hardcoding.
- Test-driven development: Unit-test key functions for both normal cases and extreme boundaries; run integration tests on call chains and end-to-end tests on the whole system. Treat tests as reusable assets. When a test fails, suspect the code first (if the logic is unchanged); when the code logic changes, update the tests accordingly.
- Automatically sync pitfalls into the designated section of the designated document, for example `## Notes\n xxx`.
- For every task involving the frontend or backend, always read the corresponding rules under `Methods and Conventions`.

## Must Not Do

- Automatically sync with remote VCS repositories.
- Automatically install third-party libraries without permission.
- Automatically flesh out design document details and start coding immediately.

## Further Reading

### Methods and Conventions

- [Project Structure](./docs/ref/project-structure.md)
- [Backend Rules](./docs/ref/backend-rules.md)
- [Frontend Rules](./docs/ref/frontend-rules.md)
- [Migrations](./docs/ref/migrations.md)

### Features

- [bootstrap](./docs/feature/01-bootstrap.md)
- [config](./docs/feature/02-config.md)
- [log](./docs/feature/03-log.md)
- [engine runtime](./docs/feature/04-engine-runtime.md)
- [bus](./docs/feature/05-bus.md)