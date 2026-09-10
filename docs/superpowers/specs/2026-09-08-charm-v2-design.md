# Charm v2 migration

> Status (2026-09-10): Deferred with user approval. The application is restored
> to Charm v1 after query-response leakage and stale inline frames were
> reproduced. Retain this document as migration research, not the current
> implementation contract. Future verification must inspect actual terminal
> cleanup across successive stages, not only empty View return values.

## Scope

Upgrade Bubble Tea to v2.0.9, Bubbles to v2.2.1, and Lip Gloss to v2.0.6 using their `charm.land` module paths. Preserve inline terminal layout, keyboard workflows, commit validation, AI/editor transitions, task execution, and the non-TTY push fallback. No Git operations or new product features are part of this migration.

## Decisions

- Return `tea.View` from application models, wrapping existing content with `tea.NewView`. Child Bubbles still render strings. Keep the normal screen buffer.
- Handle `tea.KeyPressMsg`, not the broader key interface, so key releases cannot submit, cancel, or navigate. Let textinput process v2 paste messages.
- Configure textinput focused/blurred styles with `SetStyles`, widths with `SetWidth`, and explicitly retain virtual cursors to preserve the existing composed layout.
- Use Lip Gloss's official `compat.AdaptiveColor` with `lipgloss.Color` values. This retains process-wide local-terminal theme detection without introducing a new theme architecture. Live theme switching and remote Wish sessions remain out of scope.
- Set Bubbles list/textinput default styles using the detected background. Preserve custom styles and pagination layout.
- Keep full-fidelity styled strings inside models. Use output-aware Lip Gloss writers for standalone output and downsample the version template before Cobra prints it. Redirected output must not acquire ANSI escapes.
- Retain the existing rune-width setting unless evidence shows it conflicts with v2; do not broaden migration into Unicode wrapping changes.

## Verification

Add regression tests for v2 key press/release handling, selector selection, input paste/focus/resize, virtual cursor configuration, theme adaptation, and redirected CLI output. Run targeted tests, the full race suite, build, and lint. Smoke-test a read-only command and a harmless model under a PTY. Perform one final adversarial review of the complete diff; verify and fix actionable findings.
