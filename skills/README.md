# skills

This directory holds ArtifactA's **assistant publish surfaces** — the skills that let any
skill-capable AI agent (Claude Code and others) publish an artifact to a self-hosted ArtifactA
(`here.now`) instance. Each skill is a thin wrapper over the CLI/REST publish path defined in
[ADR-0012](../docs/adr/0012-assistant-agnostic-publish-surfaces.md); see
[`artifacta-publish/SKILL.md`](./artifacta-publish/SKILL.md) for the publish skill.
