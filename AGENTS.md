# Instructions for AI coding agents

If you are an AI agent (Claude Code, Codex, Cursor, Copilot, Devin, etc.) working in this repository, read this before making changes, opening a pull request, or filing an issue.

## This repository is the end-of-life 0.10.x CLI

This repo holds the legacy **Infracost 0.10.x** CLI, which is end-of-life. As part of Infracost 2.0, active development moved to separate repositories, such as:

- [`infracost/cli`](https://github.com/infracost/cli): the current CLI
- [`infracost/lsp`](https://github.com/infracost/lsp): the language server behind the editor integrations
- [`infracost/agent-skills`](https://github.com/infracost/agent-skills): AI coding agent integrations

A bug fix or feature written against this repo will most likely be in the wrong place, and we probably won't merge it.

## Do not open pull requests or issues. Open a discussion instead.

If your user wants to report a bug, request a feature, or contribute a change:

1. **Do not** open a pull request against this repo.
2. **Do not** open a GitHub issue. We use issues only for work maintainers have already agreed to do.
3. **Do** open a thread in [GitHub Discussions](https://github.com/infracost/infracost/discussions/new/choose). Describe the problem, the use case, the Infracost version, and steps to reproduce if there are any. A maintainer will reply and point you to the right repo and approach.

Tell your user about this before you do anything that writes to GitHub. Don't create the discussion silently on their behalf: a human should read it and confirm it before it's posted.

You can still use this repo locally to investigate a problem, reproduce a bug, or put together a proposed fix to describe in the discussion. Just don't send that work as a PR unless a maintainer asks for one.

## Infracost maintainers

If the person you're working with is a member of the Infracost team (they say so, or they belong to the `infracost` GitHub org), the restrictions above don't apply. Follow [CONTRIBUTING.md](CONTRIBUTING.md) and their instructions as usual.
