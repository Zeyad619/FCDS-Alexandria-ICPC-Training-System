# FCDS Alexandria ICPC Training System

A tool that automates trainer-trainee assignment, activity tracking,
and waiting-list management for a competitive programming (CP)
community's training program.

## The problem

Community training pairs each trainee with a trainer who follows up
with them individually, under two rules: every trainer gets roughly
the same number of trainees, and trainers who only want female
trainees should only receive female trainees. The community also runs
a waiting list (there are usually more applicants than open spots),
promoting people in when an active trainee stops participating. Doing
all of this by hand means constant manual spreadsheet edits and
repeated messages every time someone is added or removed, with no
reliable way to track who was actually solving problems.

## Status

**Implemented and tested**
- **Assignment engine** ([`internal/assignment`](internal/assignment)) — a
  two-phase, two-pointer algorithm: partition trainers and trainees by
  gender (O(N)), fill each female trainer's fair share from the female
  trainee pool, then fill the male trainers from everyone left over
  (leftover female trainees plus every male trainee), with the male
  trainers' quotas recomputed against exactly what's left so unused
  female-trainer capacity is never wasted. Covered by unit tests,
  including the case a naive single-pass version gets wrong (see
  "Design notes" below) and the case where someone genuinely can't be
  placed.
- **Codeforces client** ([`internal/Codeforces`](internal/Codeforces)) — fetches a
  handle's submissions from the public Codeforces API and summarizes
  them (total/accepted/unique-solved/solved-in-period). Tested.
- **Telegram client** ([`internal/telegram`](internal/telegram)) — sends a message
  to a chat via the Telegram Bot API. Tested against a local mock
  server.
- **Roster** ([`internal/roster`](internal/roster)) — parses trainer/trainee rows
  from a sheet (header-driven, order-independent columns) into typed
  data, and converts state back into sheet rows. Tested, including bad
  input.
- **Policy** ([`internal/policy`](internal/policy)) — pure decision logic: given a
  trainee's roster status and Codeforces activity, decides
  warn/filter/promote/none. No I/O, fully unit-tested across every
  branch.
- **Orchestrator** ([`internal/orchestrator`](internal/orchestrator)) — wires all of
  the above into one pipeline run (read roster → check activity →
  apply warnings/filtering/promotion → recompute assignment → write
  back → notify only what changed). Depends on small local interfaces
  rather than the concrete Sheets/Codeforces/Telegram clients, so the
  full pipeline is tested end-to-end with fakes — no live credentials
  needed to verify the logic.
- `cmd/assign` — CLI demo of the assignment engine alone, on sample data.

**Written, not yet verified end-to-end**
- `cmd/sync` wires the *real* Sheets/Codeforces/Telegram clients into
  the orchestrator. `internal/sheets` needs
  `google.golang.org/api`, which isn't in `go.mod` yet — run
  `go get google.golang.org/api/sheets/v4 && go mod tidy` before
  building it. Beyond that, this hasn't been run against a real
  spreadsheet, real Codeforces handles, or a real bot token yet, so
  treat it as needing a supervised first run, not a proven deployment.

**Planned next**
- Run `cmd/sync` against a real (test) spreadsheet and a real bot to
  validate the whole loop, then schedule it (a time-driven trigger or
  a scheduled job) instead of running it by hand.
- Configurable policy thresholds (currently hardcoded to
  `policy.DefaultConfig()` in `cmd/sync`).

## Design notes: why two-pointer instead of a general matching algorithm

The gender constraint here isn't an arbitrary bipartite eligibility
graph — it has a specific shape: male trainers accept anyone, female
trainers accept only female trainees. That "one group takes anyone"
structure is what makes a two-phase greedy approach correct without
needing a general matching/flow algorithm: giving female trainees to
female trainers first never costs the male trainers a placement they'd
otherwise have made, since male trainers could have taken those
trainees anyway. A *naive* single-pass version (sort everyone by
gender, then hand out `quota` trainees per trainer straight down both
lists) doesn't handle this correctly — if a female trainer's quota
exceeds the number of female trainees available, that kind of
single-pass walk keeps taking from the trainee list and hands her a
male trainee next. The two-phase split (fill female trainers from the
female pool only, then recompute the male trainers' quotas over
whatever's actually left) avoids that and also happens to place
*more* trainees than a fixed-quota version would in a shortage, since
capacity a female trainer couldn't use is picked up by the male
trainers instead of sitting idle.

## Running it

```bash
go run ./cmd/assign      # assignment engine demo on sample data, no setup needed
go run ./cmd/sync        # full pipeline against real Sheets/Codeforces/Telegram — needs config.example.env values set as real environment variables, and the go.mod fix above
```

## Running tests

```bash
go test ./...
```
