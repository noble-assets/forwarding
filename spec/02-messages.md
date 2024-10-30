# 02_messages

## Overview

The `x/forwarding` module defines several messages concerning management of forwarding accounts and allowed denominations for automatic forwarding. These messages allow users to register and clear forwarding accounts, and an authority to update the list of denominations that can be forwarded.

### MsgRegisterAccount

When `MsgRegisterAccount` is submitted, it creates a new forwarding account for a specified IBC channel. The message ensures that received tokens are automatically routed to the `recipient` address, with a fallback option if the primary routing fails. The `signer` is the native address of the account registering the forwarding account. The `fallback` must be a native address.
#### Structure

```Go
{
  "type": "noble/forwarding/RegisterAccount",
  "value": {
    "signer": "noble1...",
    "recipient": "cosmos1...",
    "channel": "channel-0",
    "fallback": "noble1..."
  }
}
```

#### Fields

- **signer**: the address of the account that is registering the forwarding account
- **recipient**: the address where forwarded tokens will be sent
- **channel**: the IBC channel through which the forwarding occurs
- **fallback**: the fallback address to use if forwarding to the primary recipient fails


### MsgClearAccount

`MsgClearAccount` is used to clear a non-empty forwarding account, returning tokens to the `fallback` address. If `fallback` is `false`, tokens attempt to send at the end of the current block.

#### Structure

```Go
{
  "type": "noble/forwarding/ClearAccount",
  "value": {
    "signer": "noble1...",
    "address": "noble1...",
    "fallback": true
  }
}
```

#### Fields

- **signer**: the address of the account that is clearing the forwarding account
- **address**: the address of the forwarding account to be cleared
- **fallback**: a boolean indicating whether to use the fallback address for receiving tokens


### MsgSetAllowedDenoms

`MsgSetAllowedDenoms` is used to configure or update the list of token denominations that are allowed for automatic forwarding. This is important for maintaining control over which assets are eligible for forwarding, ensuring that only approved tokens are routed.
#### Structure

```Go
{
  "type": "noble/forwarding/SetAllowedDenoms",
  "value": {
    "signer": "noble1...",
    "denoms": [
      "ausdy",
      "uusdc"
    ]
  }
}
```

#### Fields

- **signer**: the address authorized to update the list of allowed denominations
- **denoms**: a list of new denominations that are allowed for forwarding
