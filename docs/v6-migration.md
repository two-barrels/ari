# Migrating ARI applications to v6

Status: release preparation, 2026-10-01. This checkout uses
`github.com/two-barrels/ari/v6`; that path does not establish release readiness.
Use a reviewed development revision until the coordinated release gates pass.
The README identifies v5 as the latest tagged release.

## Toolchain and imports

The module requires Go 1.25.0. Generated JSON
uses `omitzero` for absent optional event objects, so the old README's Go 1.11
requirement no longer applies.

Change all imports from `github.com/CyCoreSystems/ari/v5` to
`github.com/two-barrels/ari/v6`, including `client/native`, `client/arimocks`,
`stdbus`, `rid`, and extensions. Types from different major paths are distinct;
update the application and its adapters together. Applications using forks must
also replace the fork's module prefix and review any fork-specific APIs.

Once an actual release tag exists, require its exact version with
`go get github.com/two-barrels/ari/v6@<approved-tag>`.
Do not use this placeholder as a command. Proxy consumers should follow the
[sibling proxy migration guide](../../ari-proxy/docs/v6-migration.md) and use
matching v6 modules.

## Interfaces and options

The public resource interfaces have additional methods. Existing convenience
methods remain available as wrappers, but custom interface implementations and
old generated mocks need the new methods. Regenerate private mocks and compile
all adapters. The checked-in `client/arimocks/ari23.go` supplies compatibility
methods for the repository mocks.

New operations include event filtering, ping, bridge/channel variables,
endpoint REFER, event claim, channel redirect/progress/RTP statistics, and
stored-recording file streaming. Explicit collection methods such as
`Bridge.CreateWithoutID`, `Bridge.CreateOnCollection`, `Channel.PlayOnCollection`,
and `Channel.SnoopOnCollection` distinguish optional/requested IDs from
Asterisk-assigned IDs. Read the returned handle key rather than assuming a
locally supplied ID. A missing response ID is an error.

Use `WithOptions` methods to supply additional playback, recording, variable,
hangup, continuation, and creation options. Pointer options distinguish an
omitted value from explicit zero or false:

```go
reportEvents := false
opts := &ari.ChannelVariableSetOptions{ReportEvents: &reportEvents}
err := c.Channel().SetVariableWithOptions(key, "MY_VAR", "", opts)
```

POST/PUT requests retain JSON bodies where Asterisk accepts them. Collection
bridge creation with variables uses the flat body shape verified against
Asterisk 22.10.1; do not substitute the path-ID route's nested variables shape.

## Events and application ownership

The pinned Asterisk 23 schema generates 45 typed event names. `DecodeEvent`
preserves unknown JSON fields and represents future event types as
`*ari.UnknownEvent`. Update event switches to handle unknown types and use
`GetType()` where a typed assertion is unnecessary. Re-marshalling decoded
events retains unknown fields; typed modifications take precedence. Unknown
data preservation covers JSON event paths, not a guarantee for protobuf-only
transports or application code that constructs a new reduced event object.

`ari.CloneEvent` copies the top-level event and routing metadata; nested payload
values remain shared. Treat received payloads as immutable or make your own deep
copy before changing nested data. Absent optional event objects remain absent
instead of being serialized as fabricated empty resources.

## Errors, downloads, and shutdown

HTTP errors are checked before response decoding. Native errors include ARI's
message where supplied, and wrapped resource errors preserve status through
`native.CodeFromError`. Use status codes rather than exact error strings.

`StoredRecording.File(ctx, key)` returns `*ari.RecordingFile`. Always close
`file.Body`; copy it with an `io.Reader` operation rather than assuming a JSON
response. Native HTTP clients can be configured through `native.Options.HTTPClient`.
Check operation timeouts for media/download workloads; the short default is
not appropriate for every operation.

`stdbus.Close` is safe with concurrent subscribe/send/cancel calls and repeated
closure. A subscription attempted after closure has a closed event channel.
Consumers must check the channel receive's `ok` value and exit on closure.
Buffered events can be drained after cancellation. Application teardown should
cancel and join its own goroutines before releasing their shared dependencies.
The playback extension now preserves an early stop request and synchronizes its
active sequence pointer. Native websocket connected-state access is also
synchronized. These fixes do not certify every application lifecycle.

## Compatibility evidence and release checks

All 109 pinned routes and 175 non-path parameters have local native/proxy wire
coverage. Live Asterisk 22.10.1 checks cover selected reads, isolated bridges and
channels, event JSON forwarding, and bounded recording streaming. This does not
certify every operation or all Asterisk 14+ versions. New endpoints on older
servers may fail with version-specific errors; consult the pinned manifest and
`testfixtures/asterisk_versions.go` for selected introduction boundaries.

The user accepted Asterisk 22.10.1 as sufficient live validation for current
release scope on 2026-10-01. Live Asterisk 20/23 checks are deferred, and
NATS/RabbitMQ checks are deferred for now. Follow the
[coordinated release checklist](../../ari-proxy/docs/v6-release-checklist.md).
From the sibling proxy checkout, `go run ./tools/release-check` packages both
working trees and builds an external consumer without workspace or local
replacement dependencies. Published-tag verification is a separate gate.
