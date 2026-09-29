# datasetlint

> **Moved.** The skill in this repository now lives in [itamaker/skills](https://github.com/itamaker/skills/tree/main/skills/agent-tooling/datasetlint), together with my other skills. Install from there: `npx skills@latest add itamaker/skills --skill=datasetlint`.
>
> This repository still hosts the command-line tool's source and releases.

[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors-)

`datasetlint` is a Go CLI for auditing JSONL datasets before training or evaluation.

It catches duplicates, semantic leakage, label conflicts, and other quality issues that can quietly corrupt benchmark or training results.

![datasetlint social preview](docs/images/social-preview.png)

## Support

[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-FFDD00?style=for-the-badge&logo=buy-me-a-coffee&logoColor=black)](https://buymeacoffee.com/amaker)

## Quickstart

### Install

```bash
brew install itamaker/tap/datasetlint
```

<details>
<summary>You can also download binaries from <a href="https://github.com/itamaker/datasetlint/releases">GitHub Releases</a>.</summary>

Current release archives:

- macOS (Apple Silicon/arm64): `datasetlint_0.2.0_darwin_arm64.tar.gz`
- macOS (Intel/x86_64): `datasetlint_0.2.0_darwin_amd64.tar.gz`
- Linux (arm64): `datasetlint_0.2.0_linux_arm64.tar.gz`
- Linux (x86_64): `datasetlint_0.2.0_linux_amd64.tar.gz`

Each archive contains a single executable: `datasetlint`.

</details>

### First Run

Run:

```bash
datasetlint
```

This launches the interactive Bubble Tea terminal UI.

You can still use the direct command form:

```bash
datasetlint scan -train examples/train.jsonl -eval examples/eval.jsonl
```

## Requirements

- Go `1.22+`

## Run

```bash
go run . scan -train examples/train.jsonl -eval examples/eval.jsonl
```

Tune semantic duplicate detection:

```bash
go run . scan -train examples/train.jsonl -eval examples/eval.jsonl -semantic-threshold 0.6
```

Strict mode exits with a non-zero status when issues are found:

```bash
go run . scan -train examples/train.jsonl -eval examples/eval.jsonl -strict
```

## Build From Source

```bash
make build
```

```bash
go build -o dist/datasetlint .
```

## What It Does

1. Parses train and eval JSONL files.
2. Detects missing IDs and empty input or output fields.
3. Flags duplicate normalized inputs within a split.
4. Detects semantic near-duplicates, label conflicts, and train/eval leakage risks.
5. Summarizes label counts and token-length distributions for quick dataset inspection.

## Notes

- `-json` is useful for CI checks or automated dataset pipelines.
- Maintainer release steps live in `PUBLISHING.md`.

## Claude Code skill

This repo also ships a Claude Code skill. Install standalone:

```bash
npx skills add itamaker/datasetlint-skill
```

Or via the [`itamaker/skills`](https://github.com/itamaker/skills) plugin marketplace:

```text
/plugin marketplace add itamaker/skills
/plugin install datasetlint-skill@itamaker-skills
```

## Contributors ✨

| [![Zhaoyang Jia][avatar-zhaoyang]][author-zhaoyang] |
| --- |
| [Zhaoyang Jia][author-zhaoyang] |



[author-zhaoyang]: https://github.com/itamaker
[avatar-zhaoyang]: https://images.weserv.nl/?url=https://github.com/itamaker.png&h=120&w=120&fit=cover&mask=circle&maxage=7d

## License

[MIT](LICENSE)
