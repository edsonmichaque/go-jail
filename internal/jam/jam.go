package jam

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var fm = template.FuncMap{
	"join": func(s []string) string {
		return strings.Join(s, ", ")
	},
}

type Jail struct {
	JID int
}

type Opts struct {
	Persist   bool       `json:"Persist"`
	Name      string     `json:"Name"`
	Interface string     `json:"Interface"`
	Path      string     `json:"Path"`
	Host      *HostOpts  `json:"Host"`
	IPv4      *IPv4Opts  `json:"IP4"`
	IPv6      *IPv4Opts  `json:"IP6"`
	Exec      *ExecOpts  `json:"Exec"`
	Mount     *MountOpts `json:"Mount"`
	VNet      *VNetOpts  `json:"VNet"`
}

type MountOpts struct {
	DevFS   bool `json:"DevFS"`
	NoDevFS bool `json:"NoDevFS"`
}

type ExecOpts struct {
	PreStart  string
	Start     string
	PostStart string
	PreStop   string
	Stop      string
	PostStop  string
	Clean     bool `json:"Clean"`
}

type VNetOpts struct {
	Interface string
	Enable    bool
}

type HostOpts struct {
	Host     string
	Hostname string
}

type IPOpts struct {
	SAddrSel string   `json:"SAddrSel"`
	Addr     []string `json:"Addr"`
}

type IPv4Opts struct {
	IPOpts
}

type IPv6Opts struct {
	IPOpts
}

type AllowOpts struct{}

func (o Opts) Conf() (io.Reader, error) {
	tmpl := `
# File created by jamd
# DO NOT EDIT

{{ .Name }} {
	{{- if .Mount }}
	{{ if .DevFS}}
    mount.devfs;
	{{ end }}
	{{ if .NoDevFS}}
    mount.nodevfs;
	{{ end }}
	{{- end }}

	{{- if .VNet }}
	{{- if .VNet.Enable }}
	vnet;
	{{ end }}
	{{ end }}

	{{- if .Interface }}
	interface = {{ .Interface }};
	{{ end }}

	{{- if .Host }}
    host.hostname  = {{ .Host.Hostname }};
	{{- end }}
	{{- if .IP4 }}
    ip4.addr       = {{join .IP4.Addr }};
	{{- end }}
    path           = "{{ .Path }}";
	{{- if .Exec }}
	{{- if .Exec.Start }}
    exec.start     = "{{ .Exec.Start }}";
	{{ end }}
	{{- if .Exec.PreStart }}
    exec.prestart  = "{{ .Exec.PreStart }}";
	{{ end }}
	{{- if .Exec.PostStart }}
    exec.poststart = "{{ .Exec.PostStart }}";
	{{ end }}
	{{- if .Exec.Stop }}
    exec.stop      = "{{ .Exec.Stop }}";
	{{ end }}
	{{- if .Exec.PreStop }}
    exec.prestop   = "{{ .Exec.PreStop }}";
	{{ end }}
	{{- if .Exec.PostStop }}
    exec.poststop  = "{{ .Exec.PostStop }}";
	{{- end }}
	{{- if .Exec.Clean }}
    exec.clean;
	{{- end }}
	{{ end }}

	{{- if .Persist }}
	persist;
	{{- end }}
}
`

	return renderTemplate(o, tmpl)
}

func renderTemplate(obj Opts, tmpls string) (io.Reader, error) {
	t, err := template.New("").Funcs(fm).Parse(strings.TrimSpace(tmpls))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, obj); err != nil {
		return nil, err
	}

	return &buf, nil
}

func WriteConf(w io.Writer, r io.Reader) error {
	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	return nil
}

type Wrapper func(io.Reader) (io.Reader, error)

func Create(_ context.Context, parent string, opts *Opts) error {
	pat := filepath.Join(parent, opts.Name+".conf")

	f, err := os.OpenFile(pat, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	defer f.Close()

	conf, err := opts.Conf()
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, conf); err != nil {
		return err
	}

	return nil
}

func TarWrapper(pat string) Wrapper {
	return func(r io.Reader) (io.Reader, error) {
		bf := new(bytes.Buffer)

		t := tar.NewWriter(bf)
		defer t.Close()

		temBuf := new(bytes.Buffer)

		s, err := io.Copy(temBuf, r)
		if err != nil {
			return nil, err
		}

		b, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		h := tar.Header{
			Name: pat,
			Mode: 0o644,
			Size: s,
		}

		t.WriteHeader(&h)
		if _, err := t.Write(b); err != nil {
		}

		if _, err := io.Copy(t, temBuf); err != nil {
			return nil, err
		}

		return bf, nil
	}
}

func GzipWrapper() Wrapper {
	return func(r io.Reader) (io.Reader, error) {
		buf := new(bytes.Buffer)

		g := gzip.NewWriter(buf)
		defer g.Close()

		if _, err := io.Copy(g, r); err != nil {
			return nil, err
		}

		return buf, nil
	}
}

func ZippWrapper(pat string) Wrapper {
	return func(r io.Reader) (io.Reader, error) {
		buf := new(bytes.Buffer)

		z := zip.NewWriter(buf)
		defer z.Close()

		e, err := z.Create(pat)
		if err != nil {
			return nil, err
		}

		if _, err := io.Copy(e, r); err != nil {
			return nil, err
		}

		return buf, nil
	}
}
