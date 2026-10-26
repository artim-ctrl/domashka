# AGENTS.md

Russian homework for Spanish speakers. For now it is one page, a table of Russian sounds where every
sound and word has a button with a recording. More homework sections will follow, so design code and
content for growth.

The user is the author and the teacher.

## Conventions

- Code, identifiers, logs, error messages, docs and commit messages are in English.
- No comments in code. If something needs explaining, rename, extract a function, or put it here.
- Russian, Spanish or any other language appears only as translations (`web/src/i18n/locales/`,
  `content/page.json`) or as the Russian being taught (keys, `content/speech.json`, `audio/`).
- Go: no `init()` functions, no global mutable state, no package-level flags; dependencies are passed
  explicitly. Range over slices and maps by index or key, not by value.
- Backend layers (keep them, but do not split further than needed):
  - `domain/` pure types and rules, no I/O;
  - `application/` use cases (`PageService`, `VoiceoverService`) and the ports they need
    (`RecordingStore`, `CandidateStore`, `Synthesizer`); returns `ErrNotFound`/`ErrInvalid`/`ErrConflict`;
  - `infrastructure/` implements the ports (filesystem, afconvert, Silero/say) plus config and content loading;
  - `transport/httpserver/` is Fiber v3: one package per endpoint under `endpoint/`, each with
    `request.go` (parsing), `handler.go` (calls the application layer) and `response.go` (DTO);
    a file is omitted only when there is nothing to put in it;
  - `bootstrap/` wires everything with Uber fx; `cmd/domashka` only calls `bootstrap.Run`.
  Inner layers never import outer ones.
- No hardcoded settings on the backend: paths, file names, extensions, timeouts, limits, voices and
  encoder parameters live in `config/*.yaml`, are validated by `internal/infrastructure/config` and passed down
  as options. A new setting means a new field in `internal/infrastructure/config`, a value in `config/base.yaml`, a
  validation rule and a test. Only API routes, security headers and the CSP stay in code on purpose.
- Frontend: functions are `const` arrow functions, never `function` declarations; 4-space indentation
  (Biome formats it). UI is built from Mantine components; custom styles go into CSS modules
  (`*.module.css`, processed by `postcss-preset-mantine`) next to the component, using Mantine CSS
  variables. Contexts live in `context.ts` files, providers in their own components; data fetching sits in
  hooks; every user-facing string goes through i18n in all three languages.

## Layout

```
cmd/domashka/          main.go: signal handling and app.Run, nothing else
config/
  base.yaml            every setting with its default (local development)
  serve.yaml           layered on base for `make serve`: also serves web/dist
  production.yaml      layered on base in the Docker image: container paths, /domashka prefix
content/
  page.json            all page text, read at startup
  speech.json          what the synthesizer says for each clip, "+" marks stress
internal/
  domain/
    clip/              clip kind and key, key normalization, kind directories
    content/           page model, Text in three languages, markup tokens, validation
    voiceover/         voices and catalog, choices, candidates, recordings
  application/         PageService, VoiceoverService, ports, errors
  infrastructure/
    config/            YAML config: types, layered loading, validation
    contentfile/       reads content/*.json into the domain model
    recordings/        recordings directory: lookup by key, audio formats, choices file
    candidates/        voiceover candidates on disk
    tts/               Silero worker (tools/tts.py) and macOS say
    audioconv/         afconvert wrappers (WAV, AAC/m4a)
  transport/httpserver/
    server.go          Fiber app, start/stop; routes.go registers endpoints
    endpoint/<name>/   request.go, handler.go, response.go per endpoint
    middleware/        security headers, adminGuard
    static/            built frontend and the CSP
    httperr/ urls/ params/  error mapping, relative URLs, shared request parsing
  bootstrap/           flags, run modes, fx modules (core, voiceover, http)
tools/tts.py           Silero synthesizer process; the server starts it with `uv run --script`
web/                   React 19 + Vite + TypeScript, no router (one page)
  src/app/             App, page loading, page screen, top bar
  src/api/             HTTP client, API types and calls
  src/i18n/            languages, messages type, locales, provider
  src/admin/           voiceover admin
  src/components/      clips, rich text, switches, content blocks (with their CSS modules)
  src/player/          single <audio> player
  src/theme/           Mantine theme, color scheme manager, CSS variables
audio/                 recordings (.m4a) and choices.json, in git
candidates/            voiceover candidates, not in git
```

## Configuration

The binary takes `-config file.yaml`, repeatable: files are applied in order, later ones override only
the keys they set (lists are replaced as a whole). Without `-config` it reads `config/base.yaml`.
Unknown keys and invalid values stop the app at startup. Relative paths are relative to the working
directory (the repository root for `make` targets). Run modes stay flags: `-admin`, `-missing`,
`-voices`, `-voices-all`, `-healthcheck`.

## Commands

Only through the `Makefile` (`make help`):

- `make api` runs the server on :8080 **with the admin mode** (`-admin`); `make web` runs Vite on :5180
  with an `/api` proxy.
- `make check` runs `go vet`, `go test`, the frontend typecheck and Biome (lint + format check).
  Run it after every change; `cd web && npm run format` applies Biome fixes and formatting.
