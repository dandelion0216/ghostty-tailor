# Contributing to ghostty-tailor

Thanks for your interest! Bug reports, ideas and pull requests are all welcome. English or Japanese is fine for issues and pull requests (日本語でも大丈夫です).

ghostty-tailor is an unofficial project and is not affiliated with Ghostty. Please report problems with Ghostty itself to the [Ghostty project](https://github.com/ghostty-org/ghostty), not here.

## Where to start

- Look at the [open issues](https://github.com/dandelion0216/ghostty-tailor/issues). Issues labeled `good first issue` or `help wanted` are good places to start.
- For larger changes, please open an issue first so we can agree on the approach.

## Development setup

Requirements: [Go](https://go.dev), the [Wails v2 CLI](https://wails.io) and [Ghostty](https://ghostty.org) 1.x. Node.js is not needed — the frontend is plain HTML / CSS / JS in `frontend/dist/`.

```sh
go install github.com/wailsapp/wails/v2/cmd/wails@latest
git clone https://github.com/dandelion0216/ghostty-tailor.git
cd ghostty-tailor
wails dev           # run with live reload of the frontend
wails build -clean  # build build/bin/ghostty-tailor.app
```

Your real Ghostty config is edited while you test. ghostty-tailor backs it up to `config.ghostty.bak-tailor` on first launch, but you may want to keep your own copy too.

## Project layout

| Path | What it does |
|---|---|
| `main.go`, `app.go` | Wails entry point and the methods the frontend calls |
| `ghostty.go` | Runs the `ghostty` CLI, parses option docs, themes and fonts, live reload |
| `config.go` | Reads and edits the config file line by line, keeping comments |
| `i18n.go`, `i18n/docs.ja.json` | Language setting, backend messages, Japanese doc translations |
| `frontend/dist/` | The UI (`app.js` holds the UI text in `I18N`) |
| `packaging/` | App icon and Info.plist used by the Homebrew formula |

## Guidelines

- Keep the user's config file safe: only rewrite the lines for options that change, and keep comments and order.
- Add UI text in both English and Japanese.
- Japanese translations of option docs: polite form (です・ます), keep code, values and option names in English, and keep an English term when a Japanese rendering would sound unnatural. Each entry stores a hash of the English source; see `i18n.go`.
- Run `gofmt`, `go vet ./...` and `go test ./...` before opening a pull request.

## Releases

Releases are tagged `vX.Y.Z` and installed through the [Homebrew tap](https://github.com/dandelion0216/homebrew-tap), which builds from the tagged source.
