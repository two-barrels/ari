# ARI modernization handoff

## Goal and repository relationship

Modernize this Go ARI library and its sibling, `../ari-proxy`, for newer Asterisk REST Interface features. The initial priority was event fidelity through the proxy: ARI applications must receive every event and all event details, including fields unknown to this client's current typed models. The broader goal is support for every endpoint and option in the pinned Asterisk 23 REST specification, with performance and correctness fixes along the way.

Read `../ari-proxy/docs/ari-23-upgrade-plan.md` before extending the work. Its manifest (`ari-23-spec-manifest.json`), endpoint inventory, and parameter coverage CSV are the current contract. They are pinned to Asterisk 23 `rest-api/api-docs` commit `97d55b3306ca4aa26c0136c67b79470ca4b2b785`. Keep the two repositories and the coverage files in sync.

## State as of 2026-09-28

- The module path is `github.com/two-barrels/ari/v6`, but the README identifies v5 as the latest tagged release. Do not treat the current `/v6` path as release readiness.
- Native client methods reach all 109 pinned Asterisk 23 operations, including the five routes previously available only through sibling routes. Explicit methods support Asterisk-assigned IDs and optional IDs on collection routes. This is route availability, not live-server certification.
- The event generator uses the pinned Asterisk 23 `events.json`. It generates 45 typed event names, including seven previously missing types. `event_json.go` preserves unknown event types and JSON fields so a newer event is not silently truncated. Tests cover typed decoding, unknown fields/types, and keyless events.
- Recent option work added Asterisk info `only`; bridge creation variables, playback and recording options; channel creation variables, playback options, external media `transport_data`, hangup `reason_code`, continue `label`, and variable `report_events`. Presence-tracked pointers preserve explicit `false` and zero where needed.
- The native request layer checks HTTP status before body decoding. Stored recording file download streams an `io.ReadCloser`. Additional ARI 23 operations include application event filtering, ping, bridge/channel variables, endpoint REFER, event claim, endpoint text messages, and channel redirect/progress/RTP statistics.
- Implementation sequence #1 in the sibling upgrade plan is complete for the pinned local contract: all 175 parameters have native/proxy wire evidence; response tests cover assigned/missing IDs and representative success/error status handling. `client/native/request.go` includes ARI error `message` text. Shared `testfixtures/asterisk_versions.go` records selected Asterisk 20/22/23 introduction boundaries used by native, proxy, and manifest-checker tests. These simulate server versions; live validation remains next.
- Live validation against a non-disposable Asterisk 22.10.1 test PBX passed native read-only, isolated bridge and external-media channel, websocket event, proxy dispatch/event, and bounded stored-recording checks. The user authorized state-changing tests with cleanup under `ari-testing`; final verification showed zero bridges and channels. See `../ari-proxy/docs/ari-22-live-validation.md`. The live checks exposed a collection bridge-variable encoding mismatch, fabricated optional event payload, and lost HTTP status through `errDataGet`; fixes and regressions are in this repo. Do not store the supplied credential in either repo.
- Modernization is organized into local commits on `codex/v6-modernization`; inspect `git status` before editing and preserve any later work. `origin` is `git@github.com:two-barrels/ari.git`; `upstream` is CyCoreSystems/ari. The branch has not been pushed. Both modules/imports use the final two-barrels/v6 paths. Fork dependency updates were retained and the Go minimum is now 1.25.0. Both race suites passed on Go 1.25.7 and the host compiler, and standalone snapshot builds passed after reconciliation. The sibling proxy uses `replace github.com/two-barrels/ari/v6 => ../ari` during development; approved-tag publication remains a release gate.
- `stdbus` shutdown now synchronizes closed-state and subscription membership, detaches subscriptions before cancellation, and waits for concurrent/repeated `Close` calls with `sync.Once`. Subscribing after closure returns a cancelled subscription with a closed event channel. Concurrent `Cancel` calls wait for channel closure. `stdbus/shutdown_test.go` reproduces the original shutdown races and covers concurrent send/subscribe/cancel/close and post-close subscriptions. Phone-apps teardown must separately stop and join agent-monitor goroutines before closing its bus; that checkout is outside this workspace.

## Next work

