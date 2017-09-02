<p align="center">
<img src="docs/img/logo.png" alt="Gourmet" title="Gourmet" width="400" />
</p>

# Gourmet
[![wercker status](https://app.wercker.com/status/949708198ad9641d1d0ba724528173f5/s/master "wercker status")](https://app.wercker.com/project/byKey/949708198ad9641d1d0ba724528173f5)
[![Coverage Status](https://coveralls.io/repos/github/tonto/gourmet/badge.svg?branch=)](https://coveralls.io/github/tonto/gourmet?branch=)
[![Go Report Card](https://goreportcard.com/badge/github.com/tonto/gourmet)](https://goreportcard.com/report/github.com/tonto/gourmet)

Gourmet is a light weight load balancer written in Go as a personal experiment.

Primary use would be to load balance http requests, but it should be possible to 
easily extend it to balance any other type of load and/or protocol.

At this moment only round-robin with weight adjustment is available, more coming soon though.
All balancers also implement configurable passive health checks.

## Configuration
Here is an example configuration with all the options that are configurable at this moment:

```toml
[upstreams]
    [upstreams.backend]
    balancer="round_robin" # default round_robin 
    provider="static"      # default static

        [[upstreams.backend.servers]]
        path="api1.foo.bar"
        weight=5 # optional weight

        [[upstreams.backend.servers]]
        path="api2.foo.bar"

        [[upstreams.backend.servers]]
        path="api3.foo.bar"

    [upstreams.front]
    provider="static"
    balancer="round_robin"

        [[upstreams.front.servers]]
        path="static1.foo.bar"
        weight=2    

        [[upstreams.front.servers]]
        path="static2.foo.bar"

        [[upstreams.front.servers]]
        path="static3.foo.bar"

[server]
port=80 # default is 8080
    [[server.locations]]
    location="api/.+[/]"
    upstream="backend"

    [[server.locations]]
    location="static/.+[/]"
    upstream="front"
```

## TODO v0.1.0
- [ ] Complete test coverage
- [x] Passive health checks with max_fail and fail_timeout (per upstream server with defaults if not specified)
- [ ] Health fail recover
- [ ] SSL configuration support
- [ ] Add usage and testing section to readme
- [ ] Add minimal configuration and full to readme
- [ ] Add html and json error responses based on Accept header
- [ ] Add description about internals (eg. what headers are set, which are returns, how are errors and timeouts handled etc...)
- [ ] Implement least_conn and random 

## TODO
- [ ] Implement different server providers (etcd, consul...?)
- [ ] Add at least one more protocol 
- [ ] Provide an example setup using kubernetes
- [ ] Add cmd to generate config file
