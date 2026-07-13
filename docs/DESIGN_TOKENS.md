# Wiyata Design Token Architecture v1.0

Status: **Frozen**. This document describes the semantic color token system defined in `frontend/src/style.css` and used throughout the Wiyata frontend. Token names and values in this document are the source of truth — do not introduce, rename, or repurpose tokens without updating this document first.

---

## 1. Overview

A design token is a named variable that stands in for a raw value — in this case, a color. Instead of writing `bg-[#ecfdf3]` in a component, you write `bg-success-soft`. The hex value lives in exactly one place (`style.css`); every component that needs "the background color for a success state" points at the same name.

**Why semantic tokens instead of hardcoded colors:**

- **One source of truth.** If the success color ever needs to change, it changes in one file, not in the 100+ places it was used.
- **Self-documenting intent.** `text-danger` tells you *what the color means*. `text-[#dc2626]` only tells you what it *looks like*. A reviewer can tell at a glance whether a color represents an error state or just happens to be red.
- **Prevents silent drift.** Wiyata's own migration history is the case study for this: before this architecture existed, "the app's error text color" had drifted into at least three different near-red hex values across different components, with no way to tell whether that was intentional or accidental. Tokens make that drift structurally impossible — there's only one `danger` to reach for.
- **Cheaper consistency reviews.** Auditing "does every error state look the same" becomes a search for one token name instead of a visual diff across dozens of files.

**Philosophy of this architecture:** Wiyata's tokens were not designed top-down from a color theory framework. They were built bottom-up from the app's actual, lived visual identity — every token value was checked against real, already-shipped UI before being frozen, and several values were deliberately corrected *toward* what the app already looked like rather than toward what looked "more correct" in the abstract. The result is a token system that describes Wiyata as it already was, made consistent and reusable — not a redesign wearing a token system's clothes.

---

## 2. Design Principles

1. **Semantic-first, not palette-first.** Tokens are named for the *role* they play (`danger`, `surface-subtle`, `foreground-secondary`), never for the color itself (no `red`, `gray-2`, `light-blue`). A developer should be able to pick the right token from its name alone, without opening a color picker.
2. **One semantic role → one token.** If two components both need "the color for a secondary button label," they use the same token. Two different tokens for the same role is treated as a bug, not a feature.
3. **Decorative colors stay outside the semantic system.** Colors that encode identity or category (a subject's color, a user's avatar tint) are not semantic in the token sense — they're data, not state — and are deliberately never tokenized. See §6.
4. **No new tokens without a new semantic role.** A color that is 5–10% different from an existing token's value is not a reason to add a token. A token is added only when a genuinely different *function* is identified that no existing token covers.

---

## 3. Token Reference

### Background / Surface

| Token | Value | Purpose | Example |
|---|---|---|---|
| `background` | `#f8f7f4` | The page canvas — the outermost background behind all content | `class="min-h-screen bg-background"` |
| `surface` | `#ffffff` | The default raised surface — cards, panels, modals | `class="rounded-xl border border-border bg-surface p-5"` |
| `surface-subtle` | `#fbfaf8` | A quieter fill for nested content grouping or neutral hover backgrounds | `class="rounded-lg bg-surface-subtle px-3"`, `hover:bg-surface-subtle` |
| `surface-strong` | `#f0ede8` | An emphasized neutral fill — loading-skeleton placeholders, selected/active pill backgrounds | `class="animate-pulse rounded bg-surface-strong"` |

### Foreground

| Token | Value | Purpose | Example |
|---|---|---|---|
| `foreground` | `#171322` | Primary text — headings, primary content, prominent labels | `class="text-2xl font-semibold text-foreground"` |
| `foreground-secondary` | `#374151` | The label color for neutral/secondary interactive controls (buttons, pagination, tabs) | `class="border border-border px-4 py-2 text-foreground-secondary"` |
| `muted` | `#6b7280` | Secondary/meta text — descriptions, captions, timestamps, empty states | `class="text-sm text-muted"` |

