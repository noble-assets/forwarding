# 03_events

## Overview

The `x/forwarding` module emits events for actions such as registration or clearing of forwarding accounts and updates to the list of allowed denominations.

### AccountRegistered

`AccountRegistered` is emitted when a new forwarding account is registered.

#### Structure

```Go
{
  "type": "noble/forwarding/v1/AccountRegistered",
  "attributes": {
    "address": "cosmos1...",
    "channel": "channel-0",
    "recipient": "cosmos1...",
    "fallback": "cosmos1..."
  }
}
```

#### Fields

- **address**: the address of the newly registered forwarding account
- **channel**: the IBC channel used for forwarding
- **recipient**: the recipient address
- **fallback**: the fallback address to use if the primary forwarding fails

#### Emitted By

- **Transaction**: `noble.forwarding.v1.MsgRegisterAccount`

### AccountCleared

`AccountCleared` is emitted when a forwarding account is cleared.

#### Structure

```Go
{
  "type": "noble/forwarding/v1/AccountCleared",
  "attributes": {
    "address": "cosmos1...",
    "recipient": "cosmos1..."
  }
}
```

#### Fields

- **address**: the address of the cleared forwarding account
- **recipient**: the recipient address if the fallback is used

#### Emitted By

- **Transaction**: `noble.forwarding.v1.MsgClearAccount`

### AllowedDenomsConfigured

`AllowedDenomsConfigured` is emitted whenever the list of allowed denominations is updated. 

#### Structure

```Go
{
  "type": "noble/forwarding/v1/AllowedDenomsConfigured",
  "attributes": {
    "previous_denoms": ["uatom", "uusdc"],
    "current_denoms": ["uatom", "uusdc", "uiris"]
  }
}
```

#### Fields

- **previous_denoms**: the list of denominations allowed before the update
- **current_denoms**: the newly configured list of allowed denominations

#### Emitted By

- **Transaction**: `noble.forwarding.v1.MsgSetAllowedDenoms`