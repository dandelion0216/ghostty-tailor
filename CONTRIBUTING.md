# Contributing to ghostty-tailor

Thanks for your interest! Bug reports, ideas and pull requests are all welcome. English or Japanese is fine for issues and pull requests (日本語でも大丈夫です).

ghostty-tailor is an unofficial project and is not affiliated with Ghostty. Please report problems with Ghostty itself to the [Ghostty project](https://github.com/ghostty-org/ghostty), not here.

## Where to start

- Look at the [open issues](https://github.com/dandelion0216/ghostty-tailor/issues). Issues labeled `good first issue` or `help wanted` are good places to start.
- For larger changes, please open an issue first so we can agree on the approach.

## Development setup

Requirements: [Go](https://go.dev), the [Wails v2 CLI](https://wails.io) and [Ghostty](https://ghostty.org) 1.3 or later (1.2 or later on Linux). Node.js is not needed — the frontend is plain HTML / CSS / JS in `frontend/dist/`.

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

## Updating the Japanese doc translations

When a new Ghostty version changes option docs, those options fall back to English until they are translated again. With the new Ghostty installed:

```sh
go run -tags i18nsync . status            # how many entries are stale, missing or obsolete
go run -tags i18nsync . export todo.json  # stale and missing entries to translate
# fill in "ja" for each entry ("previous" is the outdated translation, if any)
go run -tags i18nsync . merge todo.json   # write them into i18n/docs.ja.json
go run -tags i18nsync . prune             # drop entries for options that no longer exist
```

`merge` skips entries left empty and entries whose English changed again after export. Keep one line per paragraph, separated by a blank line, as in the English doc.

## When Ghostty releases a new version

With the new Ghostty installed:

1. Run the app against it and go through each tab; check that changes are saved and reloaded.
2. `go run -tags i18nsync . status`, then update the translations as above.
3. Compare the options with the previous version (`ghostty +show-config --default --docs`): new options may need a category, a curated tab or a known type, and removed ones may still be used in a curated tab.
4. Compare the keybind actions (`ghostty +list-actions`) with `ACTIONS` in `frontend/dist/app.js`.
5. Update the tested version under "Supported Ghostty versions" in both READMEs.

## Compatibility

From 1.0 on, the files and names below are what users rely on, so they stay as they are. A change that cannot keep them is a breaking change and needs a major version.

| What | Format | Kept from 1.0 on |
|---|---|---|
| Editing the Ghostty config | Only the lines for the changed option are rewritten, in place; other lines and comments are left as they are. A new option is added at the end as `key = value`. | No reordering or reformatting of the user's file. |
| First backup | `<config file>.bak-tailor` next to the config, made once on first launch and never overwritten | Name and "never overwritten". |
| History | `<user config dir>/ghostty-tailor/history/`, one file per change named `YYYYMMDD-HHMMSS.mmm.ghostty` (local time), holding the whole config file as it was before the change; the newest 50 are kept | Location, file names and contents. History left by an older version stays readable. The number kept may change. |
| `settings.json` | `<user config dir>/ghostty-tailor/settings.json`, a JSON object; today only `lang` (`en` or `ja`) | Fields are only added, never renamed or removed. Unknown fields are ignored and missing ones fall back to defaults. |
| Export | A plain copy of the Ghostty config file | Stays a plain Ghostty config, never a format of our own. |
| Command and install | `ghostty-tailor`, `brew install dandelion0216/tap/ghostty-tailor` | Unchanged. |

`<user config dir>` is `~/Library/Application Support` on macOS and `$XDG_CONFIG_HOME` (usually `~/.config`) on Linux.

## Releases

Releases are tagged `vX.Y.Z` and installed through the [Homebrew tap](https://github.com/dandelion0216/homebrew-tap), which builds from the tagged source.
