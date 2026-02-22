# Contributing to Nibble

This project is open to contributions!

## Getting Started

Fork the repository on GitHub.

**Dev container (recommended):** Open the repo in VS Code and choose "Reopen in Container" when prompted. Everything is ready to go.

| Command | Description |
|---|---|
| `go run .` | To run the current code |
| `go run . -demo` | To run with standardized demo interfaces and network hosts (demo wifi interface has 50 hosts for stress testing) |
| `make build` | Build the `nibble` binary |
| `make fix` | Format, vet, and fix the codebase |
| `make demo` | Generate `demo.gif` using VHS |

## Making Changes

- Open or select an issue before starting significant work so we can discuss the approach.
- Keep PRs focused, one feature or fix per PR.
- Run `make fix` before submitting to ensure code is formatted and vetted.
- If you're adding a feature, update the README if it affects usage or hotkeys.
- Don't forget to keep the help.go views up to date

## Project Layout

```
main.go              Entry point
internal/
  scan/              Port scanning and banner grabbing
  ports/             Port list management
  tui/               Bubble Tea UI (views, models, rendering)
  demo/              Demo mode code
```

## Submitting a PR

1. Test manually all affected features, and verify your change works with no unintended side effects.
2. Open a pull request against `main` with a clear description of what is added/fixed and why.
