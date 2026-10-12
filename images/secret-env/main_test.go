package main

import (
	"slices"
	"testing"
)

func TestParseSecret(t *testing.T) {
	values, err := parseSecret(`{"ARGOCD_API_TOKEN":"it's a token","OTHER":""}`)
	if err != nil {
		t.Fatal(err)
	}
	if values["ARGOCD_API_TOKEN"] != "it's a token" || values["OTHER"] != "" || len(values) != 2 {
		t.Fatalf("unexpected values %v", values)
	}
}

func TestParseSecretRejects(t *testing.T) {
	for _, secret := range []string{
		`not json`,
		`null`,
		`["A"]`,
		`{"A":1}`,
		`{"A":{"B":"c"}}`,
		`{"1A":"x"}`,
		`{"A-B":"x"}`,
	} {
		if _, err := parseSecret(secret); err == nil {
			t.Errorf("parseSecret(%s) accepted", secret)
		}
	}
}

func TestRegionOf(t *testing.T) {
	cases := map[string]string{
		"arn:aws:secretsmanager:eu-central-1:000000000000:secret:a/b-AbCdEf": "eu-central-1",
		"ai-platform/agentcore/mcp-servers/argocd":                           "",
	}
	for id, want := range cases {
		if got := regionOf(id); got != want {
			t.Errorf("regionOf(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestMergeReplaces(t *testing.T) {
	got := merge([]string{"PATH=/bin", "TOKEN=old"}, map[string]string{"TOKEN": "new"})
	slices.Sort(got)
	if !slices.Equal(got, []string{"PATH=/bin", "TOKEN=new"}) {
		t.Fatalf("unexpected env %v", got)
	}
}
