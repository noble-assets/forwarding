# 05_cli

## Overview

The CLI commands for the `x/forwarding` module allow users to query information and execute transactions.

### Query Commands

#### Query Allowed Denoms

Queries the list of allowed denominations.

```Go
nobled query forwarding denoms
```

#### Query Forwarding Address

Queries the address of a forwarding account based on the specified IBC channel, recipient, and fallback address.

```Go
nobled query forwarding address [channel] [recipient] [fallback]
nobled query forwarding address channel-0 cosmos1... noble1...
```

#### Query Forwarding Stats

Queries general forwarding statistics across all channels.

```Go
nobled query forwarding stats
```

#### Query Forwarding Stats by Channel

Queries general forwarding statistics for a specific IBC channel.

```Go
nobled query forwarding stats [channel]
nobled query forwarding stats channel-0
```

### Transaction Commands

#### Register Forwarding Account

Registers a new forwarding account with the specified recipient address, IBC channel, and fallback address.

```Go
nobled tx forwarding register-account [recipient] [channel] [fallback] --from [signer]
nobled tx forwarding register-account cosmos1... channel-0 noble1... --from mywallet
```

#### Clear Forwarding Account

Clears a forwarding account, sending any remaining packets to the fallback address.

```Go
nobled tx forwarding clear-account [address] [fallback] --from [signer]
nobled tx forwarding clear-account noble1... true --from mywallet
```

#### Set Allowed Denoms

Sets the list of allowed denominations for forwarding within the module.

```bash
nobled tx forwarding set-allowed-denoms [denoms] --from [authority]
nobled tx forwarding set-allowed-denoms uatom uusdc --from noble1...
```