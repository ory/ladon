<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [go-hostpool](#go-hostpool)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

go-hostpool
===========

A Go package to intelligently and flexibly pool among multiple hosts from your Go application.
Host selection can operate in round robin or epsilon greedy mode, and unresponsive hosts are
avoided.
Usage example:

```go
hp := hostpool.NewEpsilonGreedy([]string{"a", "b"}, 0, &hostpool.LinearEpsilonValueCalculator{})
hostResponse := hp.Get()
hostname := hostResponse.Host()
err := _ // (make a request with hostname)
hostResponse.Mark(err)
```

View more detailed documentation on [godoc.org](http://godoc.org/github.com/bitly/go-hostpool)
