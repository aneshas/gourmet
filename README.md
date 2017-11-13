<p align="center">
<img src="docs/img/logo.png" alt="Gourmet" title="Gourmet" width="400" />
</p>

# Gourmet
[![wercker status](https://app.wercker.com/status/949708198ad9641d1d0ba724528173f5/s/master "wercker status")](https://app.wercker.com/project/byKey/949708198ad9641d1d0ba724528173f5)
[![codecov](https://codecov.io/gh/tonto/gourmet/branch/master/graph/badge.svg)](https://codecov.io/gh/tonto/gourmet)
[![Go Report Card](https://goreportcard.com/badge/github.com/tonto/gourmet)](https://goreportcard.com/report/github.com/tonto/gourmet)

Gourmet is a light weight load balancer written in Go as a personal experiment. 
It offers a number of features, such as multiple upstream definitions, multiple upstream backends,
configurable upstream server balance algorithms, weight, passive health checks etc...

Primary use for gourmet would be to use it as a reverse proxy to load balance http requests, 
or to use it as a simple api gateway.

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
- [ ] Health fail recover
- [ ] SSL configuration support (add server name to config)
- [ ] Implement least_conn and random 
- [x] Add access log 
- [ ] Kube provider using endpoints (watch?) and test integration using minikube
- [ ] Complete test coverage 
- [ ] End to end integration test with missbehaving servers 
- [ ] Add usage and testing section to readme
- [ ] Add minimal and full config to readme
- [ ] Explain config sections eg. upstream static and kube provider
- [ ] Deploy docker image with wercer
- [X] Move compose to main (pass it as a func to config.Parse?) 

## TODO v0.1.1
- [ ] benchmarks
- [ ] Add description about internals (eg. what headers are set, which are returns, how are errors and timeouts handled etc...)
- [ ] err template file override
- [ ] Add support for promehteus metrics like eg. infulxdb
- [ ] provide lets encrypt as an option for automatic ssl

## TODO v0.2.0 
- [ ] Use raw net and TCP instead of HTTP?
- [ ] Support for multiple servers
- [ ] Add option to pass custom headers
- [ ] Implement different server providers (etcd, consul, ...)
- [ ] Add at least one more protocol 
- [ ] Add cmd option (command) to generate config file
- [ ] Providers and protocols as .so plugins?