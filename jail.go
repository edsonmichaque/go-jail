package jail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type Options struct {
	Name string
}

type Jail struct {
	ID   string      `json:"JID"`
	Name string      `json:"Name"`
	Host HostOptions `json:"Host"`
	IP4  IPv4Options `json:"IP4"`
	IP6  IPv4Options `json:"IP6"`
	Path string      `json:"Path"`
}

func (j Jail) Marshal() (io.Reader, error) {
	tmpl := `
{{ .Name }} {
    mount.devfs;

	{{ if .Host }}
    host.hostname  = {{ .Hostname }};
    ip4.addr       = {{join .IP4Addr }};
	{{ end }}
    path           = "{{ .Path }}";
	{{ if .Exec }}
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

	t, err := template.New("").Parse(tmpl)
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
	Addr     string `json:"Addr"`
	SAddrSel string `json:"SAddrSel"`
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

func Install(_ context.Context, _ Jail) error {
	script := `
PARTITIONS=DEFAULT
DISTRIBUTIONS="kernel.txz base.txz"
export nonInteractive="YES"

#!/bin/sh
sysrc ifconfig_DEFAULT=DHCP
sysrc sshd_enable=YES
pkg install puppet
`
	script = strings.TrimSpace(script)

	fmt.Println(script)

	return nil
}

func Create(_ context.Context, j Jail) error {
	f, err := os.Create(j.Name)
	if err != nil {
		return err
	}

	defer f.Close()

	conf, err := j.Marshal()
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, conf); err != nil {
		return err
	}

	cmd := exec.Command("bsdinstall", "jail", j.Path)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func Start(_ context.Context, o Options) error {
	fmt.Printf("service jail start %s", o.Name)

	return nil
}

func Stop(_ context.Context, o Options) error {
	fmt.Printf("service jail stop %s", o.Name)

	return nil
}

func Update(_ context.Context, o Options) error {
	fmt.Printf("service jail stop %s", o.Name)

	return nil
}

func List(_ context.Context) ([]Jail, error) {
	fmt.Println("jls")

	return nil, errors.New("not implemented")
}
