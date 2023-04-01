package pocketbase_plugin_ngrok

import (
	"net/http"
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

func TestPlugin_MustRegister(t *testing.T) {
	// setup the test ApiScenario app instance
	setupTestApp := func(options *Options) func() (*tests.TestApp, error) {
		return func() (*tests.TestApp, error) {
			testApp, err := tests.NewTestApp()
			if err != nil {
				return nil, err
			}

			MustRegister(testApp, options)

			return testApp, nil
		}
	}

	proxyDestinationServer := &http.Server{
		Addr: "localhost:1234",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK from " + r.URL.Path))
		}),
	}
	defer proxyDestinationServer.Close()
	go proxyDestinationServer.ListenAndServe()

	scenarios := []tests.ApiScenario{
		{
			Name:            "/ request should be proxied when enabled",
			Method:          http.MethodPost,
			Url:             "/",
			ExpectedStatus:  200,
			ExpectedContent: []string{`OK from /`},
			TestAppFactory: setupTestApp(&Options{
				Enabled: true,
				Url:     "http://localhost:1234",
			}),
		},
		{
			Name:            "/ shouldn be proxied when options nil",
			Method:          http.MethodPost,
			Url:             "/",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory:  setupTestApp(nil),
		},
		{
			Name:            "/ shouldn be proxied when disabled",
			Method:          http.MethodPost,
			Url:             "/",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: setupTestApp(&Options{
				Enabled: false,
				Url:     "http://localhost:1234",
			}),
		},
		{
			Name:            "/api/test request should not be proxied when enabled",
			Method:          http.MethodPost,
			Url:             "/api/test",
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: setupTestApp(&Options{
				Enabled: true,
				Url:     "http://localhost:1234",
			}),
		},
		{
			Name:            "/_/test request should not be proxied when enabled",
			Method:          http.MethodPost,
			Url:             "/_/test",
			ExpectedStatus:  405,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory: setupTestApp(&Options{
				Enabled: true,
				Url:     "http://localhost:1234",
			}),
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}
