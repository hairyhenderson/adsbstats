# adsbstats

Personal tool for processing [readsb](https://github.com/wiedehopf/readsb) JSON output. Probably not useful to anyone else — it's hardwired to readsb's `aircraft.json` format and built around a specific pipeline (readsb → adsbstats → Grafana Alloy → Loki).

## Commands

### `daemon`

Polls `aircraft.json` from a readsb output directory and emits a JSONL event stream for lifecycle changes:

- `appeared` — new ICAO hex seen
- `position` — lat/lon changed
- `squawk_change` — squawk code changed
- `identified` — callsign first seen
- `disappeared` — ICAO hex no longer in feed

```
adsbstats daemon --input /run/readsb --output ~/adsb-events.jsonl
```

Flags: `--input` (default `/tmp/adsb`), `--output` (default: stdout), `--interval` (default: `1s`)

### `stats`

Reads `history_*.json` files from a readsb output directory and prints a summary table — unique aircraft count, highest altitude, fastest ground speed.

```
adsbstats stats --input /run/readsb
```

## Install

```
go install github.com/hairyhenderson/adsbstats@latest
```

No dependencies outside the standard library.
