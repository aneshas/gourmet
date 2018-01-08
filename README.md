<p align="center">
<img src="docs/img/logo.png" alt="Gourmet" title="Gourmet" width="400" />
</p>

# Gourmet
[![wercker status](https://app.wercker.com/status/949708198ad9641d1d0ba724528173f5/s/master "wercker status")](https://app.wercker.com/project/byKey/949708198ad9641d1d0ba724528173f5)
[![codecov](https://codecov.io/gh/tonto/gourmet/branch/master/graph/badge.svg)](https://codecov.io/gh/tonto/gourmet)
[![Go Report Card](https://goreportcard.com/badge/github.com/tonto/gourmet)](https://goreportcard.com/report/github.com/tonto/gourmet)

Gourmet is a light weight L7 proxy written in Go as a personal experiment. 
It offers a number of features, such as service discovery (static/dynamic), load balancing, 
passive health checks, TLS termination etc...

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
- [X] Move compose to main (pass it as a func to config.Parse?) 
- [x] Add access log 
- [x] Health fail recover
- [x] Implement random
- [ ] Complete test coverage 
- [ ] Add section to readme
- [ ] Add minimal and full config to readme (add test for minimal config)
- [ ] Add queue size to upstream server toml config
- [ ] Kube provider using endpoints (watch?) and test integration using minikube
- [ ] Implement least_conn  
- [ ] Explain config sections eg. upstream static and kube provider
- [ ] Deploy docker image with wercker
- [ ] End to end integration test (minikube / sidecar proxy) 
- [ ] SSL configuration support (add server name to config)

## v0.1.1 ideas
- [ ] benchmarks
- [ ] err template file override
- [ ] Add observability support (tracing, configurable logging, prometheus stats)
- [ ] provide lets encrypt as an option for automatic ssl?

## v0.2.0 ideas
- [ ] Use raw net and TCP instead of HTTP
- [ ] Support for multiple servers
- [ ] Add option to pass custom headers
- [ ] Implement different upstream providers/backends (service discovery options) eg. consul...
- [ ] Add more protocols
- [ ] Providers and protocols as .so plugins?