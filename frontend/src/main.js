// Elm owns the DOM. This file only moves messages:
//
//   Elm --port--> here --bindings--> Go
//
// and animates what Elm renders.

import gsap from "gsap";
import { Binaries } from "../bindings/github.com/d3uceY/stdscout/backend";
import { playSprite } from "./sprite.js";

// The brand idle: six 32x32 frames on a 192x32 strip. Frames 2 and 5 snap
// quicker, which is the little head-flick in the loop.
const BRAND_CLIP = {
  sheet: { src: "/brand/idle.png", frames: 6, frameWidth: 32, frameHeight: 32 },
  frameMs: 150,
  timing: [
    { frames: 2, frameMs: 110 },
    { frames: 5, frameMs: 110 },
  ],
};

const root = document.getElementById("app");

if (!window.Elm) {
  root.textContent = "public/elm.js is missing — run `npm run elm:build`.";
} else {
  const app = window.Elm.Main.init({ node: root, flags: null });

  // One handler per action: adding a binding is one entry, not another branch.
  const actions = {
    async list() {
      const binaries = await Binaries.List();
      // Plain objects only: port payloads are records, not class instances.
      app.ports.binariesFromJs.send(
        binaries.map(({ name, path, logo, shadowed, system }) => ({
          name,
          path,
          logo,
          shadowed,
          system,
        })),
      );
      requestAnimationFrame(() => fade(".row", { y: 6, stagger: 0.012 }));
    },
    async locations() {
      const places = await Binaries.Locations();
      app.ports.locationsFromJs.send(
        places.map(({ path, label, logo, count, system, dead }) => ({
          path,
          label,
          logo,
          count,
          system,
          dead,
        })),
      );
      requestAnimationFrame(() => fade(".folder", { y: 8, stagger: 0.01 }));
    },
    reveal: (path) => Binaries.Reveal(path),
  };

  app.ports.request.subscribe(async ({ action, target }) => {
    const handler = actions[action];
    if (!handler) {
      app.ports.errorFromJs.send(`unknown action "${action}"`);
      return;
    }
    try {
      await handler(target);
    } catch (err) {
      app.ports.errorFromJs.send(describe(err));
    }
  });

  // Only the top bar animates on launch: fading the content from zero just makes
  // the window look empty for half a second. Rows fade in when data arrives.
  fade(".topbar", { y: -6, stagger: 0 });

  // Elm renders the empty crop box; the player owns everything inside it.
  playSprite(document.querySelector(".brand-logo"), BRAND_CLIP);
}

// Wails rejects with a bare object in browser preview, so err.message is empty.
function describe(err) {
  if (!err) return "unknown error";
  if (typeof err === "string") return err;
  if (err.message) return err.message;
  const json = JSON.stringify(err);
  return json === "{}" ? "backend unavailable" : json;
}

// `amount` caps the whole stagger, otherwise 900 rows animate for ten seconds.
// The reduced-motion check lives inside the call, not in a module const: fade()
// runs at startup, before any const declared below this point would exist.
function fade(selector, { y = 0, stagger = 0.02 } = {}) {
  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;

  gsap.from(selector, {
    opacity: 0,
    y,
    duration: 0.35,
    ease: "power2.out",
    stagger: { each: stagger, amount: 0.4 },
    clearProps: "opacity,transform",
  });
}
