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
        location="api/(.+/?)"
        upstream="backend"

    [[server.locations]]
        location="static/.+/?"
        upstream="front"
```

## TODO v0.1.0
- [x] Recieve on req.Context().Done()
- [x] Passive health checks with max_fail and fail_timeout (per upstream server with defaults if not specified)
- [x] Add html and json error responses based on Accept header
- [x] Add pass regexp path match toggle
- [ ] Add my toy gopher as a mascot
- [ ] Health fail recover
- [ ] SSL configuration support (add server name to config)
- [ ] Add access logs (to stdout, errs to stderr)
- [ ] Kube provider using endpoints (watch?) and test integration using minikube
- [ ] Complete test coverage (without compose - see below)
- [ ] Add usage and testing section to readme
- [ ] Add minimal configuration and full to readme
- [ ] Explain config sections eg. upstream static and kube provider
- [ ] Deploy docker image with wercer
- [ ] Move compose to main (pass it as a func to config.Parse?) 

## TODO v0.2.0
- [ ] Some benchmarks
- [ ] err template file override
- [ ] Add support for promehteus metrics like eg. infulxdb
- [ ] provide lets encrypt as an option for automatic ssl
- [ ] Add description about internals (eg. what headers are set, which are returns, how are errors and timeouts handled etc...)
- [ ] Implement least_conn and random 

## TODO Backlog
- [ ] Use raw net and TCP instead of HTTP?
- [ ] Support for multiple servers
- [ ] Add option to pass custom headers
- [ ] Implement different server providers (etcd, consul, ...)
- [ ] Add at least one more protocol 
- [ ] Add cmd option (command) to generate config file
- [ ] Providers and protocols as .so plugins?