package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

const gcloudTimeout = 60 * time.Second

// gcloudIdentity is the active gcloud user account and project.
// Credentials come from `gcloud auth login`, not Application Default Credentials.
type gcloudIdentity struct {
	Account string
	Project string
	token   *oauth2.Token
}

func (id gcloudIdentity) clientOptions() []option.ClientOption {
	source := oauth2.ReuseTokenSource(id.token, gcloudTokenSource{})
	return []option.ClientOption{
		option.WithTokenSource(source),
		option.WithQuotaProject(id.Project),
	}
}

type gcloudTokenSource struct{}

func (gcloudTokenSource) Token() (*oauth2.Token, error) {
	id, err := loadGcloudIdentity(context.Background())
	if err != nil {
		return nil, err
	}
	return id.token, nil
}

func loadGcloudIdentity(ctx context.Context) (gcloudIdentity, error) {
	ctx, cancel := context.WithTimeout(ctx, gcloudTimeout)
	defer cancel()
	raw, err := runGcloud(ctx, "config", "config-helper", "--format=json")
	if err != nil {
		return gcloudIdentity{}, err
	}
	return parseGcloudConfig(raw)
}

func parseGcloudConfig(raw []byte) (gcloudIdentity, error) {
	var parsed struct {
		Configuration struct {
			Properties struct {
				Core struct {
					Account string `json:"account"`
					Project string `json:"project"`
				} `json:"core"`
			} `json:"properties"`
		} `json:"configuration"`
		Credential struct {
			AccessToken string `json:"access_token"`
			TokenExpiry string `json:"token_expiry"`
		} `json:"credential"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return gcloudIdentity{}, errors.New("could not read gcloud credentials")
	}

	account := strings.TrimSpace(parsed.Configuration.Properties.Core.Account)
	project := strings.TrimSpace(parsed.Configuration.Properties.Core.Project)
	if project == "(unset)" {
		project = ""
	}
	token := strings.TrimSpace(parsed.Credential.AccessToken)
	if account == "" || token == "" {
		return gcloudIdentity{}, errors.New("gcloud has no active account. Run `gcloud auth login`")
	}
	if project == "" {
		return gcloudIdentity{}, errors.New("gcloud has no active project. Run `gcloud config set project PROJECT`")
	}

	expiry, err := time.Parse(time.RFC3339Nano, parsed.Credential.TokenExpiry)
	if err != nil {
		expiry, err = time.Parse(time.RFC3339, parsed.Credential.TokenExpiry)
	}
	if err != nil {
		return gcloudIdentity{}, errors.New("gcloud did not return a token expiry. Run `gcloud auth login`")
	}

	return gcloudIdentity{
		Account: account,
		Project: project,
		token: &oauth2.Token{
			AccessToken: token,
			TokenType:   "Bearer",
			Expiry:      expiry,
		},
	}, nil
}

var runGcloud = defaultRunGcloud

func defaultRunGcloud(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Env = append(os.Environ(), "CLOUDSDK_CORE_DISABLE_PROMPTS=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, gcloudCommandError(err, stderr.String())
	}
	return out, nil
}

func gcloudCommandError(err error, stderr string) error {
	if errors.Is(err, exec.ErrNotFound) {
		return errors.New("gcloud is not installed or not on PATH")
	}
	msg := strings.TrimSpace(stderr)
	if msg == "" {
		msg = err.Error()
	}
	if len(msg) > 500 {
		msg = msg[:500]
	}
	return fmt.Errorf("failed to read the active gcloud account: %s", msg)
}
