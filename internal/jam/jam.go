package jam

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Jail struct {
	JID int
}

type Options struct {
	Name string      `json:"Name"`
	Host HostOptions `json:"Host"`
	IP4  IPv4Options `json:"IP4"`
	IP6  IPv4Options `json:"IP6"`
	Path string      `json:"Path"`
	Exec interface{}
}

func (j Options) Marshal() (io.Reader, error) {
	tmpl := `
{{ .Name }} {
    mount.devfs;

	{{- if .Host }}
    host.hostname  = {{ .Host.Hostname }};
	{{- end }}
	{{- if .IP4 }}
    ip4.addr       = {{join .IP4.Addr }};
	{{- end }}
    path           = "{{ .Path }}";
	{{- if .Exec }}
    exec.start     = "{{ .Exec.Start }}";
    exec.stop      = "{{ .Exec.Stop }}";
    exec.prestart  = "{{ .Exec.PreStart }}";
    exec.prestop   = "{{ .Exec.PreStop }}";
    exec.poststart = "{{ .Exec.PostStart }}";
    exec.poststop  = "{{ .Exec.PostStop }}";
	{{ end }}
}
`
	tmpl = strings.TrimSpace(tmpl)

	t, err := template.New("").Funcs(template.FuncMap{
		"join": func(s []string) string {
			return strings.Join(s, " ")
		},
	}).Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, j); err != nil {
		return nil, err
	}

	return &buf, nil
}

type Mount struct{}

type HostOptions struct {
	Hostname string
}

type IPOptions struct {
	Addr     []string `json:"Addr"`
	SAddrSel string   `json:"SAddrSel"`
}

type IPv4Options struct {
	IPOptions
}

type IPv6Options struct {
	IPOptions
}

type AllowOptions struct{}

type ExecOptions struct {
	Command   string `json:"Command"`
	Prepared  string `json:"Prepared"`
	PreStart  string `json:"PreStart"`
	Start     string `json:"Start"`
	PostStart string `json:"PostStart"`
	PreStop   string `json:"PreStop"`
	Stop      string `json:"Stop"`
	PostStop  string `json:"PostStop"`
	Clean     bool   `json:"Clean"`
}

func Create(_ context.Context, parent string, opts *Options) error {
	pat := filepath.Join(parent, opts.Name+".conf")

	f, err := os.OpenFile(pat, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	defer f.Close()

	conf, err := opts.Marshal()
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, conf); err != nil {
		return err
	}

	return nil
}
