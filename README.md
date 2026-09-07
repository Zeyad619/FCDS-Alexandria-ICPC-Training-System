# FCDS ICPC Training Management System

A Go-based tool for automating trainer–trainee assignment and activity tracking for a competitive programming (CP) community's training program.

## Problem

The community runs CP training sessions where each trainee is followed by a trainer. The previous workflow relied on manual spreadsheet edits, repeated assignment work, and manual checks of trainee activity.

The system is being built around four pieces:

- balanced trainer–trainee assignment with matching constraints
- Google Sheets as the operational data source/output
- Codeforces activity data for trainee monitoring
- Telegram notifications for assignment and status changes

## Current status

### Implemented

- **Assignment engine** (`internal/assignment`) — balances trainer load as evenly as possible while enforcing the gender constraint. The problem is modeled as a **max-flow** network and includes unit tests, including infeasible cases.
- **Codeforces client** (`internal/codeforces`) — fetches public user submissions through the Codeforces API and summarizes recent activity, accepted submissions, unique solved problems, and latest submission time.
- **Google Sheets adapter** (`internal/sheets`) — reads and writes rectangular ranges through the Google Sheets API and converts assignment results into sheet-ready rows.
- **Telegram client** (`internal/telegram`) — sends plain-text notifications through the Telegram Bot API with request/error handling and tests using an HTTP test server.
- Example configuration is provided in `config.example.env`; credentials are intentionally excluded from source control.

### Next

- Connect live Sheets rows to the assignment engine.
- Add persistent trainee/trainer state and assignment history.
- Define warning/filter rules from Codeforces activity.
- Implement waiting-list scoring and promotion.
- Trigger Telegram notifications automatically when assignments or trainee status change.

## Assignment model

```text
source -> trainee (capacity 1)
       -> eligible trainer (capacity 1)
       -> sink (capacity = trainer quota)
```

Trainer quotas differ by at most one trainee. A female trainer only receives female trainees, while male trainers can receive either gender. If a complete assignment is impossible, the engine explicitly returns unassigned trainees rather than violating the constraints.

## Configuration

Copy `config.example.env` and provide local credentials. Never commit Google service-account credentials, Telegram bot tokens, or other secrets.

## Running

```bash
go run ./cmd/assign
go test ./...
```

## Integrations

- [Codeforces API](https://codeforces.com/apiHelp)
- [Google Sheets API](https://developers.google.com/sheets/api)
- [Telegram Bot API](https://core.telegram.org/bots/api)
