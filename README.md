# Samulator

A tool for building OpenBazaar simluations.

# Usage

## Installation

`go install -i github.com/OpenBazaar/samulator/cmd/samulator`

## Options

```
$ samulator -h
Usage:
  samulator [OPTIONS]

Application Options:
  -b, --buyer=  path to buyer configuration
  -v, --vendor= path to vendor configuration
  -m, --mod=    path to mod configuration

Help Options:
  -h, --help    Show this help message
```

## Notes

- `samulator` currently only runs v0.13.2 of ob-go
- Recommended: You should create your configuration directories ahead of time as `samulator` will simply `ob-go start -d <config-path>` after building (the default behavior of this is to initialize a new data directory at that location).
- It is recommended that the JSON API listen ports are adjusted for the three nodes to not conflict with each other. For example, change the `Gateway` and `Swarm` addresses to listen on ports which aren't used by other nodes or processes as shown below.

```json
"Addresses": {
  "API": null,
  "Announce": null,
  "Gateway": "/ip4/0.0.0.0/tcp/4002", # change port 4002
  "NoAnnounce": null,
  "Swarm": [
     "/ip4/0.0.0.0/tcp/4001",    # change port 4001
     "/ip6/::/tcp/4001",         # change port 4001
     "/ip4/0.0.0.0/tcp/9005/ws", # change port 9005
     "/ip6/::/tcp/9005/ws"       # change port 9005
  ]
},
```

# Design/Architecture

The builder was designed with the expectation that there are no prepared binaries or supporting infrastructure available to execute a test. As such, the builder is self-sufficient in producing and executing binaries. This should be kept in mind while improving this codebase as external dependencies may make building the tool easier, they also make using the tool harder. Given that we will likely use this more than build/improve it, we should optimize accordingly.

## Builder

Produces builds of specific applications for use in a simulation. A successful `Build()` should yield a runner for that application.

Builders have a regular interface and should not change their interface from application to application.

Builders rely on Blueprints to inflate the source in preparation for a `Build()`.

### Disk Use

The builder uses `$HOME/.samulator` for workspace while building and caching binaries for use. This path is expendable and is recreated on each run (at the cost of rebuilding any needed components).

If `$HOME` is not defined, it may be provided or the current working directory will be used instead (ex: `./.samulator`).

## Blueprints

Inflates sourcecode for a specific application and is capable of manipulating the source in preparation for building.

## Cacher

Stores a copy of produced binaries for later use as the `Build()` process tends to be expensive.

The cacher uses `$HOME/.samulator/cache` to store binaries.

# Contributions/Improvements

Contributions are gladly accepted. Planned improvements include:

- [ ] Runner can change Address/Swarm ports of config
- [ ] Runners can be added to a NetworkSandbox to deterministically isolate/control communications
- [ ] Builder can create other specializations of ob-go (such as pushnode or gateway configurations)
- [ ] Runners can manipulate the node's JSON API to complete the QA tests
