# 02_messages

## Overview

The `x/forwarding` module defines several messages concerning management of forwarding accounts and allowed denominations for IBC packet forwarding. These messages allow users to register and clear forwarding accounts, and update the list of denominations that can be forwarded.

### MsgRegisterAccount

When `MsgRegisterAccount` is submitted, it creates a new forwarding account on the specified IBC channel. The message ensures that IBC packets are routed to the `recipient` address, with a fallback option if the primary routing fails. The `signer` is the address of the account who controls the forwarding account. The `fallback` must be a native address.

#### Structure

```Go
{
  "type": "noble/forwarding/MsgRegisterAccount",
  "value": {
    "signer": "cosmos1...",
    "recipient": "cosmos1...",
    "channel": "channel-0",
    "fallback": "cosmos1..."
  }
}
```

#### Fields

- **signer**: the address of the account that is registering the forwarding account
- **recipient**: the address where forwarded packets will be delivered
- **channel**: the IBC channel through which the forwarding occurs
- **fallback**: the fallback address to use if forwarding to the primary recipient fails


### MsgClearAccount

`MsgClearAccount` is used to clear a non-empty forwarding account, returning packets to the `fallback` address. If `fallback` is `false`, packets attempt to send at the end of the next block. The `signer` must have the necessary authority to perform this action.

#### Structure

```Go
{
  "type": "noble/forwarding/MsgClearAccount",
  "value": {
    "signer": "cosmos1...",
    "address": "cosmos1...",
    "fallback": true
  }
}
```

#### Fields

- **signer**: the address of the account that is clearing the forwarding account
- **address**: the address of the forwarding account to be cleared
- **fallback**: a boolean indicating whether to use the fallback address for remaining packets


### MsgSetAllowedDenoms

`MsgSetAllowedDenoms` is used to configure or update the list of token denominations that are allowed for IBC packet forwarding. This is important for maintaining control over which assets are eligible for forwarding, ensuring that only approved tokens are transferred.

#### Structure

```Go
{
  "type": "noble/forwarding/MsgSetAllowedDenoms",
  "value": {
    "signer": "cosmos1...",
    "denoms": [
      "uatom",
      "uusdc"
    ]
  }
}
```

#### Fields

- **signer**: the address authorized to update the list of allowed denominations
- **denoms**: a list of new denominations that are allowed for forwarding
