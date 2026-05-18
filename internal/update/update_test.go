package update

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestChocolateyPackageFromPath(t *testing.T) {
	root := filepath.Join("C:", "ProgramData", "chocolatey")

	cases := []struct {
		name string
		exe  string
		want string
	}{
		{
			name: "infracost package",
			exe:  filepath.Join(root, "lib", "infracost", "tools", "infracost.exe"),
			want: "infracost",
		},
		{
			name: "infracost1 package",
			exe:  filepath.Join(root, "lib", "infracost1", "tools", "infracost.exe"),
			want: "infracost1",
		},
		{
			name: "infracost2 package",
			exe:  filepath.Join(root, "lib", "infracost2", "tools", "infracost.exe"),
			want: "infracost2",
		},
		{
			name: "case insensitive",
			exe:  filepath.Join("C:", "PROGRAMDATA", "Chocolatey", "Lib", "Infracost1", "tools", "infracost.exe"),
			want: "infracost1",
		},
		{
			name: "shim path (under bin, not lib) — not detected",
			exe:  filepath.Join(root, "bin", "infracost.exe"),
			want: "",
		},
		{
			name: "binary outside choco — not detected",
			exe:  filepath.Join("C:", "Users", "foo", "Downloads", "infracost.exe"),
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := chocolateyPackageFromPath(tc.exe, root); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUpgradeCommand(t *testing.T) {
	cases := []struct {
		name       string
		method     installMethod
		wantPrefix string
		wantPin    bool
	}{
		{
			name:       "brew",
			method:     installMethod{Kind: installMethodBrew},
			wantPrefix: "$ brew upgrade infracost",
		},
		{
			name:       "choco infracost — includes pin hint",
			method:     installMethod{Kind: installMethodChocolatey, ChocoPkg: "infracost"},
			wantPrefix: "$ choco upgrade infracost",
			wantPin:    true,
		},
		{
			name:       "choco infracost1 — no pin hint",
			method:     installMethod{Kind: installMethodChocolatey, ChocoPkg: "infracost1"},
			wantPrefix: "$ choco upgrade infracost1",
		},
		{
			name:       "choco infracost2 — no pin hint",
			method:     installMethod{Kind: installMethodChocolatey, ChocoPkg: "infracost2"},
			wantPrefix: "$ choco upgrade infracost2",
		},
		{
			name:       "choco missing pkg — defaults to infracost",
			method:     installMethod{Kind: installMethodChocolatey, ChocoPkg: ""},
			wantPrefix: "$ choco upgrade infracost",
			wantPin:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := upgradeCommand(tc.method)
			if !strings.HasPrefix(got, tc.wantPrefix) {
				t.Errorf("got %q, want prefix %q", got, tc.wantPrefix)
			}
			hasPin := strings.Contains(got, "infracost1 package")
			if hasPin != tc.wantPin {
				t.Errorf("pin hint = %v, want %v (full cmd: %q)", hasPin, tc.wantPin, got)
			}
		})
	}
}

func TestParseChocolateyFeed(t *testing.T) {
	v, err := parseChocolateyFeed([]byte(`<?xml version="1.0" encoding="utf-8"?>
<feed>
  <entry>
    <properties>
      <Version>0.10.46</Version>
    </properties>
  </entry>
</feed>`))
	if err != nil {
		t.Fatal(err)
	}
	if v != "v0.10.46" {
		t.Errorf("got %q, want v0.10.46", v)
	}

	empty, err := parseChocolateyFeed([]byte(`<?xml version="1.0"?><feed></feed>`))
	if err != nil {
		t.Fatal(err)
	}
	if empty != "" {
		t.Errorf("expected empty version for feed with no entries, got %q", empty)
	}
}

func TestChocolateyFeedURL(t *testing.T) {
	for _, pkg := range []string{"infracost", "infracost1", "infracost2"} {
		url := chocolateyFeedURL(pkg)
		want := "%27" + pkg + "%27"
		if !strings.Contains(url, want) {
			t.Errorf("for pkg=%q, expected URL to contain %q, got %s", pkg, want, url)
		}
		if !strings.Contains(url, "IsLatestVersion") {
			t.Errorf("for pkg=%q, expected URL to filter IsLatestVersion, got %s", pkg, url)
		}
	}
}
