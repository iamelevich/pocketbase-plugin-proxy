package pocketbase_plugin_proxy

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase/core"
)

// DefaultSkipper skip proxy middleware for requests, where path starts with /_/ or /api/.
func DefaultSkipper(c echo.Context) bool {
	return strings.HasPrefix(c.Request().URL.Path, "/_/") || strings.HasPrefix(c.Request().URL.Path, "/api/")
}

// Options defines optional struct to customize the default plugin behavior.
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

type Plugin struct {
	// app is a Pocketbase application instance.
	app core.App

	// options is a plugin options.
	options *Options

	// parsedUrl from options.Url
	parsedUrl *url.URL

	// Skipper function for proxy middleware
	skipper middleware.Skipper
}

// Validate plugin options. Return error if some option is invalid.
func (p *Plugin) Validate() error {
	if p.options == nil {
		return fmt.Errorf("options is required")
	}

	if p.app == nil {
		return fmt.Errorf("app is required")
	}

	if p.options.Enabled {
		if p.options.Url == "" {
			return fmt.Errorf("url is required when proxy is enabled")
		}

		// Check is url valid
		if parsedUrl, err := url.Parse(p.options.Url); err != nil {
			return fmt.Errorf("url is invalid")
		} else {
			if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
				return fmt.Errorf("url schema is invalid, only http and https are supported")
			}
			// Fill plugin parsedUrl
			p.parsedUrl = parsedUrl
		}
	}

	return nil
}

/*
SetSkipper set skipper function that should return true if that route shouldn't be proxied.

If not set, the DefaultSkipper is used:

If set - you should also control the middleware behavior for /_/ and /api/ routes.

Example:

	plugin := proxyPlugin.MustRegister(app, &proxyPlugin.Options{
		Enabled: true,
		Url:     "http://localhost:3000",
	})
	plugin.SetSkipper(func(c echo.Context) bool {
		return c.Request().URL.Path == "/my-super-secret-route"
	})
*/
func (p *Plugin) SetSkipper(skipper middleware.Skipper) {
	p.skipper = skipper
}

func (p *Plugin) enableProxy(e *core.ServeEvent) {
	if p.options.Enabled {
		if p.options.ProxyLogsEnabled {
			e.Router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
				Skipper: p.skipper,
			}))
		} else {
			log.Println("Proxy logs are disabled")
		}
		e.Router.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
			Skipper: p.skipper,
			Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
				{
					URL: p.parsedUrl,
				},
			}),
		}))

		date := new(strings.Builder)
		log.New(date, "", log.LstdFlags).Print()

		bold := color.New(color.Bold).Add(color.FgGreen)
		bold.Printf(
			"%s Proxy will forward requests to %s\n",
			strings.TrimSpace(date.String()),
			color.CyanString("%s", p.parsedUrl.String()),
		)
	}
}

// MustRegister is a helper function that registers plugin and panics if error occurred.
func MustRegister(app core.App, options *Options) *Plugin {
	if p, err := Register(app, options); err != nil {
		panic(err)
	} else {
		return p
	}
}

// Register registers plugin.
func Register(app core.App, options *Options) (*Plugin, error) {
	p := &Plugin{
		app:     app,
		skipper: DefaultSkipper,
	}

	// Set default options
	if options != nil {
		p.options = options
	} else {
		p.options = &Options{}
	}

	// Validate options
	if err := p.Validate(); err != nil {
		return p, err
	}

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		p.enableProxy(e)
		return nil
	})

	return p, nil
}