Review preparation: modernization is pushed to origin/codex/v6-modernization
at fe4592b; hosted Go CI passed. This supersedes earlier unpushed/hosted-unverified
notes. Draft release notes are in docs/v6-release-notes.md. Review PR submission
is pending GitHub browser sign-in; no merge or release tag is authorized yet.

User scope decision (2026-10-01): existing Asterisk 22.10.1 live evidence is
sufficient for the current release scope. Further live PBX validation, including
Asterisk 20/23, is deferred and is not a current release gate. NATS/RabbitMQ
live testing was also deferred for now. Retain the unverified-version caveats
and local pinned Asterisk 23 contract coverage; do not restart deferred tests.

CI preparation (2026-10-01): `.github/workflows/go.yml` runs module checks,
vet, package/example builds, and race suites on Go 1.25.x and 1.26.8, plus
govulncheck on 1.26.8. GOTOOLCHAIN=local ensures the matrix compiler is used.
Main/master/codex branch pushes, PRs, and manual runs trigger checks. Hosted
runs require pushing this branch; push ARI before proxy because proxy CI checks
out two-barrels/ari at codex/v6-modernization.

Dependency refresh (2026-10-01): networking now uses x/net 0.58.0 and
x/text 0.41.0; testify is 1.12.1. Go minimum remains 1.25.0; the module
recommends patched toolchain go1.26.8. Explicit GOTOOLCHAIN overrides bypass
that recommendation, so use a patched compiler for release builds.
Both repositories passed Go 1.26.8 race suites, Go 1.25.7 compatibility
tests, the 109-operation/175-parameter checker, and standalone snapshot builds.
govulncheck on Go 1.26.8 reported no ARI vulnerabilities and no reachable
proxy vulnerabilities (three module-only proxy advisories remain).

Release preparation drafts are in `docs/v6-migration.md` and the sibling
`docs/v6-release-checklist.md`. The sibling `tools/release-check` packages both
working trees into a temporary module proxy and builds all packages plus an
external consumer without replacements. It also offers published-tag mode;
that gate cannot be completed until approved tags exist. The proxy module and
source/test/example imports now use /v6; its sibling override remains for development.

The 2026-10-01 dev-PBX rerun passed read-only, bridge, event, and full bounded
recording comparison tests under the race detector. It exposed and fixed native
websocket `connected` flag access with `atomic.Bool`; local regression coverage
is `client/native/client_shutdown_test.go`. Cleanup verified zero bridges and
channels. See the sibling live-validation report for remaining limitations.

1. Deferred: live Asterisk 20/23 validation. Asterisk 22.10.1 is accepted for current scope; no live Asterisk 23 certification has been completed. Generated ARI handlers parse JSON-body fields for many POST/PUT operations; keep that encoding where accepted.
2. Extend response/error cases where live tests expose a gap; the pinned local wire contract is complete, but representative responses do not prove every server response shape.
3. Audit remaining path/query escaping and request context propagation. Cover error responses and status codes for each resource; watch media and recording operations that may exceed the current short request timeout.
4. Complete compatibility and migration notes before a coordinated major release. Tag `ari/v6` only after the contract, integration, and standalone-consumer gates pass; then replace the proxy's local module override with a real tag. Keep v5 maintenance separate if required.

## Commands and working conventions

- From this repo: `GOWORK=off GOCACHE=/tmp/ari-go-cache go test ./...` and `git diff --check`.
- From `../ari-proxy`: `GOWORK=off GOCACHE=/tmp/ari-go-cache go test ./...` and `GOWORK=off GOCACHE=/tmp/ari-go-cache go run ./tools/ari-contract`.
- Some native tests use loopback HTTP listeners and may need permission outside a restricted sandbox. The full Go suites and contract checker passed after all 175 parameter evidence rows were filled; rerun after code changes.
- Use the pinned manifest and Go source/tests when changing APIs. Preserve existing convenience methods as wrappers when adding options or route choices. Update public interfaces, native implementation, checked-in mock compatibility methods in `client/arimocks/ari23.go`, proxy request/handler/client, tests, and inventory together.
- Application subscription, global variable writes, and channel dial use JSON bodies; empty global variable values and zero dial timeouts remain explicit. See `client/native/body_write_wire_test.go`.
