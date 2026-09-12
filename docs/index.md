---
title: "SPL Toolkit"
layout: default
---

# SPL Toolkit documentation

SPL Toolkit 0.1.1 provides offline legacy field mapping and discovery, structured analysis and lineage, field-list and JSON Schema/OCSF declaration validation, and explicit safe rewriting. Go is canonical; CLI, Python and REST use that core. Structured operations support bounded SPL and explicitly selected standalone SPL2.

Start with the [quickstart](quickstart.md), then use the [canonical CLI guide](cli.md), [configuration reference](configuration.md), or [API overview](api/index.md). Installation and source-build instructions are in [installation](installation.md), and measured results are in [performance](performance.md).

The current discovery result has `datamodels`, `datasets`, `lookups`, `macros`, `sources`, `sourcetypes`, and flat `input_fields` arrays. Flat input fields are not lineage. If unsupported macro syntax prevents parsing, macro-only recovery returns the macro names and empty arrays for all other categories.

The [rewrite guide](rewrite.md) explains preview/apply, exact audit evidence, linked source bindings, destination validation and held forms. The [API reference](API.md) describes structured analysis and schema validation; the [SPL2 guide](spl2.md) defines its separate supported grammar and effects.

This build does not claim complete Splunk syntax, arbitrary semantic equivalence, event-instance validation, raw-to-data-model translation, learned mappings or SPL2 modules. Read report coverage rather than treating partial results as complete. Local verification is distinct from the historical release matrix and external Splunk execution.
