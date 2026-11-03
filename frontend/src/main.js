// ─── The bridge ─────────────────────────────────────────────────────────────
//
//   Elm (owns the DOM)
//     │  ports: command / animate            (Elm → JS)
//     │  ports: stateFromJs / logFromJs      (JS  → Elm)
//     ▼
//   this file
//     │  Wails bindings: CounterService.*    (JS → Go, call/response)
//     │  Wails events:   "notice"            (Go → JS, push)
//     ▼
//   Go
//
// Nothing here builds markup — Elm does that. Nothing in Elm knows Wails
// exists. Keep it that way and either side can be swapped out.

import { Events } from "@wailsio/runtime";
import gsap from "gsap";
import { CounterService } from "../bindings/github.com/d3uceY/stdscout";

const stage = document.getElementById("app");

if (!window.Elm) {
  stage.textContent =
    "public/elm.js is missing — run `npm run elm:build` before starting the app.";
} else {
  boot(
    window.Elm.Main.init({
      node: stage,
      flags: { startedAt: clock() },
    }),
  );
}

function clock() {
  return new Date().toTimeString().slice(0, 8);
}

function boot(app) {
  wirePorts(app);
  wireEvents(app);
  intro();
}

// ─── Elm → JS → Go ──────────────────────────────────────────────────────────

// One table instead of a switch: adding a command is one line, and Elm already
// decides which actions exist.
const actions = {
  sync: () => CounterService.State(),
  increment: (amount) => CounterService.Increment(amount),
  decrement: (amount) => CounterService.Increment(-amount),
  reset: () => CounterService.Reset(),
};

function wirePorts(app) {
  app.ports.command.subscribe(async ({ action, amount }) => {
    const run = actions[action];
    if (!run) {
      log(app, "warn", `unknown action "${action}"`);
      return;
    }

    try {
      const state = await run(amount);
      app.ports.stateFromJs.send({
        count: state.count,
        lastAction: state.lastAction,
        updatedAt: state.updatedAt,
      });
      log(app, "ok", `${action}(${amount}) → count = ${state.count}`);
    } catch (err) {
      // Most likely cause during plain `vite` development: no Go backend. Send a
      // zero snapshot anyway so Elm stops showing "waiting on go" forever.
      app.ports.stateFromJs.send({ count: 0, lastAction: "unavailable", updatedAt: clock() });
      log(app, "err", `go call failed: ${describe(err)}`);
    }
  });

  // Elm queues a port message *during* update and the DOM is patched after, so
  // wait one frame or GSAP animates the element that is about to be replaced.
  app.ports.animate.subscribe((target) => {
    requestAnimationFrame(() => {
      if (target === "count") {
        gsap.fromTo(
          "#count",
          { y: 10, scale: 0.94, opacity: 0.35 },
          { y: 0, scale: 1, opacity: 1, duration: 0.5, ease: "power3.out" },
        );
      }

      if (target === "log") {
        const newest = document.querySelector(".log-row");
        if (newest) {
          gsap.fromTo(
            newest,
            { x: -14, opacity: 0 },
            { x: 0, opacity: 1, duration: 0.32, ease: "power2.out" },
          );
        }
      }
    });
  });
}

// ─── Go → JS → Elm ──────────────────────────────────────────────────────────

function wireEvents(app) {
  Events.On("notice", (event) => {
    log(app, "info", event.data);
  });
}

function log(app, level, message) {
  app.ports.logFromJs.send({ at: clock(), level, message });
}

// Wails rejects with plain strings or bare objects, so `${err.message}` alone
// often renders as an empty line.
function describe(err) {
  if (!err) return "unknown error";
  if (typeof err === "string") return err;
  if (err.message) return err.message;
  const json = JSON.stringify(err);
  return json === "{}" ? "backend unavailable" : json;
}

// ─── GSAP: one intro timeline ───────────────────────────────────────────────

function intro() {
  gsap
    .timeline({ defaults: { ease: "power3.out" } })
    .from(".topbar > *", { y: -12, opacity: 0, duration: 0.45, stagger: 0.08 })
    .from(".panel", { y: 26, opacity: 0, duration: 0.55, stagger: 0.1 }, "-=0.2")
    .from(".footer > *", { opacity: 0, x: -6, duration: 0.3, stagger: 0.05 }, "-=0.35");
}
