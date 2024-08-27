# 01_state

## Overview

The `x/forwarding` module maintains state related to forwarding accounts, which are specialized accounts used to automatically route tokens from Noble through predefined channels. The state contains account details, channel information, and statistics related to forwarding operations.

### ForwardingAccount

The `ForwardingAccount` structure stores the data needed for forwarding. This includes routing information, account creation details, and a fallback address.

#### Structure

```Go
{
  "BaseAccount": {
    "address": "noble1...",
    "pub_key": null,
    "account_number": "0",
    "sequence": "0"
  },
  "channel": "channel-0",
  "recipient": "cosmos1...",
  "created_at": "1620000000",
  "fallback": "noble1..."
}
```

#### Fields

- **BaseAccount**: inherits from `cosmos.auth.v1beta1.BaseAccount`
- **channel**: specifies the IBC channel through which tokens are forwarded
- **recipient**: the address that receives the forwarded packets
- **created_at**: timestamp at creation
- **fallback**: a fallback address to be used if forwarding to the primary recipient fails

#### State Update

The state is updated by the following messages:
- **`MsgRegisterAccount`**: updates the `ForwardingAccount` state by creating a new account
- **`MsgClearAccount`**: updates the `ForwardingAccount` state by clearing an account

### Genesis State

The genesis state of the `x/forwarding` module sets up the initial configuration, including which denominations are allowed for forwarding and the initial statistics related to registered accounts and forwarding transactions.

#### Structure

```Go
{
  "allowed_denoms": [
    "uatom",
    "uusdc"
  ],
  "num_of_accounts": {
    "channel-0": "1",
    "channel-1": "1"
  },
  "num_of_forwards": {
    "channel-0": "1",
    "channel-1": "1"
  },
  "total_forwarded": {
    "channel-0": "1000000uatom",
    "channel-1": "500000uusdc"
  }
}
```

#### Fields

- **allowed_denoms**: a list of denominations that are allowed to be forwarded
- **num_of_accounts**: a map linking channel IDs to the number of registered forwarding accounts
- **num_of_forwards**: a map linking channel IDs to the number of forwarding transactions
- **total_forwarded**: a map linking channel IDs to the total amount (of denom) forwarded through the channel

### State Update

The state is updated by the following messages:
- **`MsgSetAllowedDenoms`**: updates the `allowed_denoms` field, changing which denominations are permitted for forwarding
