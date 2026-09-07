# FCDS Training Assignment Tool

A tool for automating trainer–trainee assignment and activity
tracking for a competitive programming (CP) community's training
program.

## The problem

The community runs CP training sessions where each trainee is paired
with a trainer who follows up with them individually. Assigning
trainees to trainers was previously done by hand in a spreadsheet,
under two rules:

- every trainer should end up with roughly the same number of trainees
- trainers who only want female trainees should only receive female trainees

On top of that, the community runs a waiting list (there are usually
more applicants than open spots). When an active trainee stops
participating they should be dropped to free a spot for someone from
the waiting list — but doing this by hand means constant manual
spreadsheet edits and repeated messages every time someone is added or
removed, and there was no reliable way to track who was actually
solving problems and who wasn't.

## Status: work in progress

**Implemented**
- Core assignment engine ([`internal/assignment`](internal/assignment)) — computes a
  trainer/trainee assignment that balances load as evenly as possible
  while respecting the gender constraint. Modeled as a max-flow
  problem (see below). Covered by unit tests, including an infeasible
  case where trainees are correctly reported as unassigned instead of
  being force-matched incorrectly.
- CLI demo ([`cmd/assign`](cmd/assign)) that runs the engine against a small
  hardcoded sample so the logic can be exercised end-to-end.

**Planned next**
- Google Sheets integration — read trainer/trainee lists from a live
  sheet and write the resulting assignment back to it, so the sheet
  everyone already looks at stays current automatically.
- Codeforces integration — track submissions to flag inactive
  trainees for a warning/filter step, and identify strong waiting-list
  candidates to promote.
- A notification step for trainer/trainee changes.

## How the assignment engine works

The assignment is modeled as a max-flow problem:

```
source -> trainee (capacity 1) -> eligible trainer (capacity 1) -> sink (capacity = trainer's quota)
```

Each trainer's quota is computed so that quotas differ by at most one
trainee across the whole group. An edge from a trainee to a trainer
only exists if the pairing is allowed — a female trainer's incoming
edges only come from female trainees, while male trainers can accept
either. Running max-flow from source to sink and reading off which
trainee→trainer edges carry flow gives a valid assignment; if the
constraints make a full assignment impossible (for example, more male
trainees than the male trainers' combined quota), the affected
trainees come back in `Assignment.Unassigned` instead of being
silently dropped or matched against the rules.

## Running it

```bash
go run ./cmd/assign
```

## Running tests

```bash
go test ./...
```
## Integrations

- [Codeforces API](https://codeforces.com/apiHelp)
- [Google Sheets API](https://developers.google.com/sheets/api)
- [Telegram Bot API](https://core.telegram.org/bots/api)

