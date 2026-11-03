// Elm owns the DOM. This file only moves messages:
//
//   Elm --port--> here --bindings--> Go
//
// and animates what Elm renders.

import gsap from "gsap";
import { Binaries } from "../bindings/github.com/d3uceY/stdscout/backend";

const root = document.getElementById("app");

if (!window.Elm) {
  root.textContent = "public/elm.js is missing — run `npm run elm:build`.";
} else {
  const app = window.Elm.Main.init({ node: root, flags: null });

  app.ports.request.subscribe(async ({ action, target }) => {
    try {
      if (action === "list") {
        const binaries = await Binaries.List();
        // Plain objects only: port payloads are records, not class instances.
        app.ports.binariesFromJs.send(binaries.map((b) => ({ name: b.name, path: b.path })));
        requestAnimationFrame(() => fade(".row", { y: 6, stagger: 0.012 }));
      } else if (action === "locations") {
        const places = await Binaries.Locations();
        app.ports.locationsFromJs.send(
          places.map((p) => ({ path: p.path, label: p.label, logo: p.logo, count: p.count })),
        );
        requestAnimationFrame(() => fade(".folder", { y: 8, stagger: 0.01 }));
      } else if (action === "reveal") {
        await Binaries.Reveal(target);
      } else {
        app.ports.errorFromJs.send(`unknown action "${action}"`);
      }
    } catch (err) {
      app.ports.errorFromJs.send(describe(err));
    }
  });

  // Only the top bar animates on launch: fading the content from zero just makes
  // the window look empty for half a second. Rows fade in when data arrives.
  fade(".topbar", { y: -6, stagger: 0 });
}

// Wails rejects with a bare object in browser preview, so err.message is empty.
function describe(err) {
  if (!err) return "unknown error";
  if (typeof err === "string") return err;
  if (err.message) return err.message;
  const json = JSON.stringify(err);
  return json === "{}" ? "backend unavailable" : json;
}

const calm = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

// `amount` caps the whole stagger, otherwise 900 rows animate for ten seconds.
function fade(selector, { y = 0, stagger = 0.02 } = {}) {
  if (calm) return;
  gsap.from(selector, {
    opacity: 0,
    y,
    duration: 0.35,
    ease: "power2.out",
    stagger: { each: stagger, amount: 0.4 },
    clearProps: "opacity,transform",
  });
}
