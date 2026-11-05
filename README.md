<p align="center">
  <img src="docs/brand/idle.gif" width="128" alt="StdScout mascot">
</p>

# StdScout

Every executable on your PATH, and the folders that keep them. Windows only.

**Applications** — one row per runnable file, filterable by name or path. A duplicate is marked `shadowed` when an earlier folder wins, so you can see which copy typing the name actually runs.

**Locations** — the PATH folders, their binary counts, and the ones that no longer exist.

Windows' own binaries are hidden until you flip **System**.

## Install

[Releases](https://github.com/d3uceY/stdscout/releases) has a portable `.exe` and an NSIS installer. Both are unsigned: SmartScreen → **More info** → **Run anyway**.

## Build

Go 1.26+, Node 24+, Elm 0.19.3, Wails 3.

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0
cd frontend && npm ci && npm run build && cd ..
go build -o bin/stdscout.exe .
```

`wails3 task dev` runs the hot-reloading dev loop.

## Release

Push a `vX.Y.Z` tag. The workflow builds the Windows app, stamps the tag into the binary with `-X main.AppVersion`, and uploads the portable `.exe`, the installer, and `SHA256SUMS.txt`.
