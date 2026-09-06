---
title: "SPL Toolkit"
layout: default
---

# SPL Toolkit documentation

SPL Toolkit 0.1.1 provides offline field mapping, seven-category discovery, and validation for the bundled Milestone 1 SPL grammar. Go is the canonical implementation; the CLI, Python package, and REST server use that core.

Start with the [quickstart](quickstart.md), then use the [canonical CLI guide](cli.md), [configuration reference](configuration.md), or [API overview](api/index.md). Installation and source-build instructions are in [installation](installation.md), and measured results are in [performance](performance.md).

The current discovery result has `datamodels`, `datasets`, `lookups`, `macros`, `sources`, `sourcetypes`, and flat `input_fields` arrays. Flat input fields are not lineage. If unsupported macro syntax prevents parsing, macro-only recovery returns the macro names and empty arrays for all other categories.

This release does not claim complete Splunk syntax, semantic completeness, general rewrite safety, schema validation, data-model rewriting, automatic translation, learned mappings, or SPL2 module support.

The roadmap is intentionally narrow: later work can expand the standalone SPL2 parser, introduce structured references and lineage, and add explicit schema-aware analysis. Those are future milestones rather than current APIs.
