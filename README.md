# domashka

Russian homework for Spanish speakers. For now it is one page: a table of Russian sounds where every
sound and word has a button with a recording. Go serves the page text and the recordings, React renders
them.

## Running

```sh
make install   # frontend dependencies
make api       # server on :8080
make web       # frontend on :5180, /api is proxied to the server
```

`make serve` builds everything and runs it as one server on :8080. `make help` lists the other targets.

## Configuration

All backend settings live in YAML files in [`config/`](config): addresses, timeouts, directories and
file names, accepted audio formats, request limits, voices and synthesizer and encoder parameters.
`config/base.yaml` holds every default; `config/serve.yaml` and `config/production.yaml` are layered on
top of it and override only what differs:

```sh
domashka -config config/base.yaml -config config/production.yaml
```

The app validates the result at startup and refuses to run with unknown keys or invalid values.

## Code layout

The Go module lives at the repository root: `cmd/domashka` only starts the app, everything else is in
`internal/`, split into `domain`, `application`, `infrastructure` and `transport/httpserver` (Fiber v3)
and wired with Uber fx in `bootstrap`. The frontend is in `web/` (React, Mantine). See
[AGENTS.md](AGENTS.md) for the full map and conventions.

## Production

The app lives at `example.com/domashka/`. The image (`Dockerfile`, ~10 MB) contains the server, the
built frontend and the recordings from `audio/`; there is no admin mode or synthesizer in it, voiceovers
are chosen locally and committed. Server tests run during the build, so a broken `page.json` or
`speech.json` fails the build.

### Deploying to a server with nginx

The server only needs Docker Engine with the compose plugin and ssh access for a user in the `docker`
group. Sources, git, make, Node and Go are not needed there: the image is built on your machine.

1. Once, nginx: paste [`deploy/nginx.conf`](deploy/nginx.conf) inside the site's `server { … }` block
   (where TLS is configured), then `sudo nginx -t && sudo systemctl reload nginx`.
2. Every time, from this machine, from a committed state:

   ```sh
   make deploy SERVER=user@example.com
   ```

   The script ([`deploy/deploy.sh`](deploy/deploy.sh)) builds the image for the server's architecture,
   sends it over ssh, puts `compose.yaml` into `~/domashka/` and starts the container, waiting for
   `healthy`. If port 8080 is taken on the server, use `PORT=8081 make deploy …` and the same port in
   `nginx.conf`. `ALLOW_DIRTY=1` deploys with uncommitted changes.

Rollback: images on the server are tagged by commit:
`docker tag domashka:<commit> domashka:latest && cd ~/domashka && docker compose up -d --no-build`.
After a server reboot the container starts by itself (`restart: unless-stopped`).

Locally the production image runs the same way: `make docker-up` (127.0.0.1:8080/domashka/),
`make docker-down`.

The prefix is `server.base_path` in `config/production.yaml` (`/domashka`; empty means the site root). The UI and
the API use relative URLs, so one build fits any prefix; the proxy forwards the path as is. Caddy works
instead of nginx too:

```caddy
example.com {
	handle /domashka* { reverse_proxy 127.0.0.1:8080 }
}
```

Security (covered by tests in `internal/transport/httpserver/security_test.go`):

- container: non-root, read-only filesystem, no capabilities, memory and process limits, log rotation;
  the port is bound to 127.0.0.1 only (a port published on 0.0.0.0 would bypass the proxy and the host
  firewall), the outside world goes through a TLS proxy;
- server: read, write and idle timeouts; CSP allows scripts only from the app's own origin (the theme script in
  `index.html` by hash, computed at startup; inline styles are allowed for Mantine), `nosniff`, no framing; no directory listings; recordings
  are read once at startup;
- a file path is never built from the URL, only from page keys;
- base images are pinned by digest, updates come from Dependabot.

An app under a prefix shares the origin with the rest of `example.com`, including `localStorage`.
There are no secrets or logins in production, so this is acceptable, but if other people's apps live
on the domain, a subdomain is better.

## Recordings and the admin mode

Recordings live in `audio/` as `.m4a` and are committed together with `audio/choices.json`, which records
the voice and text of each one. Serving them does not need the synthesizer.

Voiceovers are chosen in the UI itself. `make api` starts the server with `-admin`; in the browser run
`localStorage.setItem("is_admin", "true")` once and reload the page. After that, clicking any sound or
word opens a drawer on the right:

- the text for the synthesizer, with `+` before the stressed vowel (`молок+о`); it can be edited and
  generated again;
- candidates from 10 voices: Silero (5 voices, CC BY-NC license: non-commercial use only), Silero CIS
  (4 voices, MIT, but the speakers come from other CIS languages, so an accent is possible) and macOS
  Milena. If there are no candidates yet, they are generated automatically;
- "Choose" makes the candidate the recording on the page; ← → move between clips, 1…0 play a candidate,
  "Next without a choice" jumps to a clip that has not been chosen yet.

Without `-admin` (`make serve`, as for students) nothing can be generated or saved, even with `is_admin`
set. With `-admin` the server accepts requests only from localhost and only from its own page, so another
site open in the same browser cannot touch the admin mode (neither by a forged request nor by DNS
rebinding). Candidates go to `candidates/` (not in git). `make voices` pregenerates candidates for every
clip that has none, so you do not wait in the admin mode.

Silero runs in a separate Python process (`tools/tts.py`) that the server starts through `uv`: pinned
dependencies install themselves, models (≈240 MB) are downloaded on first use into
`~/.cache/domashka/silero` and loaded only if their SHA-256 matches the one in `tools/tts.py` (a model is
a pickle, i.e. executable code). The macOS voice and saving to `.m4a` need macOS (`say`, `afconvert`).

Default texts are in `content/speech.json`. Synthesizers read a lone consonant as the
letter name ("ш" → "ша"), so sounds are written differently there: the continuous ж з с ф х ш щ are
tripled ("шшш"), other consonants are syllables ("б+а"), soft ones are "н+я". The server does not start
if a clip on the page has no text or a word has a typo.

Files can be added by hand too: `.m4a .mp3 .ogg .opus .webm .wav .aac` are accepted, case and stress
marks in the file name do not matter; `make missing` shows what is missing. Without `-admin` the
directory is read once at startup, so restart the server after editing it by hand.

## Page text

All text is in `content/page.json` (read at startup; restart `make api` after editing). Sections consist of `paragraph`, `callout`, `list`, `table` and `alphabet` blocks. Strings
support markup:

- `**bold**`, `*italic*`;
- `{{s:ш}}` is a sound button, `{{w:дом}}` a word button;
- `{{w:молоко|молоко́}}` is a clip key, and after `|` the label to show.

## Languages

The UI is in Spanish, English and Russian. The language comes from the saved choice, otherwise from the
browser language, otherwise Spanish. Every string in `page.json` is either a plain string (the same in
all languages: `"{{w:база}} — {{w:ваза}}"`) or an object `{"es": …, "en": …, "ru": …}`. At startup the
server checks that each such string has all three translations with the same `{{…}}` buttons; otherwise
it refuses to start and says where the error is. UI labels (column headers, hints, the admin mode) are in
`web/src/i18n/locales/`.

## Theme

Light, dark or system: the switch is in the top right corner, the choice is kept in `localStorage`.
The UI uses [Mantine](https://mantine.dev); the theme (colors, fonts, background) is in
`web/src/theme/theme.ts`, component styles are CSS modules next to the components.
