# raceday

NASCAR race tracker for your tmux status bar and terminal.
Shows live leaderboard during races, next race schedule when idle.

## Install

```bash
go install github.com/jfmyers/tmux-raceday/cmd/raceday@latest
```

Or build from source:

```bash
git clone https://github.com/jfmyers/tmux-raceday.git
cd tmux-raceday
go build -o raceday ./cmd/raceday/
```

## tmux Setup

Add to your `.tmux.conf`:

```tmux
# Status bar — shows live race or next scheduled race
set -g status-right '#(raceday --status --driver 9)'
set -g status-interval 10

# Full TUI — press prefix + r for interactive leaderboard
bind r display-popup -E -w 80% -h 80% "raceday --driver 9"
```

Replace `9` with your favorite driver's car number.

## Usage

### Status bar mode

```bash
raceday --status                # next race or live status
raceday --status --driver 24    # include driver position
```

When no race is live:
```
🏁 DAYTONA 500 | Today 1:30 PM | FOX | #24
```

During a live race:
```
🟢 DAYTONA 500 | Lap 142/200 | P1 #8 Busch | #24 Byron P6 [-2]
```

Flag indicators: 🟢 green 🟡 caution 🔴 red 🏁 checkered

### Full TUI mode

```bash
raceday              # launch interactive TUI
raceday --driver 9   # highlight favorite driver
```

Views (press number to switch):

| Key | View | Description |
|-----|------|-------------|
| `1` | Race | Live leaderboard with positions, gaps, speed, pits |
| `2` | Schedule | Race weekend events with countdowns |
| `3` | Entry | Full entry list with teams and manufacturers |
| `4` | Standings | Season points, wins, top-5/10 |

Keyboard shortcuts:

| Key | Action |
|-----|--------|
| `j`/`k` | Scroll up/down |
| `/` | Search for a driver |
| `f` | Jump to favorite driver |
| `tab` | Cycle sort column |
| `q` | Quit |

## Configuration

Generate a config file:

```bash
raceday --init-config
```

This creates `~/.config/raceday/config.yaml`:

```yaml
drivers:
  - 9
  - 24
series: 1           # 1=Cup, 2=Xfinity, 3=Trucks
theme: default
notify:
  cautions: true
  lead_changes: false
  desktop: false
```

The `--driver` flag overrides the config file.

## Data Source

Uses NASCAR's public CDN feeds (`cf.nascar.com`) — the same data
that powers NASCAR.com. No API key required. Updates every few
seconds during live sessions.
