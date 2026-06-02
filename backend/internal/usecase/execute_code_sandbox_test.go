package usecase

import (
	"os"
	"strings"
	"testing"
)

func TestIsSensitiveEnvKey(t *testing.T) {
	sensitive := []string{
		"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_SESSION_TOKEN",
		"DATABASE_URL",
		"DB_PASSWORD",
		"COGNITO_CLIENT_SECRET",
		"SES_FROM_ADDRESS",
		"BEDROCK_MODEL_ID",
		"SOME_API_TOKEN",
		"MY_PRIVATE_KEY",
	}
	for _, k := range sensitive {
		if !isSensitiveEnvKey(k) {
			t.Errorf("%q should be treated as sensitive", k)
		}
	}
	benign := []string{"PATH", "HOME", "LANG", "GOCACHE", "GOROOT", "GO111MODULE", "TERM"}
	for _, k := range benign {
		if isSensitiveEnvKey(k) {
			t.Errorf("%q should NOT be treated as sensitive", k)
		}
	}
}

func TestSandboxEnv_StripsSecretsKeepsBenign(t *testing.T) {
	t.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "/v2/credentials/abc")
	t.Setenv("DATABASE_URL", "postgresql://user:pw@host/db")
	t.Setenv("COGNITO_CLIENT_SECRET", "super-secret")
	t.Setenv("SANDBOX_TEST_BENIGN", "ok")

	env := sandboxEnv("EXTRA=1")
	joined := strings.Join(env, "\n")

	for _, leaked := range []string{
		"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI",
		"DATABASE_URL",
		"COGNITO_CLIENT_SECRET",
		"super-secret",
	} {
		if strings.Contains(joined, leaked) {
			t.Errorf("sandbox env must not contain %q", leaked)
		}
	}
	if !strings.Contains(joined, "SANDBOX_TEST_BENIGN=ok") {
		t.Error("benign var should be preserved")
	}
	if !strings.Contains(joined, "EXTRA=1") {
		t.Error("extra var should be appended")
	}
	// PATH のような go/php 実行に必要な env は残ること。
	if os.Getenv("PATH") != "" && !strings.Contains(joined, "PATH=") {
		t.Error("PATH should be preserved for the interpreter/toolchain")
	}
}
