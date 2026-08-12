// Package cli implements the herenow command-line interface. In v0 the CLI talks
// to the local store directly (single-user, no server round-trip to publish);
// `serve` runs the viewer. A remote/API client is added with the MCP connector.
package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	herenowv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/herenow/v1"
	"github.com/agarwalvivek29/here.now/services/herenow-api/internal/api"
	"github.com/agarwalvivek29/here.now/services/herenow-api/internal/config"
	"github.com/agarwalvivek29/here.now/services/herenow-api/internal/domain"
	"github.com/agarwalvivek29/here.now/services/herenow-api/internal/infra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const usage = `here.now — self-hostable host for AI-generated artifacts

Usage:
  herenow login             set up your local identity + session token
  herenow publish <file>    publish an artifact, print its link
  herenow ls                list your artifacts
  herenow serve             run the viewer server
  herenow audit verify      verify the audit-log hash chain

Docs: docs/PLAN.md is legacy; see docs/PRODUCT.md, docs/ARCHITECTURE.md
`

func Run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return nil
	}
	switch args[0] {
	case "login":
		return login()
	case "publish":
		return publish(args[1:])
	case "ls":
		return ls()
	case "serve":
		return serve()
	case "audit":
		return audit(args[1:])
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	default:
		return fmt.Errorf("unknown command %q (try: herenow help)", args[0])
	}
}

func open(c config.Config) (*infra.FileStore, *infra.BlobFS, error) {
	st, err := infra.NewFileStore(filepath.Join(c.DataDir, "meta"))
	if err != nil {
		return nil, nil, err
	}
	bl, err := infra.NewBlobFS(filepath.Join(c.DataDir, "blobs"))
	if err != nil {
		return nil, nil, err
	}
	return st, bl, nil
}

func login() error {
	c, err := config.Load()
	if err != nil {
		return err
	}
	if c.Sub == "" {
		u := os.Getenv("USER")
		if u == "" {
			u = "local"
		}
		c.Sub = "local:" + u
		c.Email = u + "@localhost"
	}
	if c.Token == "" {
		c.Token = domain.NewSlug() + domain.NewSlug()
	}
	if err := config.Save(c); err != nil {
		return err
	}
	fmt.Printf("logged in as %s\n", c.Email)
	fmt.Printf("data dir:    %s\n", c.DataDir)
	fmt.Printf("browser login (sets session cookie once):\n  %s/login?token=%s\n", c.BaseURL, c.Token)
	return nil
}

func publish(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: herenow publish <file>")
	}
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()
	c, err := config.Load()
	if err != nil {
		return err
	}
	if c.Sub == "" {
		return fmt.Errorf("not logged in — run: herenow login")
	}
	st, bl, err := open(c)
	if err != nil {
		return err
	}
	art := &herenowv1.Artifact{
		Slug:        domain.NewSlug(),
		OwnerSub:    c.Sub,
		Title:       filepath.Base(args[0]),
		Visibility:  herenowv1.Visibility_VISIBILITY_PRIVATE, // private by default
		ContentType: "text/html; charset=utf-8",
		CreatedAt:   timestamppb.Now(),
	}
	if err := bl.Put(art.GetSlug(), f); err != nil {
		return err
	}
	if err := st.PutArtifact(art); err != nil {
		return err
	}
	_ = st.Append(&herenowv1.AuditEvent{
		Ts: timestamppb.Now(), PrincipalSub: c.Sub, Slug: art.GetSlug(),
		Action: herenowv1.AuditAction_AUDIT_ACTION_PUBLISH, Allowed: true,
	})
	fmt.Printf("%s/a/%s\n", c.BaseURL, art.GetSlug())
	return nil
}

func ls() error {
	c, err := config.Load()
	if err != nil {
		return err
	}
	st, _, err := open(c)
	if err != nil {
		return err
	}
	arts, err := st.ListByOwner(c.Sub)
	if err != nil {
		return err
	}
	if len(arts) == 0 {
		fmt.Println("no artifacts yet — herenow publish <file>")
		return nil
	}
	for _, a := range arts {
		fmt.Printf("%s  %-20s  %s/a/%s  %s\n",
			a.GetCreatedAt().AsTime().Format("2006-01-02 15:04"),
			a.GetVisibility().String(), c.BaseURL, a.GetSlug(), a.GetTitle())
	}
	return nil
}

func audit(args []string) error {
	if len(args) < 1 || args[0] != "verify" {
		return fmt.Errorf("usage: herenow audit verify")
	}
	c, err := config.Load()
	if err != nil {
		return err
	}
	n, err := infra.VerifyAuditLog(filepath.Join(c.DataDir, "meta"))
	if err != nil {
		return err
	}
	fmt.Printf("audit chain OK (%d events)\n", n)
	return nil
}

func serve() error {
	c, err := config.Load()
	if err != nil {
		return err
	}
	st, bl, err := open(c)
	if err != nil {
		return err
	}
	srv := &api.Server{Store: st, Blob: bl}
	// Config-driven auth selection: OIDC browser SSO when configured (ADR-0007),
	// otherwise the Local single-token adapter for zero-dependency/dev deploys.
	if c.OIDCEnabled() {
		if c.SessionSecret == "" {
			return fmt.Errorf("OIDC configured but ARTIFACTA_SESSION_SECRET is empty — set it to key session cookies")
		}
		secure := strings.HasPrefix(strings.ToLower(c.BaseURL), "https://")
		p, err := api.NewOIDCProvider(context.Background(),
			c.OIDCIssuer, c.OIDCClientID, c.OIDCClientSecret, c.OIDCRedirectURL, c.SessionSecret, secure)
		if err != nil {
			return fmt.Errorf("oidc setup: %w", err)
		}
		srv.Auth = p
		srv.OIDC = p
		fmt.Printf("auth: oidc browser sso (issuer %s)\n", c.OIDCIssuer)
	} else {
		srv.Auth = &api.Local{Token: c.Token, ID: c.Identity()}
		fmt.Printf("auth: local single-token adapter\n")
	}
	fmt.Printf("here.now serving on %s  (base URL %s)\n", c.Addr, c.BaseURL)
	return http.ListenAndServe(c.Addr, srv.Routes())
}
