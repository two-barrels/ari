<!-- Created by two-barrels in 2026 for ARI v6 modernization. SPDX-License-Identifier: Apache-2.0 -->

# ARI v6.0.0-rc.1 release notes

Module: `github.com/two-barrels/ari/v6`. Published prerelease: `v6.0.0-rc.1`,
at commit `05e769a972ff5b11bb6db96195135cfe52f6db75`. Existing v5 consumers
remain intact; stable v6 and phone-apps adoption remain future work.

## Changes

- Native methods cover all 109 operations and 175 non-path parameters in the
  pinned Asterisk 23 specification. Existing convenience methods remain wrappers.
- Generated models cover 45 event names. Unknown event types and JSON fields are
  preserved, including nested details and keyless events.
- Added explicit collection/path-ID route choices, optional server-assigned IDs,
  new operations, and presence-aware options for explicit false/zero values.
- Stored recording downloads stream through an `io.ReadCloser`.
- Native errors preserve HTTP status and Asterisk error messages.
- Fixed concurrent event-bus shutdown, playback early-stop cancellation, active
  sequence access, and websocket connection-state shutdown races.
- Updated networking and test dependencies. Go minimum is 1.25.0; the recommended
  toolchain is patched Go 1.26.8.

## Migration

Update imports to `github.com/two-barrels/ari/v6` and review
[v6 migration](v6-migration.md), particularly logger interfaces, optional values,
recording stream closure, and application shutdown. Public types make this a
major-version migration. Existing v5 tags remain available; phone-apps has not
been migrated.

## Validation and limits

Hosted Go CI passed on the modernization commit `fe4592b`: Go 1.25/1.26.8
package/example builds and race suites, module verification, vet, and a Go 1.26.8
vulnerability scan. Local standalone snapshot builds and the shared pinned
contract checker passed. These are source snapshots, not verification of future
published tags.

Selected native and proxy operations passed live Asterisk 22.10.1 checks with
cleanup. The user accepts that evidence for current release scope. Additional
live Asterisk 20/23 validation is deferred; all endpoints are not live-certified.
NATS/RabbitMQ tests are deferred for now. The remaining escaping, context,
timeout, and multi-node error audits remain documented in the coordinated
release checklist.

## Publication order

ARI `v6.0.0-rc.1` was published first. Its downloaded packages pass race tests;
proxy now requires that exact tag without a sibling replacement. The matching
proxy candidate follows after its release gates. Both remain prereleases.