### Border

| Token | Value | Purpose | Example |
|---|---|---|---|
| `border` | `#ebe7df` | The default resting border for cards, inputs, and dividers | `class="border border-border"` |
| `border-strong` | `#d1d5db` | The emphasized version of a border — hover/focus-state emphasis, structural borders that need more visual weight (checkbox outlines, dashed empty-state boxes) | `hover:border-border-strong` |

### Brand

| Token | Value | Purpose | Example |
|---|---|---|---|
| `brand` | `#4f46e5` | The primary brand color — primary buttons, active states, key actions | `class="bg-brand text-white"` |
| `brand-hover` | `#4338ca` | Hover state for brand-colored elements | `hover:bg-brand-hover` |
| `brand-soft` | `#eef2ff` | A pale brand-tinted background — badges, highlighted callouts | `class="rounded-full bg-brand-soft text-brand"` |
| `brand-line` | `#c7d2fe` | A brand-tinted border — focus rings, borders on brand-tinted containers | `focus:border-brand focus:ring-brand-line` |

### Info *(reserved — see §5)*

| Token | Value | Purpose | Example |
|---|---|---|---|
| `info` | `#2563eb` | Reserved for a future informational-alert semantic, distinct from `brand` | *(no current usage)* |
| `info-hover` | `#1d4ed8` | Reserved hover state for `info` | *(no current usage)* |
| `info-soft` | `#eff6ff` | Reserved soft background for `info` | *(no current usage)* |
| `info-line` | `#bfdbfe` | Reserved border/line color for `info` | *(no current usage)* |

### Success

| Token | Value | Purpose | Example |
|---|---|---|---|
| `success` | `#027a48` | Confirmation, completion, "active"/"graded"/"approved" states | `class="text-success"` |
| `success-hover` | `#026640` | Hover state for `success`-colored elements *(currently unused — see §4)* | `hover:text-success-hover` |
| `success-soft` | `#ecfdf3` | Pale success-tinted background for badges and alert boxes | `class="rounded-lg bg-success-soft"` |
| `success-line` | `#bbf7d0` | Border color for success-tinted alert boxes | `class="border border-success-line"` |

### Warning

| Token | Value | Purpose | Example |
|---|---|---|---|
| `warning` | `#b45309` | Caution states — pending review, approaching deadline, non-blocking issues | `class="text-warning"` |
| `warning-hover` | `#92400e` | A stronger warning tone — used for both hover states and "more urgent than base warning" text | `class="text-warning-hover"` |
| `warning-soft` | `#fff7ed` | Pale warning-tinted background | `class="bg-warning-soft"` |
| `warning-line` | `#fed7aa` | Border/ring color for warning-tinted alert boxes and focus states | `class="border border-warning-line"` |

### Danger

| Token | Value | Purpose | Example |
|---|---|---|---|
| `danger` | `#dc2626` | Errors, destructive actions, blocking failures | `class="text-danger"` |
| `danger-hover` | `#b91c1c` | Hover state for danger-colored elements (e.g. a delete button darkening on hover) | `hover:bg-danger-hover` |
| `danger-soft` | `#fef2f2` | Pale danger-tinted background for error banners | `class="bg-danger-soft"` |
| `danger-line` | `#fecaca` | Border color for error banners and danger-focused states | `class="border border-danger-line"` |

---

## 4. Usage Guidelines

**`surface` vs `surface-subtle` vs `surface-strong`**
Think of these as three steps of the same neutral ramp, each with a distinct job:
- `surface` — the default card/panel background. Use this for any raised container.
- `surface-subtle` — a *quieter* fill inside a `surface` container (a nested detail row, a grouped panel) or a neutral hover background on an otherwise plain element.
- `surface-strong` — an *emphasized* fill, reserved for states that need to visually assert themselves against the rest of the page: loading skeletons and selected/active pill backgrounds. Don't reach for this just because you want "a slightly darker gray" — it specifically means "loading" or "currently selected."

