package pocketbase_plugin_proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestPlugin_Validate(t *testing.T) {
	type fields struct {
		app     core.App
		options *Options
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "options is nil",
			fields: fields{
				app:     pocketbase.New(),
				options: nil,
			},
			wantErr: true,
		},
		{
			name: "App is nil",
			fields: fields{
				app:     nil,
				options: &Options{},
			},
			wantErr: true,
		},
		{
			name: "Empty url, but disabled",
			fields: fields{
				app: pocketbase.New(),
				options: &Options{
					Enabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "Enabled, but empty url",
			fields: fields{
				app: pocketbase.New(),
				options: &Options{
					Enabled: true,
					Url:     "",
				},
			},
			wantErr: true,
		},
		{
			name: "Enabled, but invalid url",
			fields: fields{
				app: pocketbase.New(),
				options: &Options{
					Enabled: true,
					Url:     "!@#$%^&*()_+",
				},
			},
			wantErr: true,
		},
		{
			name: "Enabled, but valid url with wrong scheme",
			fields: fields{
				app: pocketbase.New(),
				options: &Options{
					Enabled: true,
					Url:     "redis://localhost:6379",
				},
			},
			wantErr: true,
		},
		{
			name: "Enabled and valid options with http url",
			fields: fields{
				app: pocketbase.New(),
				options: &Options{
					Enabled: true,
					Url:     "http://localhost:300",
				},
			},
			wantErr: false,
		},
		{
			name: "Enabled and valid options with https url",
			fields: fields{
				app: pocketbase.New(),
				options: &Options{
					Enabled: true,
					Url:     "https://localhost:300",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{
				app:     tt.fields.app,
				options: tt.fields.options,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlugin_Register(t *testing.T) {
	_, err := Register(nil, nil)
	if err == nil {
		t.Errorf("Register() should fail when app is nil")
	}
}

func TestPlugin_MustRegister(t *testing.T) {
	// setup the test ApiScenario app instance
	setupTestApp := func(options *Options) func(t testing.TB) *tests.TestApp {
		return func(t testing.TB) *tests.TestApp {
			testApp, err := tests.NewTestApp()
			if err != nil {
				t.Fatal("Cannot initialize test app", err)
			}

			MustRegister(testApp, options)

			return testApp
		}
	}

	proxyDestination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK from " + r.URL.Path))
	}))
	defer proxyDestination.Close()

	proxyURL := proxyDestination.URL

	scenarios := []tests.ApiScenario{
		{
			Name:            "/ request should be proxied when enabled",
			Method:          http.MethodPost,
			URL:             "/",
			ExpectedStatus:  200,
			ExpectedContent: []string{`OK from /`},
			TestAppFactory: setupTestApp(&Options{
				Enabled: true,
				Url:     proxyURL,
			}),
		},
		{
			Name:            "/ request should be proxied when enabled and ProxyLogsEnabled",
			Method:          http.MethodPost,
			URL:             "/",
			ExpectedStatus:  200,
			ExpectedContent: []string{`OK from /`},
			TestAppFactory: setupTestApp(&Options{
				Enabled:          true,
				Url:              proxyURL,
				ProxyLogsEnabled: true,
			}),
		},
		{
			Name:            "/ shouldn be proxied when options nil",
			Method:          http.MethodPost,
			URL:             "/",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory:  setupTestApp(nil),
		},
		{
			Name:            "/ shouldn be proxied when disabled",
			Method:          http.MethodPost,
			URL:             "/",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: setupTestApp(&Options{
				Enabled: false,
				Url:     proxyURL,
			}),
		},
		{
			Name:            "/api/test request should not be proxied when enabled",
			Method:          http.MethodPost,
			URL:             "/api/test",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: setupTestApp(&Options{
				Enabled: true,
				Url:     proxyURL,
			}),
		},
		{
			Name:            "/_/test request should not be proxied when enabled",
			Method:          http.MethodPost,
			URL:             "/_/test",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: setupTestApp(&Options{
				Enabled: true,
				Url:     proxyURL,
			}),
		},
		{
			Name:            "/my-super-api-path request should not be proxied when enabled with custom skipper",
			Method:          http.MethodPost,
			URL:             "/my-super-api-path",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: func(t testing.TB) *tests.TestApp {
				testApp, err := tests.NewTestApp()
				if err != nil {
					t.Fatal("Cannot initialize test app", err)
				}

				p := MustRegister(testApp, &Options{
					Enabled: true,
					Url:     proxyURL,
				})

				p.SetSkipper(func(c *core.RequestEvent) bool {
					return c.Request.URL.Path == "/my-super-api-path"
				})

				return testApp
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

// receivedRequest captures what the proxy destination server receives.
type receivedRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func TestPlugin_ProxifiesRequests(t *testing.T) {
	// Proxy destination that echoes received request data as JSON
	proxyDestination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()

		headers := make(map[string]string)
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		received := receivedRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Query:   r.URL.RawQuery,
			Headers: headers,
			Body:    string(body),
		}
		jsonBytes, _ := json.Marshal(received)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonBytes)
	}))
	defer proxyDestination.Close()

	setupTestApp := func(options *Options) func(t testing.TB) *tests.TestApp {
		return func(t testing.TB) *tests.TestApp {
			testApp, err := tests.NewTestApp()
			if err != nil {
				t.Fatal("Cannot initialize test app", err)
			}
			MustRegister(testApp, options)
			return testApp
		}
	}

	t.Run("forwards HTTP method", func(t *testing.T) {
		for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
			method := method
			t.Run(method, func(t *testing.T) {
				scenario := tests.ApiScenario{
					Name:            method + " request",
					Method:          method,
					URL:             "/",
					ExpectedStatus:  200,
					ExpectedContent: []string{`"method":"` + method + `"`},
					TestAppFactory:  setupTestApp(&Options{Enabled: true, Url: proxyDestination.URL}),
				}
				scenario.Test(t)
			})
		}
	})

	t.Run("forwards request path", func(t *testing.T) {
		scenarios := []struct {
			path string
		}{
			{"/"},
			{"/some/path"},
			{"/nested/deep/path"},
			{"/trailing/slash/"},
		}
		for _, tc := range scenarios {
			tc := tc
			t.Run("path"+tc.path, func(t *testing.T) {
				scenario := tests.ApiScenario{
					Name:            "path " + tc.path,
					Method:          http.MethodGet,
					URL:             tc.path,
					ExpectedStatus:  200,
					ExpectedContent: []string{`"path":"` + tc.path + `"`},
					TestAppFactory:  setupTestApp(&Options{Enabled: true, Url: proxyDestination.URL}),
				}
				scenario.Test(t)
			})
		}
	})

	t.Run("forwards query parameters", func(t *testing.T) {
		scenario := tests.ApiScenario{
			Name:   "query params",
			Method: http.MethodGet,
			URL:    "/search?q=hello&page=2&filter=active",
			// JSON escapes & as \u0026, so we verify each param is present
			ExpectedStatus:  200,
			ExpectedContent: []string{`q=hello`, `page=2`, `filter=active`},
			TestAppFactory:  setupTestApp(&Options{Enabled: true, Url: proxyDestination.URL}),
		}
		scenario.Test(t)
	})

	t.Run("forwards request headers", func(t *testing.T) {
		scenario := tests.ApiScenario{
			Name:   "custom headers",
			Method: http.MethodGet,
			URL:    "/",
			Headers: map[string]string{
				"X-Custom-Header": "custom-value",
				"X-Request-Id":    "test-123",
				"Accept":          "application/json",
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"X-Custom-Header":"custom-value"`, `"X-Request-Id":"test-123"`, `"Accept":"application/json"`},
			TestAppFactory:  setupTestApp(&Options{Enabled: true, Url: proxyDestination.URL}),
		}
		scenario.Test(t)
	})

	t.Run("forwards request body", func(t *testing.T) {
		body := `{"payload":"proxified-body-xyz789"}`
		scenario := tests.ApiScenario{
			Name:            "request body",
			Method:          http.MethodPost,
			URL:             "/",
			Body:            strings.NewReader(body),
			Headers:         map[string]string{"Content-Type": "application/json"},
			ExpectedStatus:  200,
			ExpectedContent: []string{`proxified-body-xyz789`},
			TestAppFactory:  setupTestApp(&Options{Enabled: true, Url: proxyDestination.URL}),
		}
		scenario.Test(t)
	})
}
