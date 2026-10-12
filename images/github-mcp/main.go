// Command github-mcp serves GitHub's MCP server read-only over streamable HTTP, authenticated
// with a GitHub App installation token from an octo-sts token broker.
//
// The upstream HTTP server takes its GitHub token from each request's Authorization header,
// which a platform such as Amazon Bedrock AgentCore Runtime uses for its own signature. So the
// upstream server listens on loopback, and this process fronts it: it proves its AWS identity
// to the broker with an STS web identity token, exchanges that for an installation token, and
// sets it on every request it forwards, buying a new one before the hour is up.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ghhttp "github.com/github/github-mcp-server/v2/pkg/http"
)

const (
	upstreamPort      = 8082
	webIdentitySecs   = 300
	tokenRefreshAfter = 45 * time.Minute
	tokenUsableFor    = 55 * time.Minute
)

func main() {
	if err := run(); err != nil {
		slog.Error("github-mcp stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	tokens, err := newTokenSource()
	if err != nil {
		return err
	}
	if _, err := tokens.Token(context.Background()); err != nil {
		return fmt.Errorf("first token: %w", err)
	}

	upstreamErr := make(chan error, 1)
	go func() {
		upstreamErr <- ghhttp.RunHTTPServer(ghhttp.ServerConfig{
			Version:           "github-mcp",
			Port:              upstreamPort,
			ListenHost:        "127.0.0.1",
			ContentWindowSize: 5000,
			ReadOnly:          true,
			EnabledToolsets:   splitList(os.Getenv("GITHUB_TOOLSETS")),
		})
	}()
	if err := waitForUpstream(upstreamErr); err != nil {
		return err
	}

	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8000"
	}
	server := &http.Server{
		Addr:              net.JoinHostPort(os.Getenv("MCP_HOST"), port),
		Handler:           proxy(tokens),
		ReadHeaderTimeout: 60 * time.Second,
	}
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.ListenAndServe() }()
	slog.Info("github-mcp listening", "addr", server.Addr)

	select {
	case err := <-upstreamErr:
		if err == nil {
			return nil
		}
		return fmt.Errorf("upstream server: %w", err)
	case err := <-serverErr:
		return err
	}
}

func proxy(tokens *tokenSource) http.Handler {
	return proxyTo(&url.URL{Scheme: "http", Host: net.JoinHostPort("127.0.0.1", strconv.Itoa(upstreamPort))}, tokens)
}

func proxyTo(upstream *url.URL, tokens *tokenSource) http.Handler {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(upstream)
			r.Out.Host = upstream.Host
			r.Out.Header.Del("Authorization")
			r.Out.Header.Del("Origin")
			if token, err := tokens.Token(r.In.Context()); err == nil {
				r.Out.Header.Set("Authorization", "Bearer "+token)
			} else {
				slog.Error("no GitHub token for the request", "error", err)
			}
		},
		FlushInterval: -1,
	}
}

func waitForUpstream(upstreamErr <-chan error) error {
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(upstreamPort))
	for range 100 {
		select {
		case err := <-upstreamErr:
			return fmt.Errorf("upstream server: %w", err)
		default:
		}
		if conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond); err == nil {
			return conn.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("upstream server did not start listening")
}

type tokenSource struct {
	mint func(context.Context) (string, error)

	mu       sync.Mutex
	token    string
	mintedAt time.Time
	now      func() time.Time
}

func newTokenSource() (*tokenSource, error) {
	if token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN"); token != "" {
		return &tokenSource{mint: func(context.Context) (string, error) { return token, nil }, now: time.Now}, nil
	}
	domain, scope, identity := os.Getenv("TOKEN_BROKER_DOMAIN"), os.Getenv("TOKEN_BROKER_SCOPE"), os.Getenv("TOKEN_BROKER_IDENTITY")
	if domain == "" || scope == "" || identity == "" {
		return nil, errors.New("set TOKEN_BROKER_DOMAIN, TOKEN_BROKER_SCOPE and TOKEN_BROKER_IDENTITY, or GITHUB_PERSONAL_ACCESS_TOKEN")
	}
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	broker := &broker{sts: sts.NewFromConfig(cfg), http: &http.Client{Timeout: 20 * time.Second}, domain: domain, scope: scope, identity: identity}
	return &tokenSource{mint: broker.mint, now: time.Now}, nil
}

func (s *tokenSource) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	age := s.now().Sub(s.mintedAt)
	if s.token != "" && age < tokenRefreshAfter {
		return s.token, nil
	}
	token, err := s.mint(ctx)
	if err != nil {
		if s.token != "" && age < tokenUsableFor {
			slog.Warn("could not renew the GitHub token, using the current one", "error", err)
			return s.token, nil
		}
		return "", err
	}
	s.token, s.mintedAt = token, s.now()
	return token, nil
}

type webIdentityIssuer interface {
	GetWebIdentityToken(context.Context, *sts.GetWebIdentityTokenInput, ...func(*sts.Options)) (*sts.GetWebIdentityTokenOutput, error)
}

type broker struct {
	sts                     webIdentityIssuer
	http                    *http.Client
	domain, scope, identity string
}

func (b *broker) mint(ctx context.Context) (string, error) {
	issued, err := b.sts.GetWebIdentityToken(ctx, &sts.GetWebIdentityTokenInput{
		Audience:         []string{b.domain},
		SigningAlgorithm: aws.String("RS256"),
		DurationSeconds:  aws.Int32(webIdentitySecs),
	})
	if err != nil {
		return "", fmt.Errorf("STS web identity token: %w", err)
	}
	query := url.Values{"scope": {b.scope}, "identity": {b.identity}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+b.domain+"/sts/exchange?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+aws.ToString(issued.WebIdentityToken))
	resp, err := b.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("token broker: %w", err)
	}
	defer resp.Body.Close()
	var body struct {
		Token string `json:"token"`
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token broker refused identity %q for %s: %s", b.identity, b.scope, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.Token == "" {
		return "", fmt.Errorf("token broker answered without a token: %v", err)
	}
	return body.Token, nil
}

func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
