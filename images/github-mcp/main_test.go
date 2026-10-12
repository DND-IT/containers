package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type fakeIssuer struct{ audience []string }

func (f *fakeIssuer) GetWebIdentityToken(_ context.Context, in *sts.GetWebIdentityTokenInput, _ ...func(*sts.Options)) (*sts.GetWebIdentityTokenOutput, error) {
	f.audience = in.Audience
	return &sts.GetWebIdentityTokenOutput{WebIdentityToken: aws.String("jwt")}, nil
}

func TestTokenSourceRenewsAfterRefreshWindow(t *testing.T) {
	now := time.Unix(0, 0)
	minted := 0
	s := &tokenSource{now: func() time.Time { return now }, mint: func(context.Context) (string, error) {
		minted++
		return "t" + string(rune('0'+minted)), nil
	}}
	first, _ := s.Token(context.Background())
	now = now.Add(30 * time.Minute)
	cached, _ := s.Token(context.Background())
	now = now.Add(20 * time.Minute)
	renewed, _ := s.Token(context.Background())
	if first != "t1" || cached != "t1" || renewed != "t2" {
		t.Fatalf("got %s %s %s", first, cached, renewed)
	}
}

func TestTokenSourceKeepsTokenWhileUsableIfRenewalFails(t *testing.T) {
	now := time.Unix(0, 0)
	fail := false
	s := &tokenSource{now: func() time.Time { return now }, mint: func(context.Context) (string, error) {
		if fail {
			return "", errors.New("broker down")
		}
		return "t1", nil
	}}
	_, _ = s.Token(context.Background())
	fail = true
	now = now.Add(50 * time.Minute)
	if token, err := s.Token(context.Background()); err != nil || token != "t1" {
		t.Fatalf("within the usable window: %q %v", token, err)
	}
	now = now.Add(10 * time.Minute)
	if _, err := s.Token(context.Background()); err == nil {
		t.Fatal("expired token was served")
	}
}

func TestBrokerExchangesTheWebIdentityToken(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sts/exchange" || r.URL.Query().Get("scope") != "DND-IT" || r.URL.Query().Get("identity") != "github-mcp-dev" || r.Header.Get("Authorization") != "Bearer jwt" {
			http.Error(w, "bad request "+r.URL.String(), http.StatusForbidden)
			return
		}
		_, _ = w.Write([]byte(`{"token":"ghs_x"}`))
	}))
	defer server.Close()
	domain := strings.TrimPrefix(server.URL, "https://")
	issuer := &fakeIssuer{}
	b := &broker{sts: issuer, http: server.Client(), domain: domain, scope: "DND-IT", identity: "github-mcp-dev"}
	token, err := b.mint(context.Background())
	if err != nil || token != "ghs_x" {
		t.Fatalf("got %q %v", token, err)
	}
	if len(issuer.audience) != 1 || issuer.audience[0] != domain {
		t.Fatalf("audience %v", issuer.audience)
	}
	b.identity = "other"
	if _, err := b.mint(context.Background()); err == nil {
		t.Fatal("a refusal was accepted")
	}
}

func TestProxySetsTokenAndLoopbackHost(t *testing.T) {
	var seen *http.Request
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = r }))
	defer upstream.Close()
	u, _ := url.Parse(upstream.URL)
	handler := proxyTo(u, &tokenSource{now: time.Now, mint: func(context.Context) (string, error) { return "ghs_x", nil }})
	req := httptest.NewRequest(http.MethodPost, "http://bedrock-agentcore.example/mcp", strings.NewReader("{}"))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential=x")
	req.Header.Set("Origin", "https://example.com")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if seen == nil || seen.Header.Get("Authorization") != "Bearer ghs_x" || seen.Host != u.Host || seen.Header.Get("Origin") != "" || seen.URL.Path != "/mcp" {
		t.Fatalf("upstream saw %+v", seen)
	}
}

func TestSplitList(t *testing.T) {
	if got := splitList(" repos, issues ,,pull_requests "); strings.Join(got, "|") != "repos|issues|pull_requests" {
		t.Fatalf("got %v", got)
	}
}
