// Command secret-env runs another command with the string values of an AWS Secrets
// Manager secret added to its environment.
//
// With SECRET_ENV_ID unset it execs the command unchanged. With it set, it reads that
// secret (a name or an ARN) as a JSON object of strings, exports each one, and execs the
// command. Anything else - an unreadable secret, a value that is not a string, a key
// that is not a variable name - stops it before the command starts, so a server never
// runs half-configured.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const secretIDVar = "SECRET_ENV_ID"

var variableName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "secret-env: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: secret-env <command> [args...]")
	}
	path, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	env := os.Environ()
	if id := os.Getenv(secretIDVar); id != "" {
		secret, err := readSecret(context.Background(), id)
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", id, err)
		}
		values, err := parseSecret(secret)
		if err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
		env = merge(env, values)
	}
	return syscall.Exec(path, args, env)
}

func readSecret(ctx context.Context, id string) (string, error) {
	var opts []func(*config.LoadOptions) error
	if region := regionOf(id); region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return "", err
	}
	out, err := secretsmanager.NewFromConfig(cfg).GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: aws.String(id)})
	if err != nil {
		return "", err
	}
	if out.SecretString == nil {
		return "", errors.New("the secret has no string value")
	}
	return *out.SecretString, nil
}

func regionOf(id string) string {
	parts := strings.Split(id, ":")
	if len(parts) > 3 && parts[0] == "arn" {
		return parts[3]
	}
	return ""
}

func parseSecret(secret string) (map[string]string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(secret), &raw); err != nil || raw == nil {
		return nil, errors.New("the secret is not a JSON object")
	}
	values := make(map[string]string, len(raw))
	for name, value := range raw {
		if !variableName.MatchString(name) {
			return nil, fmt.Errorf("%q is not a variable name", name)
		}
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%q is not a string", name)
		}
		values[name] = s
	}
	return values, nil
}

func merge(env []string, values map[string]string) []string {
	merged := make([]string, 0, len(env)+len(values))
	for _, entry := range env {
		name, _, _ := strings.Cut(entry, "=")
		if _, replaced := values[name]; !replaced {
			merged = append(merged, entry)
		}
	}
	for name, value := range values {
		merged = append(merged, name+"="+value)
	}
	return merged
}