- `make missing` lists missing recordings; `make voices` pregenerates voiceover candidates.
- `make docker-up` / `make docker-down` run the production image through compose locally.
- `make deploy SERVER=user@host` deploys (`deploy/deploy.sh`: built here, image over ssh, only Docker on
  the server). This touches production: only when the user asks.

Port 5173 is taken by another container on this machine; do not move Vite back to it.

The frontend linter and formatter is Biome (`web/biome.json`), not ESLint: typescript-eslint does not
support TypeScript 7 yet.

## Content

- A string in `page.json` is either a plain string (same in all languages) or an object
  `{"es": …, "en": …, "ru": …}`. Write new text in all three languages at once.
- Markup in strings: `**bold**`, `*italic*`, `{{s:ш}}` is a sound button, `{{w:дом}}` a word button,
  `{{w:молоко|молоко́}}` a clip key and its label.
- The server does not start (and `make check` fails) when a translation is missing, a translation has
  different `{{…}}` buttons, a clip on the page has no line in `speech.json`, or a word there has a typo.
- A new word or sound on the page needs a line in `speech.json` with stress (`к+ошка`).

## Voiceover

Recordings are generated by a synthesizer and chosen by a human in the admin mode:
`localStorage.is_admin = "true"` plus the server with `-admin`. Clicking a sound opens a drawer with
candidates from 10 voices; "Choose" writes `audio/<sounds|words>/<key>.m4a` and `audio/choices.json`.

- **Never overwrite `audio/` or `choices.json` yourself**: they are the user's choice by ear.
  Candidates in `candidates/` can be generated freely.
- The user asked to delete the "record with your own voice" page (`/record`); do not suggest it again.
- Synthesizers read a lone consonant as the letter name ("ш" → "ша"). So in `speech.json` the
  continuous ж з с ф х ш щ are tripled (`шшш`) and other consonants are syllables (`б+а`). The user may
  choose otherwise (e.g. `+эс`); that is their decision.
- Silero rejects a string of a single "ё" (it is outside `а-я`): `tools/tts.py` replaces it with "йо".
  Look for similar empty `ValueError` crashes in Silero's text preprocessing.
- Licenses: Silero v5_5_ru is CC BY-NC (non-commercial only); Silero CIS is MIT, but its speakers come
  from other CIS languages. Remind the user about this if money comes up.
- Silero models live in `~/.cache/domashka/silero`; `uv` builds the Python environment itself.
- `say` and `afconvert` exist only on macOS, so the admin mode and generation work only there.

## Production

The `Dockerfile` builds an image from the server, `web/dist` and `audio/` (distroless, non-root,
~10 MB, base images pinned by digest and updated by Dependabot). The admin mode, the synthesizer and
`candidates/` are not in the image and must not be. Server tests run during the build, so broken
content fails the build. The server shuts down gracefully on SIGTERM; `/healthz` and `-healthcheck`
are for Docker. Production is `example.com/domashka/` behind the host's nginx; `deploy/nginx.conf` goes
inside the site's `server { … }` block (the path is forwarded as is; JS/CSS/JSON compression is done
there, Go does not compress).

## Security

Production is open to the internet; the admin mode is local only. Rules that are easy to break:

- The UI lives under a prefix (`server.base_path: /domashka` in `config/production.yaml`). All URLs in
  the frontend and in server responses are **relative** (`api/page`, `api/audio/…`); Vite builds with
  `base: "./"`. An absolute `/api/...` breaks production.
- File paths are built only from page keys (`FindClip`, directory scan), never from the URL.
- New admin endpoints go only into the `opts.Admin` block in `internal/transport/httpserver/routes.go`
  (and `AdminEndpoints`); `adminGuard` protects them (localhost Host, same-origin, `Content-Type: application/json`, which a foreign page cannot send
  without a preflight). The frontend sends JSON; keep it that way.
- CSP (`internal/transport/httpserver/static/static.go`): only `'self'`; the inline theme script in
  `index.html` is allowed by a hash computed at startup. Styles also allow `'unsafe-inline'` because Mantine injects its CSS
  variables in a `<style>` tag and uses inline `style` attributes; scripts must never get it (tested).
  The theme script mirrors Mantine's color scheme logic and the `domashka.theme` key in
  `src/theme/theme.ts`. A new inline script, a third-party font or CDN: first think whether it is
  needed, then update the CSP and its test.
- A new Silero model only with its SHA-256 in `MODELS` (`tools/tts.py`): a model is a pickle, i.e. code.
- `internal/transport/httpserver/security_test.go` holds security regressions; do not weaken them to make
  them pass.

## Verification

- After changes, run `make check`.
- Check the UI in a real browser (headless Chrome + CDP) on a **copy** of `audio/` and a separate port:
  layer a temporary YAML over `config/base.yaml` that overrides `server.addr`, `library.dir` and
  `candidates.dir`. The user may have `make api`/`make web` running from Pitwall; do not
  kill them without asking, and stop them gracefully (SIGINT) if needed.
- Vite listens on `[::1]` only: use `localhost`, not `127.0.0.1`.
- Do not remove Docker containers with data (`compose down`) without asking: neighbouring projects keep
  data in anonymous volumes.
