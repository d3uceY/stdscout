# stdscout

Wails 3 desktop skeleton: **Go backend · Elm UI · JavaScript bridge · GSAP motion.**

Elm owns the DOM and never talks to Wails. JavaScript owns the wiring and never
builds markup. Go owns the state.

```
┌──────────────┐   ports    ┌──────────────┐   bindings   ┌──────────────┐
│     Elm      │ ─────────► │  src/main.js │ ───────────► │      Go      │
│  (renders)   │ ◄───────── │  (wiring)    │ ◄─────────── │  (services)  │
└──────────────┘   ports    └──────────────┘   events     └──────────────┘
```

## What the demo proves

| Direction | Mechanism | Visible as |
| --- | --- | --- |
| Elm → JS | `command` port | button click |
| JS → Go | `CounterService.Increment/Reset/State` bindings | counter value changes |
| Go → JS | `notice` event (`Events.On`) | "event stream" rows |
| JS → Elm | `stateFromJs` / `logFromJs` ports | re-rendered count + log |
| Elm → JS → GSAP | `animate` port | digit bump, new log row slide-in |

## Layout

```
main.go                     app setup, window, `notice` event emitter
counterservice.go           the only Go service; all state lives here
build/                      Wails build config, icons, per-OS taskfiles
                            (windows, darwin, linux — desktop targets only)
frontend/
  elm.json                  Elm app manifest (source dir: elm/src)
  elm/src/Main.elm          model / update / view + port declarations
  src/main.js               Elm ports ⇄ Wails bindings/events, GSAP
  index.html                mount point only
  public/
    style.css               all styling
    elm.js                  compiled Elm — generated, gitignored
  bindings/                 generated JS bindings — do not edit
```

## Commands

```bash
wails3 dev              # hot reload, Vite on :9245
wails3 build            # compile Elm + Vite + Go -> bin/stdscout.exe
wails3 task package     # installer
npm run elm:build       # compile Elm only (in frontend/)
npm run elm:format      # elm-format
```

`npm run build` and `npm run dev` both compile Elm first, so `public/elm.js` is
never stale.

## Adding a round trip

1. **Go** — add an exported method on `CounterService` in `counterservice.go`.
2. **Bindings** — `wails3 generate bindings -clean=true`.
3. **JS** — add one line to the `actions` table in `frontend/src/main.js`.
4. **Elm** — add the action name to the button's `Requested "..."` value.

Adding a *new* channel needs three edits: a `port` declaration in
`Main.elm`, a `subscribe` (or `send`) in `main.js`, and a message in `update`.

## Where the wiring lives

- **Ports must be declared in a `port module`.** `Main.elm` starts with
  `port module Main exposing (main)`.
- **Outgoing port messages reach JS after the DOM patch** — `main.js` wraps GSAP
  in `requestAnimationFrame` for that reason.
- **Port payloads are records of primitives.** `CounterState` and `LogEntry` are
  type aliases, not custom types, so they cross the boundary as-is.
- **`public/elm.js` is a classic script**, not an ES module: it assigns
  `window.Elm`, which `main.js` uses after the module loads.
