# Scorecard

A minimal scorecard for mini golf, Scrabble, or anything else you keep score at.
One self-contained HTML file, no build step, no dependencies, no accounts.

Live: https://jeks313.github.io/scorecard/

There are two versions, sharing the same UI:

- **This directory** — the static, github.io-hosted build. Scores live in
  the browser's `localStorage`, so history is per device.
- **`server/`** — a hosted build for k3s (`scores.app.hyde.ca`) with a small
  Go backend storing the same data in SQLite instead, so history follows you
  across devices. See `server/main.go` and `server/build.sh`. The k3s
  manifest lives in `system-config/k3s/scorecard.yaml`.

## What it does

- **Running totals pinned at the top.** The header stays in view while you scroll
  through holes, so the current score is always visible.
- **Remembered players and games.** Tap a chip to pick who's playing; `+ Add`
  adds a new name in place. Same for game names.
- **Saved history.** Every game is stored with its name and date, and listed with
  final scores in finishing order. Tap one to reopen it.
- **Lowest or highest wins**, per game — mini golf and Scrabble rank in opposite
  directions.
- **Works offline** and installs to a phone home screen.

## Notes

This static build stores scores in the browser's `localStorage`, so history is
**per device** — the phone and desktop keep separate records. There is no
backend and nothing is uploaded. (The `server/` build shares one history
across every device instead — see above.)

`Copy` puts a tab-separated grid on the clipboard for Excel or Sheets.

Enter and the arrow keys move down a player's column, so a whole card can be
keyed in without touching the screen.
