[![Test](https://github.com/iamelevich/pocketbase-plugin-proxy/actions/workflows/test.yml/badge.svg)](https://github.com/iamelevich/pocketbase-plugin-proxy/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/iamelevich/pocketbase-plugin-proxy/branch/master/graph/badge.svg?token=MAXWWCGHWD)](https://codecov.io/gh/iamelevich/pocketbase-plugin-proxy)

<!-- TOC -->
* [Overview](#overview)
  * [Requirements](#requirements)
  * [Installation](#installation)
  * [Example](#example)
* [pocketbase\_plugin\_proxy](#pocketbasepluginproxy)
  * [Index](#index)
  * [func DefaultSkipper](#func-defaultskipper)
  * [type Options](#type-options)
  * [type Plugin](#type-plugin)
    * [func MustRegister](#func-mustregister)
    * [func Register](#func-register)
    * [func \(\*Plugin\) SetSkipper](#func-plugin-setskipper)
    * [func \(\*Plugin\) Validate](#func-plugin-validate)
* [Contributing](#contributing)
<!-- TOC -->

# Overview

This plugin allow proxify requests to other host. It can be useful if you want to use separate server as frontend but use one address for both frontend and backend.

## Requirements

- Go 1.18+
- [Pocketbase](https://github.com/pocketbase/pocketbase) 0.13+

## Installation

```bash
go get github.com/iamelevich/pocketbase-plugin-proxy
```

## Example

You can check examples in [examples folder](/examples)

```go
package main

import (
	"log"

	proxyPlugin "github.com/iamelevich/pocketbase-plugin-proxy"
	"github.com/pocketbase/pocketbase"
)

func main() {
	app := pocketbase.New()

	// Setup proxy plugin
	proxyPlugin.MustRegister(app, &proxyPlugin.Options{
		Enabled: true,
		Url:     "http://localhost:3000",
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
```

<!-- gomarkdoc:embed:start -->

<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# pocketbase\_plugin\_proxy

```go
import "github.com/iamelevich/pocketbase-plugin-proxy"
```

## Index

- [func DefaultSkipper(c echo.Context) bool](<#func-defaultskipper>)
- [type Options](<#type-options>)
- [type Plugin](<#type-plugin>)
  - [func MustRegister(app core.App, options *Options) *Plugin](<#func-mustregister>)
  - [func Register(app core.App, options *Options) (*Plugin, error)](<#func-register>)
  - [func (p *Plugin) SetSkipper(skipper middleware.Skipper)](<#func-plugin-setskipper>)
  - [func (p *Plugin) Validate() error](<#func-plugin-validate>)


## func [DefaultSkipper](<https://github.com/iamelevich/pocketbase-plugin-proxy/blob/master/plugin.go#L18>)

```go
func DefaultSkipper(c echo.Context) bool
```

DefaultSkipper skip proxy middleware for requests, where path starts with /\_/ or /api/.

## type [Options](<https://github.com/iamelevich/pocketbase-plugin-proxy/blob/master/plugin.go#L23-L34>)

Options defines optional struct to customize the default plugin behavior.

```go
type Options struct {
    // Enabled defines if proxy should be enabled.
    Enabled bool

    //Url to the target.
    //
    //Only http and https links are supported.
    Url string

    // Are proxy logs enabled?
    ProxyLogsEnabled bool
}
```

## type [Plugin](<https://github.com/iamelevich/pocketbase-plugin-proxy/blob/master/plugin.go#L36-L48>)

```go
type Plugin struct {
    // contains filtered or unexported fields
}
```

### func [MustRegister](<https://github.com/iamelevich/pocketbase-plugin-proxy/blob/master/plugin.go#L132>)

```go
func MustRegister(app core.App, options *Options) *Plugin
```

MustRegister is a helper function that registers plugin and panics if error occurred.

### func [Register](<https://github.com/iamelevich/pocketbase-plugin-proxy/blob/master/plugin.go#L141>)

```go
func Register(app core.App, options *Options) (*Plugin, error)
```

Register registers plugin.

### func \(\*Plugin\) [SetSkipper](<https://github.com/iamelevich/pocketbase-plugin-proxy/blob/master/plugin.go#L97>)

```go
func (p *Plugin) SetSkipper(skipper middleware.Skipper)
```

SetSkipper set skipper function that should return true if that route shouldn't be proxied.

If not set, the DefaultSkipper is used:

If set \- you should also control the middleware behavior for /\_/ and /api/ routes.

Example:

```
plugin := proxyPlugin.MustRegister(app, &proxyPlugin.Options{
	Enabled: true,
	Url:     "http://localhost:3000",
})
plugin.SetSkipper(func(c echo.Context) bool {
	return c.Request().URL.Path == "/my-super-secret-route"
})
```

### func \(\*Plugin\) [Validate](<https://github.com/iamelevich/pocketbase-plugin-proxy/blob/master/plugin.go#L51>)

```go
func (p *Plugin) Validate() error
```

Validate plugin options. Return error if some option is invalid.



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)


<!-- gomarkdoc:embed:end -->

# Contributing

This pocketbase plugin is free and open source project licensed under the [MIT License](LICENSE.md).
You are free to do whatever you want with it, even offering it as a paid service.

