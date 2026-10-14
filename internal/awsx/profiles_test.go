package awsx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListProfiles(t *testing.T) {
	dir := t.TempDir()
	config := `[default]
region = us-east-1

[profile dev]
region = eu-west-2
sso_session = my-sso
sso_account_id = 123456789012

[sso-session my-sso]
sso_start_url = https://example.awsapps.com/start

[services s3-local]
s3 = endpoint_url = http://localhost
`
	creds := `[default]
aws_access_key_id = AKIA
aws_secret_access_key = secret

[legacy]
aws_access_key_id = AKIB
`
	if err := os.WriteFile(filepath.Join(dir, "config"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "credentials"), []byte(creds), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AWS_CONFIG_FILE", filepath.Join(dir, "config"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(dir, "credentials"))

	profiles, err := ListProfiles()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	byName := map[string]Profile{}
	for _, p := range profiles {
		names = append(names, p.Name)
		byName[p.Name] = p
	}
	if len(names) != 3 || names[0] != "default" || names[1] != "dev" || names[2] != "legacy" {
		t.Fatalf("unexpected profiles: %v", names)
	}
	if !byName["dev"].IsSSO() || byName["dev"].Region() != "eu-west-2" {
		t.Errorf("dev profile parsed incorrectly: %+v", byName["dev"])
	}
	if byName["default"].IsSSO() || byName["default"].Region() != "us-east-1" {
		t.Errorf("default profile parsed incorrectly: %+v", byName["default"])
	}
	if byName["legacy"].IsSSO() {
		t.Error("legacy profile should not be SSO")
	}
}

func TestListProfilesMissingFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AWS_CONFIG_FILE", filepath.Join(dir, "config"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(dir, "credentials"))
	profiles, err := ListProfiles()
	if err != nil || len(profiles) != 0 {
		t.Errorf("expected no profiles and no error, got %v, %v", profiles, err)
	}
}
