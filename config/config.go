package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	defaultMsBetweenPoll = 30000
	defaultVcs           = "git"
	defaultBaseUrl       = "{url}/blob/master/{path}{anchor}"
	defaultAnchor        = "#L{line}"
)

type UrlPattern struct {
	BaseUrl string `json:"base-url"`
	Anchor  string `json:"anchor"`
}

type Repo struct {
	Url              string         `json:"url"`
	MsBetweenPolls   int            `json:"ms-between-poll"`
	Vcs              string         `json:"vcs"`
	VcsConfigMessage *SecretMessage `json:"vcs-config"`
	UrlPattern       *UrlPattern    `json:"url-pattern"`
}

type Config struct {
	DbPath string           `json:"dbpath"`
	Repos  map[string]*Repo `json:"repos"`
}

// SecretMessage is just like json.RawMessage but it will not
// marshal its value as JSON. This is to ensure that vcs-config
// is not marshalled into JSON and send to the UI.
type SecretMessage []byte

// This always marshals to an empty object.
func (s *SecretMessage) MarshalJSON() ([]byte, error) {
	return []byte("{}"), nil
}

// See http://golang.org/pkg/encoding/json/#RawMessage.UnmarshalJSON
func (s *SecretMessage) UnmarshalJSON(b []byte) error {
	if b == nil {
		return errors.New("SecretMessage: UnmarshalJSON on nil pointer")
	}
	*s = append((*s)[0:0], b...)
	return nil
}

// Get the JSON encode vcs-config for this repo. This returns nil if
// the repo doesn't declare a vcs-config.
func (r *Repo) VcsConfig() []byte {
	if r.VcsConfigMessage == nil {
		return nil
	}
	return *r.VcsConfigMessage
}

// Populate missing config values with default values.
func initRepo(r *Repo) {
	if r.MsBetweenPolls == 0 {
		r.MsBetweenPolls = defaultMsBetweenPoll
	}

	if r.Vcs == "" {
		r.Vcs = defaultVcs
	}

	if r.UrlPattern == nil {
		r.UrlPattern = &UrlPattern{
			BaseUrl: defaultBaseUrl,
			Anchor:  defaultAnchor,
		}
	} else {
		if r.UrlPattern.BaseUrl == "" {
			r.UrlPattern.BaseUrl = defaultBaseUrl
		}

		if r.UrlPattern.Anchor == "" {
			r.UrlPattern.Anchor = defaultAnchor
		}
	}
}

func (c *Config) LoadFromFile(filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := json.NewDecoder(r).Decode(c); err != nil {
		return err
	}

	if !filepath.IsAbs(c.DbPath) {
		path, err := filepath.Abs(
			filepath.Join(filepath.Dir(filename), c.DbPath))
		if err != nil {
			return err
		}
		c.DbPath = path
	}

	for _, repo := range c.Repos {
		initRepo(repo)
	}

	return nil
}

func (c *Config) ToJsonString() (string, error) {
	b, err := json.Marshal(c.Repos)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
