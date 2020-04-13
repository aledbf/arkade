package get

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"
)

func Download(tool *Tool, os, arch, version string) (string, error) {
	ver := tool.Version
	if len(version) > 0 {
		ver = version
	}
	dlURL, err := tool.GetURL(os, arch, ver)
	if err != nil {
		return "", err
	}

	return dlURL, nil
}

type Tool struct {
	Name           string
	Repo           string
	Owner          string
	Version        string
	URLTemplate    string
	BinaryTemplate string
}

func (t Tool) Latest() bool {
	return len(t.Version) == 0
}

func (tool Tool) GetURL(os, arch, version string) (string, error) {
	if len(tool.URLTemplate) == 0 {
		if len(version) == 0 {
			releases := fmt.Sprintf("https://github.com/%s/%s/releases/latest", tool.Owner, tool.Name)
			var err error
			version, err = findGitHubRelease(releases)
			if err != nil {
				return "", err
			}
		}

		var err error
		t := template.New(tool.Name + "binary")
		var funcs = map[string]interface{}{"HasPrefix": func(s, prefix string) bool { return strings.HasPrefix(s, prefix) }}

		t = t.Funcs(funcs)
		t, err = t.Parse(tool.BinaryTemplate)
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		pref := map[string]string{
			"OS":   os,
			"Arch": arch,
			"Name": tool.Name,
		}
		fmt.Println(pref)
		err = t.Execute(&buf, pref)
		if err != nil {
			return "", err
		}
		res := buf.String()
		fmt.Println(res)
		return fmt.Sprintf(
			"https://github.com/%s/%s/releases/download/%s/%s",
			tool.Owner, tool.Name, version, res), nil
	}

	var err error
	t := template.New(tool.Name)
	t, err = t.Parse(tool.URLTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]string{
		"OS":      os,
		"Arch":    arch,
		"Version": version,
	})
	if err != nil {
		return "", err
	}
	res := buf.String()
	return res, nil
}

func findGitHubRelease(url string) (string, error) {
	timeout := time.Second * 5
	client := makeHTTPClient(&timeout, false)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if res.StatusCode != 302 {
		return "", fmt.Errorf("incorrect status code: %d", res.StatusCode)
	}

	loc := res.Header.Get("Location")
	if len(loc) == 0 {
		return "", fmt.Errorf("unable to determine release of kubeseal")
	}
	version := loc[strings.LastIndex(loc, "/")+1:]
	return version, nil
}

func MakeTools() []Tool {
	tools := []Tool{
		Tool{
			Owner: "openfaas",
			Repo:  "faas-cli",
			Name:  "faas-cli",
			BinaryTemplate: `{{ if HasPrefix .OS "ming" -}}
{{.Name}}.exe
{{- else if eq .OS "darwin" -}}
{{.Name}}-darwin
{{- else if eq .Arch "armv6l" -}}
{{.Name}}-armhf
{{- else if eq .Arch "armv7l" -}}
{{.Name}}-armhf
{{- else if eq .Arch "aarch64" -}}
{{.Name}}-arm64
{{- end -}}`,
		},
		Tool{
			Owner:       "kubernetes",
			Repo:        "kubernetes",
			Name:        "kubectl",
			Version:     "v1.18.0",
			URLTemplate: `https://storage.googleapis.com/kubernetes-release/release/{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl`,
		},
	}
	return tools
}

// makeHTTPClient makes a HTTP client with good defaults for timeouts.
func makeHTTPClient(timeout *time.Duration, tlsInsecure bool) http.Client {
	return makeHTTPClientWithDisableKeepAlives(timeout, tlsInsecure, false)
}

// makeHTTPClientWithDisableKeepAlives makes a HTTP client with good defaults for timeouts.
func makeHTTPClientWithDisableKeepAlives(timeout *time.Duration, tlsInsecure bool, disableKeepAlives bool) http.Client {
	client := http.Client{}

	if timeout != nil || tlsInsecure {
		tr := &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: disableKeepAlives,
		}

		if timeout != nil {
			client.Timeout = *timeout
			tr.DialContext = (&net.Dialer{
				Timeout: *timeout,
			}).DialContext

			tr.IdleConnTimeout = 120 * time.Millisecond
			tr.ExpectContinueTimeout = 1500 * time.Millisecond
		}

		if tlsInsecure {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: tlsInsecure}
		}

		tr.DisableKeepAlives = disableKeepAlives

		client.Transport = tr
	}

	return client
}