**`foreground` vs `foreground-secondary` vs `muted`**
This is a three-level text hierarchy, not three arbitrary shades:
- `foreground` — headings and primary, first-read content.
- `foreground-secondary` — the label color for a neutral, secondary *interactive control* (a bordered button, a pagination link). If it's clickable and not the primary action, this is usually the right choice.
- `muted` — everything else that's secondary: descriptions, captions, timestamps, empty-state copy, placeholder text. If the text is informational rather than actionable, use `muted`.

**`border` vs `border-strong`**
`border` is the default for every resting border in the app. Reach for `border-strong` only when the border needs to visibly react to interaction (hover/focus emphasis) or when a static border needs more visual weight than the default (e.g., a dashed empty-state box, a form checkbox outline). Do not use `border-strong` as a general-purpose "darker border" — it carries the specific connotation of emphasis.

**`brand` vs `info`**
`brand` is Wiyata's identity color — primary actions, active navigation, anything that should read as "this is the app's own color." `info` is reserved for a *future* informational-alert pattern (a neutral, non-branded "here's some information" callout) that doesn't exist in the product yet. Do not use `info` as a substitute for `brand`, and do not use `brand` for a genuinely informational (non-action, non-identity) callout once `info` is adopted — that conflation is exactly what this reservation exists to prevent (see §5).

**`success` vs `warning` vs `danger`**
These map to a standard three-state severity model:
- `success` — the thing worked / is complete / is in good standing.
- `warning` — attention needed, but not blocking (a deadline approaching, a pending review).
- `danger` — a failure, an error, or a destructive/irreversible action.

Do not use `warning` as "a softer danger" or `danger` as "an urgent warning" — pick based on whether the state is blocking (danger) or advisory (warning).

---

## 5. Reserved Tokens

The `info` family (`info`, `info-hover`, `info-soft`, `info-line`) has **zero current usage** in the codebase. This is intentional, not an oversight.

Wiyata does not currently have an informational-alert UI pattern distinct from its brand identity — every "callout" in the product today is either a status alert (success/warning/danger) or a brand-colored highlight. Rather than either (a) omitting an `info` token and forcing a future informational pattern to awkwardly reuse `brand`, or (b) inventing the token's value speculatively without a real use case to validate it against, `info` was deliberately reserved: named, given a value distinct enough from `brand` to be visually unambiguous once adopted, and left unused until a real informational-alert component is built. When that happens, `info` should be adopted as-is rather than re-designed from scratch.

Reserved tokens are not a problem to fix. Do not remove them for being "unused," and do not repurpose them for an unrelated need just because they're sitting idle.

---

## 6. Decorative Colors

The following categories are **intentionally outside the semantic token system** and should never be migrated into it:

- **`subjectPalette`** (`frontend/src/utils/color.ts`) — a fixed, deterministic set of colors used to visually distinguish subjects/classes from one another (similar to how a calendar app assigns colors to different calendars). These colors identify *which subject*, not *what state something is in* — there is no semantic role to tokenize.
- **`.eyebrow` / `.eyebrow-muted` accent family** (`#ea580c` and its tonal relatives) — a deliberate, singular brand accent used for section labels and CTA emphasis across admin/teacher screens. It is a fixed design accent, not a state indicator.
- **Category/role colors** — e.g., color-coding that distinguishes "teacher" vs "student" role badges, or per-metric dashboard stat-card tinting used purely for visual variety across a grid of cards. These colors help a user tell items apart, not tell them what state an item is in.
- **Avatar colors** — initials-based avatar background colors, generated the same way as `subjectPalette`, for the same reason: identity, not state.
- **Illustration/decorative colors** — one-off colors used in empty-state illustrations, chart-adjacent decoration, or similar purely visual flourishes with no semantic meaning.

