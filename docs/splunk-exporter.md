---
title: "Splunk environment exporter"
layout: page
---

# Splunk environment exporter

`spl-toolkit-export` is a separate, bounded acquisition command. It connects to one authenticated Splunk management origin, captures observed inventory and configured knowledge objects, then emits a Snapshot v2 and an optional acquisition report v1. The toolkit's analysis and environment validation remain offline. Capture completeness describes the declared method, visibility and scope; it does not certify query execution or Splunk Cloud support.

Build it with `make build-exporter`, or include it with `make build-all`. The release payload is named `spl-toolkit-export-VERSION-TARGET` (`.exe` on Windows), on the existing Linux amd64, macOS arm64 and Windows amd64 targets. The exporter requires no native library. `--help` and `--version` run without connecting or loading credentials.

```sh
build/spl-toolkit-export --help
build/spl-toolkit-export --version
build/spl-toolkit-export --management-url https://splunk.example.test:8089 --credential-env SPLUNK_TOKEN --ca-file trusted-ca.pem --owner nobody --app search --namespace nobody/search --index main --output snapshot.json --report export-report.json
```

The last command requires an existing credential in the named environment variable and an accessible management origin; the hostname is illustrative. For session authentication, use `--auth-mode session` with an existing Splunk session key. The default `bearer` mode sends an existing Bearer token. Exactly one of `--credential-env NAME` and `--credential-file PATH` is required. A credential file may end with one conventional line terminator. There is no secret-bearing CLI flag and the command does not perform login.

The exporting principal needs access to server information, the relevant configuration endpoints and distributed index inventory, and permission to create/read/delete its discovery jobs and run metadata searches on selected indexes. Existing ACLs, search-peer visibility and search restrictions still apply. Denials, unavailable APIs and incomplete results become acquisition gaps when usable evidence remains; an unusable origin or artifact is fatal. Endpoint availability and permissions must be checked in the intended deployment. No live Splunk Cloud or cross-platform release acceptance is established by this guide.

## Scope and retained evidence

Repeat `--owner`, `--app`, `--namespace` and `--index` to select exact values. Omission means all values visible to the exporting principal. Selectors are deduplicated and sorted. Owner and app values are literal names; `*` and `-` are reserved and cannot be explicit exact selectors. Namespace is the canonical escaped owner/app pair: percent-escape each path component independently, then join with `/`. For owner `analyst/team` and app `search`, use namespace `analyst%2Fteam/search`; pass the literal owner separately with `--owner 'analyst/team'`. Namespace filters returned context, while owner/app select management endpoint contexts. Missing or conflicting ACL context leaves a coverage gap.

The distributed index catalog identifies indexes and event/metric modes. `--index` selects indexes for source and sourcetype observation without trimming the retained confirmed index catalog. Requested but unconfirmed indexes remain explicit unmatched selections. Unreported datatype modes require both event and metric attempts; successful applicable attempts remain separate evidence.

Sources and sourcetypes are observed through fixed metadata searches. Snapshot v2 records `method: splunk_metadata`, `visibility: exporting_principal`, `time_precision: bucket_overlap` and `absence_meaning: not_observed`. A complete capture means the method completed for its selected index/datatype/window; it does not prove event-level exhaustiveness or global absence. Missing observed identities mean **not observed**. Configured collection absence retains its separate scoped semantics. Offline schema bindings to absent observed objects remain unresolved warnings.

Configured adapters acquire macros, saved searches, event types, lookups, data models, tags, calculated fields and field extractions. Field-extraction identities retain `props:` or `transforms:` prefixes to distinguish endpoint families. Dataset, module, function and external-command acquisition are unsupported and explicitly unavailable, so the initial exporter normally produces a partial artifact. The identity-only tag feed may omit owner/app ACL context; that uncertainty remains a gap rather than fabricated scope.

Supported macro and saved-search definitions retain exact source text. A bare macro name is effectively zero-argument even when Splunk advertises a nonempty string `args`; those argument names are ignored. Missing `iseval` defaults to false. Explicit null or invalid `args`/`iseval` remains a gap. Eval macros, validation metadata and unsupported forms do not gain expansion support merely by being captured.

Artifacts can contain sensitive definition text and inventory names. Credentials, origin URLs, job identifiers, response bodies and server messages are omitted from diagnostics; that does not imply universal artifact redaction. Review exact retained definitions before sharing a capture.

