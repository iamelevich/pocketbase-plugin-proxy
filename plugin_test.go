package pocketbase_plugin_ngrok

import (
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
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
					Url:     "invalid url",
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
