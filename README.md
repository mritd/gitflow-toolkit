# GitFlow Toolkit

GitFlow Toolkit is a CLI tool written in Go for standardizing git commit messages following the [Angular commit message specification](https://docs.google.com/document/d/1QrDFcIiPjSLDn3EL15IJygNPiHORgU1_OOAqWjiDU5Y/edit#heading=h.greljkmo14y0). It provides an interactive TUI for creating commits, branches, and managing git operations.

## Features

- Interactive commit message creation with type, scope, subject, body, and footer
- Automatic `Signed-off-by` generation
- Git subcommand integration (`git ci`, `git ps`, `git feat`, etc.)
- Commit message validation via git hooks
- Adaptive terminal width support for light and dark themes

## Installation

Download the latest binary from the [Release page](https://github.com/mritd/gitflow-toolkit/releases) and run the install command:

```bash
# Download (replace with your platform: linux-amd64, darwin-arm64, etc.)
wget https://github.com/mritd/gitflow-toolkit/releases/download/v2.2.0/gitflow-toolkit-darwin-arm64
chmod +x gitflow-toolkit-darwin-arm64

# Install (creates symlinks for git subcommands)
sudo ./gitflow-toolkit-darwin-arm64 install

# Optional: Install with commit-msg hook for validation
sudo ./gitflow-toolkit-darwin-arm64 install --hook
```

Or install via Go:

```bash
go install github.com/mritd/gitflow-toolkit/v2@latest
```

## Commands

| Command             | Description                                    |
|---------------------|------------------------------------------------|
| `git ci`            | Interactive commit message creation            |
| `git ps`            | Push current branch to remote                  |
| `git feat NAME`     | Create branch `feat/NAME`                      |
| `git fix NAME`      | Create branch `fix/NAME`                       |
| `git hotfix NAME`   | Create branch `hotfix/NAME`                    |
| `git docs NAME`     | Create branch `docs/NAME`                      |
| `git style NAME`    | Create branch `style/NAME`                     |
| `git refactor NAME` | Create branch `refactor/NAME`                  |
| `git chore NAME`    | Create branch `chore/NAME`                     |
| `git perf NAME`     | Create branch `perf/NAME`                      |
| `git test NAME`     | Create branch `test/NAME`                      |

## Commit Message Format

The tool enforces the Angular commit message format:

```
type(scope): subject

body

footer

Signed-off-by: Name <email>
```

Supported types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `hotfix`

## Configuration

| Environment Variable        | Description                              | Default |
|-----------------------------|------------------------------------------|---------|
| `GIT_SSH_STRICT_HOST_KEY`   | Enable strict SSH host key checking      | `false` |

## Uninstall

```bash
sudo gitflow-toolkit uninstall
```

## License

MIT
