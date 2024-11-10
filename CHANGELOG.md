# CHANGELOG

## v2.0.0

*Nov 10, 2024*

Second release of the `x/forwarding` module, aligning it with the latest developments in Cosmos and Noble.

### DEPENDENCIES

- Upgrade Cosmos SDK to the latest v0.50.x release. ([#1](https://github.com/noble-assets/forwarding/pull/1))

### FEATURES

- Support [app wiring](https://docs.cosmos.network/v0.50/build/building-apps/app-go-v2) for compatibility with Noble's core codebase. ([#1](https://github.com/noble-assets/forwarding/pull/1))
- Utilize [collections](https://docs.cosmos.network/v0.50/build/packages/collections) for managing module state. ([#1](https://github.com/noble-assets/forwarding/pull/1))
- Support an optional fallback address for easy recovery of funds. ([#12](https://github.com/noble-assets/forwarding/pull/12))
- Allow fine-tuned control over which denoms are forwarded. ([#13](https://github.com/noble-assets/forwarding/pull/13))
- Allow accounts with a balance to be registered signerlessly. ([#18](https://github.com/noble-assets/forwarding/pull/18))

### IMPROVEMENTS

- Update module path for v2 release line. ([#3](https://github.com/noble-assets/forwarding/pull/3))
- Emit events for key module actions (registering, clearing, etc). ([#13](https://github.com/noble-assets/forwarding/pull/13))
- Reorganize repository to align with Noble's standards. ([#19](https://github.com/noble-assets/forwarding/pull/19))

## v1.1.0

*Apr 29, 2024*

This is a minor, non-consensus breaking, release to the `v1` line.

### BUG FIXES

- Register forwarding accounts as a `GenesisAccount` implementation. ([#8](https://github.com/noble-assets/forwarding/pull/8))

### FEATURES

- Implement generic stats query across all channels. ([#5](https://github.com/noble-assets/forwarding/pull/5))

## v1.0.0

*Apr 24, 2024*

This is a standalone release of the `x/forwarding` module introduced to Noble's codebase in the `v4.1.0` Fusion release.

The module has been ported from [noble-assets/noble@`bb9ee09`](https://github.com/noble-assets/noble/commit/bb9ee097f09ce3e923242e9bac3ac997f0d44f8b), which was the latest commit hash that [Halborn](https://www.halborn.com) audited. We have additionally upstreamed [noble-assets/noble#350](https://github.com/noble-assets/noble/pull/350), which introduces a small fix to the forwarding IBC middleware.

