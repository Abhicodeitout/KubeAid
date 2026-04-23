# KubeAid v2 Roadmap Tracker

This tracker is the single source of truth for v2 planning and execution.

## Status Legend

- `planned`: defined, not started
- `in-progress`: actively being built
- `blocked`: waiting on dependency/decision
- `done`: completed and merged

## Release Goal

Target: `v0.0.2` (differentiating release)

Theme: make KubeAid stand out with incident-first debugging workflows.

## Milestones

| Milestone | Scope | Target | Status |
|---|---|---|---|
| M1 | UX + architecture for differentiator features | Week 1 | planned |
| M2 | Implement core feature set and tests | Week 2 | planned |
| M3 | Docs, hardening, and release prep | Week 3 | planned |

## Feature Tracker

| ID | Item | Why it matters | Priority | Owner | Status | Issue |
|---|---|---|---|---|---|---|
| V2-01 | Incident Timeline Mode (`kube-debugger timeline`) | Shows sequence of failure causes instead of raw disconnected events | P0 | TBD | planned | [#7](https://github.com/Abhicodeitout/KubeAid/issues/7) |
| V2-02 | Fix Confidence Score in suggestions | Makes remediation guidance trustworthy and actionable | P0 | TBD | planned | [#8](https://github.com/Abhicodeitout/KubeAid/issues/8) |
| V2-03 | Cross-environment Drift Doctor | Explains why prod fails while staging/dev passes | P1 | TBD | planned | [#9](https://github.com/Abhicodeitout/KubeAid/issues/9) |
| V2-04 | Runbook Auto-Generator (`report --runbook`) | Converts incidents into reusable operational playbooks | P1 | TBD | planned | [#10](https://github.com/Abhicodeitout/KubeAid/issues/10) |
| V2-05 | Release automation for multi-OS assets | Prevents manual release steps and missing artifacts | P1 | TBD | planned | [#11](https://github.com/Abhicodeitout/KubeAid/issues/11) |

## Engineering Checklist

- [ ] Define CLI UX for `timeline` command
- [ ] Define confidence scoring model and output format
- [ ] Add tests for new analyzer paths
- [ ] Add integration test for timeline generation
- [ ] Add/update docs with examples for each new v2 feature
- [ ] Ensure Linux/macOS/Windows artifacts are generated in CI
- [ ] Prepare release notes template for `v0.0.2`

## Decision Log

| Date | Decision | Context |
|---|---|---|
| 2026-04-23 | Focus v2 on differentiation over broad feature count | Keeps release sharp and demo-worthy |

## Weekly Update Template

Copy this section each week:

```md
### Week of YYYY-MM-DD
- Progress:
- Risks:
- Blockers:
- Next:
```
