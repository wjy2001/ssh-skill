package ssh

import (
	"testing"

	"ssh-skill/internal/types"
)

func TestFindServer(t *testing.T) {
	vault := &types.Vault{
		Servers: []types.ServerConfig{
			{ID: "prod", Name: "Production", Host: "10.0.0.1"},
			{ID: "staging", Name: "Staging", Host: "10.0.0.2"},
		},
	}

	// Found.
	srv, err := FindServer(vault, "prod")
	if err != nil {
		t.Fatalf("FindServer prod: %v", err)
	}
	if srv.Name != "Production" {
		t.Fatalf("expected Production, got %s", srv.Name)
	}

	// Not found.
	_, err = FindServer(vault, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent server")
	}
}

func TestFindServerEmptyVault(t *testing.T) {
	vault := &types.Vault{Servers: []types.ServerConfig{}}
	_, err := FindServer(vault, "any")
	if err == nil {
		t.Fatal("expected error for empty vault")
	}
}

func TestBuildAuthMethodsPassword(t *testing.T) {
	cfg := &types.ServerConfig{
		User: "root",
		Auth: types.AuthConfig{
			Method:            types.AuthPassword,
			EncryptedPassword: "test-password",
		},
	}
	methods, err := buildAuthMethods(cfg)
	if err != nil {
		t.Fatalf("buildAuthMethods password: %v", err)
	}
	if len(methods) != 1 {
		t.Fatalf("expected 1 auth method, got %d", len(methods))
	}
}

func TestBuildAuthMethodsPasswordEmpty(t *testing.T) {
	cfg := &types.ServerConfig{
		User: "root",
		Auth: types.AuthConfig{
			Method:            types.AuthPassword,
			EncryptedPassword: "",
		},
	}
	_, err := buildAuthMethods(cfg)
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestBuildAuthMethodsKeyMissingPath(t *testing.T) {
	cfg := &types.ServerConfig{
		User: "root",
		Auth: types.AuthConfig{
			Method:         types.AuthKey,
			PrivateKeyPath: "",
		},
	}
	_, err := buildAuthMethods(cfg)
	if err == nil {
		t.Fatal("expected error for empty key path")
	}
}

func TestBuildAuthMethodsUnknownMethod(t *testing.T) {
	cfg := &types.ServerConfig{
		User: "root",
		Auth: types.AuthConfig{
			Method: "unknown",
		},
	}
	_, err := buildAuthMethods(cfg)
	if err == nil {
		t.Fatal("expected error for unknown auth method")
	}
}

func TestBuildAuthMethodsAgentNoSocket(t *testing.T) {
	// Ensure SSH_AUTH_SOCK is not set in test environment.
	t.Setenv("SSH_AUTH_SOCK", "")

	cfg := &types.ServerConfig{
		User: "root",
		Auth: types.AuthConfig{
			Method: types.AuthAgent,
		},
	}
	_, err := buildAuthMethods(cfg)
	if err == nil {
		t.Fatal("expected error when SSH_AUTH_SOCK not set")
	}
}
