// Package awsx wires up AWS configuration, profiles and clients.
package awsx

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// Profile is a named profile found in the shared AWS config or credentials files.
type Profile struct {
	Name  string
	Props map[string]string
}

// IsSSO reports whether the profile authenticates through IAM Identity Center.
func (p Profile) IsSSO() bool {
	_, session := p.Props["sso_session"]
	_, startURL := p.Props["sso_start_url"]
	return session || startURL
}

// Region returns the region configured on the profile, if any.
func (p Profile) Region() string {
	return p.Props["region"]
}

// ListProfiles returns every profile declared in the shared config and
// credentials files, honouring AWS_CONFIG_FILE and AWS_SHARED_CREDENTIALS_FILE.
func ListProfiles() ([]Profile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath := envOr("AWS_CONFIG_FILE", filepath.Join(home, ".aws", "config"))
	credsPath := envOr("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(home, ".aws", "credentials"))

	found := map[string]*Profile{}
	if err := parseProfiles(configPath, true, found); err != nil {
		return nil, err
	}
	if err := parseProfiles(credsPath, false, found); err != nil {
		return nil, err
	}

	profiles := make([]Profile, 0, len(found))
	for _, p := range found {
		profiles = append(profiles, *p)
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	return profiles, nil
}

// LookupProfile finds a profile by name, if it is declared in the shared files.
func LookupProfile(name string) (Profile, bool) {
	profiles, err := ListProfiles()
	if err != nil {
		return Profile{}, false
	}
	for _, p := range profiles {
		if p.Name == name {
			return p, true
		}
	}
	return Profile{}, false
}

// ResolveProfile picks the profile to use: the flag value, then AWS_PROFILE,
// then an interactive choice from the profiles on disk. An empty result means
// the default credential chain should be used.
func ResolveProfile(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if env := os.Getenv("AWS_PROFILE"); env != "" {
		ui.Info("Using AWS profile from AWS_PROFILE: %s", env)
		return env, nil
	}

	profiles, err := ListProfiles()
	if err != nil {
		return "", fmt.Errorf("reading AWS profiles: %w", err)
	}
	if len(profiles) == 0 {
		ui.Warn("No AWS profiles found in ~/.aws; using the default credential chain")
		return "", nil
	}

	options := make([]ui.Option[string], 0, len(profiles))
	for _, p := range profiles {
		label := p.Name
		var notes []string
		if region := p.Region(); region != "" {
			notes = append(notes, region)
		}
		if p.IsSSO() {
			notes = append(notes, "sso")
		}
		if len(notes) > 0 {
			label = fmt.Sprintf("%s (%s)", p.Name, strings.Join(notes, ", "))
		}
		options = append(options, ui.Option[string]{Label: label, Value: p.Name})
	}
	return ui.Select("Select an AWS profile (type to filter)", options)
}

// parseProfiles reads an INI-style AWS file. In the config file profile
// sections are written as "[profile name]" (except "[default]"); in the
// credentials file they are plain "[name]".
func parseProfiles(path string, configFile bool, into map[string]*Profile) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	var current *Profile
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			name := strings.TrimSpace(line[1 : len(line)-1])
			if configFile && name != "default" {
				if !strings.HasPrefix(name, "profile ") {
					current = nil // sso-session, services, etc.
					continue
				}
				name = strings.TrimSpace(strings.TrimPrefix(name, "profile "))
			}
			if _, ok := into[name]; !ok {
				into[name] = &Profile{Name: name, Props: map[string]string{}}
			}
			current = into[name]
			continue
		}
		if current == nil {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		if _, exists := current.Props[key]; !exists {
			current.Props[key] = strings.TrimSpace(value)
		}
	}
	return scanner.Err()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
