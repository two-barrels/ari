# ARI v6 release notes — draft

Module: `github.com/two-barrels/ari/v6`. No v6 tag has been published.

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

Review and freeze the ARI commit, choose an approved v6 tag, publish ARI first,
and verify that exact tag from an external consumer. Then update the proxy to
require it and remove the sibling replacement. No draft note authorizes merging,
tagging, or artifact publication.
