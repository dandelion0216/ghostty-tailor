## What and why

<!-- What does this change, and why? Link the issue if there is one (e.g. Closes #1). -->

## How I checked it

<!-- e.g. go test ./..., wails build, tried the Colors tab with Ghostty 1.4.0 on macOS -->

## Checklist

- [ ] `go vet ./...` and `go test ./...` pass
- [ ] The app builds (`wails build` or `go build -tags desktop,production`)
- [ ] UI text is added in both English and Japanese (`I18N` in `frontend/dist/app.js`, `messages` in `i18n.go`)
- [ ] README (English and Japanese) is updated if behavior changed
