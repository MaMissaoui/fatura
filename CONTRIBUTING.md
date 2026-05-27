# Contributing to Fatura

Please take a moment to review this document in order to make the contribution process easy and effective for everyone involved.

## Using the issue tracker

[GitHub Issues](https://github.com/MaMissaoui/fatura/issues) is the preferred channel for [bug reports](#bug-reports), [feature requests](#feature-requests) and [submitting pull requests](#pull-requests).

Please respect the following restrictions:

- Please **do not** use the issue tracker for personal support requests (email [ma.missaoui@gmail.com](mailto:ma.missaoui@gmail.com)).
- Please **do not** derail or troll issues. Keep the discussion on topic and respect the opinions of others.

## Bug Reports

A bug is a _demonstrable problem_ that is caused by the code in the repository. Good bug reports are extremely helpful — thank you!

Guidelines for bug reports:

1. **Use the GitHub issue search** — check if the issue has already been reported.
2. **Check if the issue has been fixed** — try to reproduce it using the latest `main` branch.
3. **Demonstrate the problem** — provide clear steps that can be reproduced.

A good bug report should not leave others needing to chase you up for more information. Please try to be as detailed as possible: What is your OS and version? What steps will reproduce the issue? What did you expect to happen?

## Feature Requests

Feature requests are welcome. Take a moment to find out whether your idea fits with the scope and aims of the project (local-first, privacy-focused invoicing and time tracking). Please provide as much detail and context as possible.

## Pull Requests

Good pull requests — patches, improvements, new features — are a fantastic help. They should remain focused in scope and avoid containing unrelated commits.

**Please ask first** before embarking on any significant pull request (e.g. implementing features, refactoring code), otherwise you risk spending time on something that might not be accepted.

### Development setup

See [DEVELOPMENT.md](DEVELOPMENT.md) for full setup instructions.

Quick start:

```bash
git clone https://github.com/MaMissaoui/fatura.git
cd fatura
go mod download
pnpm install
wails dev
```

### Commit style

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add invoice duplication
fix: correct tax rate default toggle
docs: update DEVELOPMENT.md
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`

**IMPORTANT**: By submitting a patch, you agree to allow the project owner to license your work under the same license as that used by the project ([GPLv3](LICENSE)).
