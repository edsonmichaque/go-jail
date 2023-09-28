package jam

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var fm = template.FuncMap{
	"join": func(s []string) string {
		return strings.Join(s, ", ")
	},
}

type State int

const (
	StateRunning State = iota
)

type Jail struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	State     State
	Config    *CreateOptions
	cmd       *exec.Cmd
	stdin     *bytes.Buffer
	stdout    *bytes.Buffer
	stderr    *bytes.Buffer
}

func (j *Jail) startArgs() []string {
	return []string{
		"-f",
		j.Config.configFilePath(),
		"-c",
		j.Name,
		"-i",
	}
}

func (j *Jail) runCommand(cmd string, args []string) error {
	j.cmd = exec.Command(cmd, args...)

	if err := j.cmd.Run(); err != nil {
		return err
	}

	if !j.cmd.ProcessState.Success() {
		return errors.New("error running")
	}

	return nil
}

func (j *Jail) Start() error {
	if err := j.runCommand("/usr/sbin/jail", j.startArgs()); err != nil {
		return err
	}

	out, err := j.cmd.Output()
	if err != nil {
		return err
	}

	j.ID, err = strconv.ParseInt(string(bytes.TrimSpace(out)), 10, 64)
	if err != nil {
		return err
	}

	return nil
}

func (j Jail) save() ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (j *Jail) stop() error {
	args := []string{
		"-f", j.Config.configFilePath(),
		"-r",
		j.Name,
	}

	cmd := exec.CommandContext(context.Background(), "/usr/sbin/jail", args...)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

type CreateOptions struct {
	Persist   bool          `json:"Persist"`
	Name      string        `json:"Name"`
	Interface string        `json:"Interface"`
	Path      string        `json:"Path"`
	Host      *HostOptions  `json:"Host"`
	IPv4      *IPv4Options  `json:"IP4"`
	IPv6      *IPv4Options  `json:"IP6"`
	Exec      *ExecOptions  `json:"Exec"`
	Mount     *MountOptions `json:"Mount"`
	VNet      *VNetOptions  `json:"VNet"`
	ConfigDir string        `json:"ConfigDir"`
}

func (o CreateOptions) configFilePath() string {
	pat := o.ConfigDir

	if pat == "" {
		pat = "/var/jam/conf"
	}

	return filepath.Join(pat, o.Name+".conf")
}

type MountOptions struct {
	DevFS   bool `json:"DevFS"`
	NoDevFS bool `json:"NoDevFS"`
}

type ExecOptions struct {
	PreStart  string
	Start     string
	PostStart string
	PreStop   string
	Stop      string
	PostStop  string
	Clean     bool `json:"Clean"`
}

type VNetOptions struct {
	Interface string
	Enable    bool
}

type HostOptions struct {
	Host     string
	Hostname string
}

type IPOptions struct {
	SAddrSel string   `json:"SAddrSel"`
	Addr     []string `json:"Addr"`
}

type IPv4Options struct {
	IPOptions
}

type IPv6Options struct {
	IPOptions
}

type AllowOptions struct{}

func (o CreateOptions) buildConfig() (io.Reader, error) {
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

func renderTemplate(obj CreateOptions, tmpls string) (io.Reader, error) {
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

func writeConfig(w io.Writer, r io.Reader) error {
	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	return nil
}

type Wrapper func(io.Reader) (io.Reader, error)

func Create(_ context.Context, parent string, createOpts *CreateOptions) error {
	pat := filepath.Join(parent, createOpts.Name+".conf")

	configFile, err := os.OpenFile(pat, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	defer configFile.Close()

	config, err := createOpts.buildConfig()
	if err != nil {
		return err
	}

	if _, err = io.Copy(configFile, config); err != nil {
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

func ZipWrapper(pat string) Wrapper {
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
