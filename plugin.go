package pocketbase_plugin_proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/pocketbase/pocketbase/core"
)

type Skipper func(c *core.RequestEvent) bool

// DefaultSkipper skip proxy middleware for requests, where path starts with /_/ or /api/.
func DefaultSkipper(c *core.RequestEvent) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/_/") || strings.HasPrefix(c.Request.URL.Path, "/api/")
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
	skipper Skipper
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
	plugin.SetSkipper(func(c *core.RequestEvent) bool {
		return c.Request.URL.Path == "/my-super-secret-route"
	})
*/
func (p *Plugin) SetSkipper(skipper Skipper) {
	p.skipper = skipper
}

// singleJoiningSlash joins base path and path, normalizing slashes.
func singleJoiningSlash(base, path string) string {
	baseSlash := strings.HasSuffix(base, "/")
	pathSlash := strings.HasPrefix(path, "/")
	switch {
	case baseSlash && pathSlash:
		return base + path[1:]
	case !baseSlash && !pathSlash:
		return base + "/" + path
	}
	return base + path
}

func (p *Plugin) enableProxy(se *core.ServeEvent) {
	if p.options.Enabled {
		se.Router.BindFunc(func(e *core.RequestEvent) error {
			if p.skipper(e) {
				return e.Next()
			}
			if p.options.ProxyLogsEnabled {
				log.Println("Proxying request from ", e.Request.URL.String(), " to ", p.parsedUrl.String()+e.Request.URL.Path)
			}
			// Build backend URL: base + path (and raw query if any)
			backendURL := *p.parsedUrl
			backendURL.Path = singleJoiningSlash(backendURL.Path, e.Request.URL.Path)
			backendURL.RawQuery = e.Request.URL.RawQuery
			req, err := http.NewRequestWithContext(e.Request.Context(), e.Request.Method, backendURL.String(), e.Request.Body)
			if err != nil {
				return err
			}
			req.Header = e.Request.Header.Clone()
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer func() {
				_ = resp.Body.Close()
			}()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			// Copy status and headers from backend response
			for k, v := range resp.Header {
				for _, vv := range v {
					e.Response.Header().Add(k, vv)
				}
			}
			e.Response.WriteHeader(resp.StatusCode)
			_, _ = e.Response.Write(body)
			return nil
		})

		date := new(strings.Builder)
		log.New(date, "", log.LstdFlags).Print()

		bold := color.New(color.Bold).Add(color.FgGreen)
		_, _ = bold.Printf(
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

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		p.enableProxy(se)
		return se.Next()
	})

	return p, nil
}