## Observation time

With neither bound supplied, `all_retained` dispatches `earliest_time=0` and omits `latest_time`. Optional `--earliest` and `--latest` bounds must be absolute UTC timestamps, for example `2026-10-02T00:00:00.123456789Z`. Up to nine fractional digits are accepted, and zero-offset spellings normalize to `Z`. If both bounds are supplied, earliest must precede latest. Relative time strings are rejected.

Bounded metadata uses bucket overlap, so source/sourcetype identities can reflect events elsewhere in overlapping buckets. The capture does not claim exact event-time filtering. Fractional observation bounds are retained in dispatch and artifacts; whole-second job execution limits are a separate concern.

## Transport, jobs and limits

Verified HTTPS with system trust is the default; `--ca-file` adds PEM trusted certificates. `--allow-insecure` explicitly permits HTTP and skips HTTPS certificate verification. It emits a warning and sets report `allow_insecure: true`; the opt-in alone does not create a collection gap. All redirects are refused, including redirects within the same origin.

Only fixed discovery jobs are dispatched: distributed index enumeration and source/sourcetype metadata for confirmed indexes. There is no arbitrary SPL input, host inventory, raw-event scan or event write. Configuration acquisition performs bounded GETs. The exporter creates random job IDs, tracks ownership, and only cleans up jobs it owns. Cleanup receives an independent 10-second deadline after acquisition cancellation; a cleanup failure is reported and forces exit 3 when the artifact remains usable.

| Option or boundary | Default / limit |
| --- | --- |
| `--request-timeout` | `30s` per request |
| `--job-timeout` | `2m` per discovery job |
| `--overall-timeout` | `10m` per acquisition |
| `--max-rows` | `10000` rows per discovery job |
| Discovery result pages | 500 rows |
| Configuration pages | 100 entries |
| Each HTTP response | 8 MiB |
| Snapshot and report together | 64 MiB including trailing newlines |

Timeouts and row limits must be positive and representable; they are adjustable. Discovery dispatch floors the remaining job/overall duration to whole seconds for Splunk `max_time`; less than one second means no dispatch, avoiding a zero unlimited server timeout. Row caps, warnings, timeouts, conflicting evidence and unavailable work retain partial/unavailable coverage rather than claiming completeness.

The HTTP response limit is independent of the offline REST endpoint's 8 MiB **request body** limit. A valid exporter artifact within the 64 MiB combined output cap can still exceed that REST request limit; CLI, Go or native/Python validation can consume local artifacts without that REST transport boundary.

## Output and offline consumption

Acquisition and owned-job cleanup finish before output. Snapshot/report JSON is validated and staged in each destination directory before rename. Destinations must have existing parents and cannot alias each other, the credential file or the CA file. Separate destination commits are not a transaction: a committed snapshot remains usable if the later report commit fails. Without `--output`, the snapshot is written to stdout; `--report` is optional. Diagnostics go to stderr.

On POSIX, newly staged output files use private `0600` permissions. Windows privacy follows destination ACLs.

| Exit | Meaning |
| --- | --- |
| 0 | Complete usable capture; also successful offline help/version |
| 3 | Usable partial capture or owned-job cleanup failure |
| 2 | Invalid options, unusable origin/artifact, or output failure |

Use the emitted snapshot with `spl-toolkit environment validate --snapshot snapshot.json --format json`, or with `pkg/environment`, native C/Python and the inline REST validation endpoint. Upgrade readers that only accept Snapshot v1 before consuming exporter output: the snapshot version is **2**, while the validation request envelope, schema bundle, validation report and exporter acquisition report retain version **1**. Existing Snapshot v1 inputs remain supported.

The [synthetic observed partial snapshot](https://github.com/delgado-jacob/spl-toolkit/blob/main/examples/environment/observed-partial-snapshot.json) and [matching export report](https://github.com/delgado-jacob/spl-toolkit/blob/main/examples/environment/export-report.json) illustrate the initial gaps; they are not live capture evidence. Register both `contracts/v1` and `contracts/v2` locally for [machine-contract validation](contracts.md). The [environment API](API.md#offline-environment-artifacts), [CLI guide](cli.md#offline-environment-validation) and [compatibility guide](compatibility.md) describe the offline boundaries.
