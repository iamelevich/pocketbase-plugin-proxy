package pocketbase_plugin_ngrok

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

func (p *Plugin) enableProxy(e *core.ServeEvent) error {
	if p.options.Enabled {
		// Skip PocketBase routes
		skipperFunc := func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().URL.Path, "/_/") || strings.HasPrefix(c.Request().URL.Path, "/api/")
		}

		if p.options.ProxyLogsEnabled {
			e.Router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
				Skipper: skipperFunc,
			}))
		} else {
			log.Println("Proxy logs are disabled")
		}
		e.Router.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
			Skipper: skipperFunc,
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
	return nil
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
	p := &Plugin{app: app}

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
		if err := p.enableProxy(e); err != nil {
			return err
		}
		return nil
	})

	return p, nil
}