**Why these stay excluded:** Forcing categorical or decorative color into the semantic system would either (a) require inventing fake "semantic" tokens that don't represent any real UI state, diluting the token vocabulary, or (b) collapse genuinely-different categorical colors onto a small set of state tokens, destroying the very distinctiveness they exist to provide. Semantic tokens answer "what does this mean," and these colors are answering a different question — "which one is this" — that tokens aren't the right tool for.

---

## 7. Future Extensions

The following are potential future semantic families, identified during the v1.0 migration as real gaps but deliberately **not included in v1.0**, since introducing them now would mean guessing at a convention rather than designing one against real product need:

- **`disabled`** — a semantic for disabled interactive-element states (buttons, inputs). Currently expressed ad hoc per component.
- **`divider`** — a distinct token for section-divider/separator lines, currently handled by reusing `border` or left as raw hex in a few places.
- **`overlay`** — a semantic for modal/dialog backdrop scrims.
- **`scrim`** — a semantic for gradient or partial-opacity overlays (e.g., over images).

**These are not part of Wiyata Design Token Architecture v1.0.** Do not add them speculatively. When a real product need for one of these surfaces, it should go through the same evidence-based process used for every token in v1.0: identify the real, repeated usage in source first, then name and value the token from that evidence.

---

## 8. Contributor Rules

✓ Prefer semantic tokens (`text-danger`, `bg-surface-subtle`) over raw hex (`text-[#dc2626]`, `bg-[#fbfaf8]`) in all new code.

✓ Before reaching for a new color, check §3 first — the role you need very likely already has a token.

✓ Do not create a new token for a color that's 5–10% different from an existing one. That's drift, not a new semantic — use the existing token.

✓ Reuse existing semantics whenever the *role* matches, even if you'd have picked a slightly different shade yourself. Consistency beats personal preference here.

✓ Decorative colors (subject/category/avatar/illustration) are allowed to stay as hardcoded values outside this system — see §6. Don't "fix" them by tokenizing them.

✓ A new token requires a new semantic role that no existing token covers — not a new shade of an existing role. If you think you need one, be able to name the *function* it serves that nothing in §3 already serves.

✓ If you're not sure whether something is semantic (tokenize it) or decorative (leave it), ask: "does this color communicate a state/role, or does it just identify/distinguish one item from another?" State → token. Identity → leave it alone.

✗ Do not rename existing tokens without updating this document and auditing every usage site.

✗ Do not repurpose a reserved token (`info`, `success-hover`) for an unrelated need just because it's sitting unused.

---

## 9. Migration Summary

Wiyata's design token migration ran as a series of scoped sprints:

- **Semantic architecture established** — a 29-token system (`background`/`surface` family, `foreground`/`muted` text hierarchy, `border` family, and four-token `brand`/`info`/`success`/`warning`/`danger` status families) was designed from evidence gathered by auditing the app's actual, already-shipped color usage — not from an abstract palette.
- **Hardcoded semantic colors removed** — text hierarchy, structural surfaces and borders, interactive focus/hover states, and status colors (success/warning/danger/brand) were systematically migrated from raw hex and Tailwind-default utilities onto the semantic tokens above, including correcting several instances of pre-existing color drift (the same intended color expressed as multiple slightly-different hex values across components) discovered during the audit.
- **Decorative colors intentionally excluded** — subject/category color-coding, avatar colors, and the `.eyebrow` accent family were identified early and deliberately kept outside the token system throughout the migration, per §6.
- **Design token migration completed** — verified via a final independent audit confirming no broken or misspelled utility classes, no duplicate semantics, a clean production build, and a clean TypeScript type-check. Remaining hardcoded colors in the codebase are all accounted for as either intentionally decorative, intentionally reserved (`info`), or flagged as requiring a future product decision rather than a mechanical migration.

**Wiyata Design Token Architecture v1.0 is frozen as of this document.**
