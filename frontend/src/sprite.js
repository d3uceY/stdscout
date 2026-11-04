// Sprite sheet player, ported from vectile's `frontend/src/lib/sprite`.
//
// A sheet is one horizontal strip of frames drawn at one art-pixel scale. A
// clip says which frames play in what order and how long each is held. Frame
// numbers are 1-based and inclusive, the way the art is exported.
//
// Left behind on purpose: `content` crops, `anchor` boxes and `lift`. Those
// exist to keep differently-shaped mascot poses on one ground line. StdScout's
// sheet is a plain uniform strip, so every frame is framed the same way.
// Everything that decides *timing* is here, per-frame overrides included.

/** Time per frame when a clip sets neither `frameMs` nor a matching rule. */
export const SPRITE_FRAME_MS = 90;

/** Resolve a clip into steps: the frames that play once, then the frames that
    repeat. Each step carries its own hold time, so the player never does maths
    while a frame is on screen. */
export function framePlan(clip) {
  const total = Math.max(1, Math.floor(clip.sheet.frames));
  const play = clampRange(clip.play ?? [1, total], total);
  const intro = steps(clip, play, total);

  if (clip.loop === false) return { intro, loop: null };

  return { intro, loop: steps(clip, clampRange(clip.loop ?? play, total), total) };
}

/** How long one frame is held: per-frame durations from the Aseprite export win,
    then the clip's `frameMs`, then each timing rule in order, where the last
    rule covering the frame wins. */
export function frameMsFor(clip, frame, total) {
  const sheetFrames = Math.max(1, Math.floor(total ?? clip.sheet.frames));
  const exported = clip.frameMsByFrame?.[frame - 1];
  if (exported) return exported;

  let ms = clip.frameMs ?? SPRITE_FRAME_MS;

  for (const rule of clip.timing ?? []) {
    const [first, last] = clampRange(spanRange(rule.frames), sheetFrames);
    if (frame >= first && frame <= last) ms = rule.frameMs;
  }

  return ms;
}

/** Normalize any span to a first/last pair. */
export function spanRange(span) {
  return typeof span === "number" ? [span, span] : span;
}

/** The player currently mounted in a host, so a new clip can replace it. */
const players = new WeakMap();

/**
 * Play `clip` inside `host`, which must be a fixed-size, `overflow: hidden` box.
 *
 * The image is the whole strip and the player steps it with a transform, so the
 * work stays on the compositor and the pixel grid stays honest. Reduced motion
 * freezes on the first frame. A `loop: false` clip holds its last frame and the
 * returned promise resolves there, which is how a one-shot reaction hands the
 * box back to the idle loop.
 */
export function playSprite(host, clip) {
  if (!host) return Promise.resolve();
  stopSprite(host);

  // CSS owns the on-screen size; the engine just follows it, so one art pixel
  // never lands on a half screen pixel.
  const { frames, frameWidth, frameHeight } = clip.sheet;
  const width = host.clientWidth || frameWidth;
  const height = host.clientHeight || frameHeight;

  const img = document.createElement("img");
  img.src = clip.sheet.src;
  img.alt = "";
  img.draggable = false;
  img.style.width = `${frames * width}px`;
  img.style.height = `${height}px`;
  host.appendChild(img);

  const draw = (frame) => {
    img.style.transform = `translateX(${-(frame - 1) * width}px)`;
  };

  const plan = framePlan(clip);
  const first = plan.intro[0];
  draw(first.frame);

  let timer = 0;
  players.set(host, {
    stop() {
      clearTimeout(timer);
      img.remove();
    },
  });

  if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return Promise.resolve();

  return new Promise((resolve) => {
    // One timer at a time so every frame can hold for its own time.
    let step = 0;
    const tick = () => {
      const { intro, loop } = plan;
      const current =
        step < intro.length ? intro[step] : loop ? loop[(step - intro.length) % loop.length] : undefined;

      // Nothing left to show: the clip ran out, so it holds its last frame.
      if (!current) {
        resolve();
        return;
      }

      draw(current.frame);
      step++;
      timer = setTimeout(tick, current.frameMs);
    };

    tick();
  });
}

/** Take whatever is playing out of `host` and leave the box empty. */
export function stopSprite(host) {
  const player = host && players.get(host);
  if (!player) return;

  player.stop();
  players.delete(host);
}

function steps(clip, range, total) {
  return framesIn(range).map((frame) => ({ frame, frameMs: frameMsFor(clip, frame, total) }));
}

function framesIn([first, last]) {
  return Array.from({ length: last - first + 1 }, (_, i) => first + i);
}

/** Clamp a range to the sheet and order it, so `[8, 3]` reads as 3..8. */
function clampRange([first, last], total) {
  const a = clampFrame(first, total);
  const b = clampFrame(last, total);
  return a <= b ? [a, b] : [b, a];
}

function clampFrame(frame, total) {
  return Math.min(total, Math.max(1, Math.round(frame) || 1));
}
