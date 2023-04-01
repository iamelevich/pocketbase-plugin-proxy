[![Test](https://github.com/iamelevich/pocketbase-plugin-proxy/actions/workflows/test.yml/badge.svg)](https://github.com/iamelevich/pocketbase-plugin-proxy/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/iamelevich/pocketbase-plugin-proxy/branch/master/graph/badge.svg?token=MAXWWCGHWD)](https://codecov.io/gh/iamelevich/pocketbase-plugin-proxy)

<!-- TOC -->
  * [Overview](#overview)
    * [Requirements](#requirements)
    * [Installation](#installation)
    * [Example](#example)
  * [Contributing](#contributing)
<!-- TOC -->

## Overview

This plugin allow proxify requests to other host. It can be useful if you want to use separate server as frontend but use one address for both frontend and backend.

### Requirements

- Go 1.18+
- [Pocketbase](https://github.com/pocketbase/pocketbase) 0.13+

### Installation

```bash
go get github.com/iamelevich/pocketbase-plugin-proxy
```

### Example

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

	// Setup ngrok
	proxyPlugin.MustRegister(app, &proxyPlugin.Options{
		Enabled: true,
		Url:     "http://localhost:3000",
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
```

## Contributing

This pocketbase plugin is free and open source project licensed under the [MIT License](LICENSE.md).
You are free to do whatever you want with it, even offering it as a paid service.
