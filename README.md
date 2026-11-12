# ghostty-tailor

A GUI for tailoring your [Ghostty](https://ghostty.org) configuration — browse every option, tweak it, and see the change in your running terminal right away.

> **Status:** early design stage. Nothing to install yet.

> ghostty-tailor is an unofficial community project and is not affiliated with the Ghostty project.

## Why

Ghostty is configured with a plain text file. It is powerful, but it is hard to know which options exist, what they do, and what values they accept. ghostty-tailor puts all of that on one screen.

## Planned features

- Browse and search every config option, with the documentation Ghostty ships with
- Pick themes, colors and fonts visually
- Apply changes to the running Ghostty instantly (live reload)
- Keep your existing config file readable — only the options you change are written

## Planned stack

- [Go](https://go.dev) + [Wails](https://wails.io)
- macOS and Linux
- Distributed via Homebrew

## License

[MIT](LICENSE)
