# 04_queries

## Overview

The `x/forwarding` module provides several gRPC and REST query endpoints to retrieve information about allowed denominations, forwarding accounts, and statistics.

### QueryDenoms

`QueryDenoms` retrieves the list of denominations that are currently allowed for forwarding within the module.

#### Request

```Go
{
  "type": "noble/forwarding/v1/QueryDenomsRequest",
  "value": {}
}
```

#### Response

```Go
{
  "type": "noble/forwarding/v1/QueryDenomsResponse",
  "value": {
    "allowed_denoms": [
      "uatom",
      "uusdc"
    ]
  }
}
```

#### Fields

- **allowed_denoms**: a list of denominations that are currently allowed for forwarding

### QueryAddress

`QueryAddress` retrieves the address of a forwarding account based on the specified IBC channel, recipient, and fallback address

#### Request

```Go
{
  "type": "noble/forwarding/v1/QueryAddressRequest",
  "value": {
    "channel": "channel-0",
    "recipient": "cosmos1...",
    "fallback": "cosmos1..."
  }
}
```

#### Response

```Go
{
  "type": "noble/forwarding/v1/QueryAddressResponse",
  "value": {
    "address": "cosmos1...",
    "exists": true
  }
}
```

#### Fields

- **channel**: the IBC channel through which packets are forwarded
- **recipient**: the recipient address
- **fallback**: the fallback address to use if forwarding to the primary recipient fails
- **address**: the forwarding account's address
- **exists**: a boolean indicating whether the forwarding account exists

### QueryStats

`QueryStats` retrieves statistics related to forwarding operations across all channels

#### Request

```Go
{
  "type": "noble/forwarding/v1/QueryStatsRequest",
  "value": {}
}
```

#### Response

```Go
{
  "type": "noble/forwarding/v1/QueryStatsResponse",
  "value": {
    "stats": {
      "channel-0": {
        "num_of_accounts": "1",
        "num_of_forwards": "1",
        "total_forwarded": "1000000uatom"
      },
      "channel-1": {
        "num_of_accounts": "1",
        "num_of_forwards": "1",
        "total_forwarded": "500000uusdc"
      }
    }
  }
}
```

#### Fields

- **stats**: a map containing stats related to the forwarding, delineated by channel

### QueryStatsByChannel

`QueryStatsByChannel` retrieves statistics for a given IBC channel.

#### Request

```Go
{
  "type": "noble/forwarding/v1/QueryStatsByChannelRequest",
  "value": {
    "channel": "channel-0"
  }
}
```

#### Response

```Go
{
  "type": "noble/forwarding/v1/QueryStatsByChannelResponse",
  "value": {
    "num_of_accounts": "10",
    "num_of_forwards": "100",
    "total_forwarded": [
      {
        "denom": "uatom",
        "amount": "1000000"
      }
    ]
  }
}
```

#### Fields

- **channel**: the IBC channel for which statistics are being retrieved
- **num_of_accounts**: the number of registered accounts on the channel
- **num_of_forwards**: the number of forwarded packets on the channel
- **total_forwarded**: the total amount of assets forwarded on the channel, delineated by denomination