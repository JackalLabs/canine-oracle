![Jackal Oracle Cover](./assets/jklorc.png)
# Jackal Oracle

[![Build](https://github.com/JackalLabs/canine-oracle/actions/workflows/build.yml/badge.svg)](https://github.com/JackalLabs/canine-oracle/actions/workflows/build.yml)
[![Test](https://github.com/JackalLabs/canine-oracle/actions/workflows/test.yml/badge.svg)](https://github.com/JackalLabs/canine-oracle/actions/workflows/test.yml)
[![golangci-lint](https://github.com/JackalLabs/canine-oracle/actions/workflows/golangci.yml/badge.svg)](https://github.com/JackalLabs/canine-oracle/actions/workflows/golangci.yml)

## Overview
The Jackal Oracle is a server that acts as a middle-man between a Web2 API & the Jackal Blockchain. These servers are equipped with their own keys & will automatically update data feeds.

## Quickstart
This assumes you have either already set up a node or are using another RPC provider in your `~/.jackal-oracle/config/client.toml` file.

You must send tokens to the address that is generated from `gen-key` before starting your node.

```sh
jorcd client config chain-id {current-chain-id}

jorcd client gen-key

jorcd feed create {name}

jorcd feed set-feed {name} {api-link} {interval-seconds}

jorcd start
```

For example, if we wanted an oracle to update the price of Jackal Tokens from Osmosis every 10 seconds, we could do so like this.
```sh
jorcd client config chain-id jackal-1

jorcd client gen-key

jorcd feed create jklprice

jorcd feed set-feed jklprice https://api-osmosis.imperator.co/tokens/v2/price/jkl 10

jorcd start

```

