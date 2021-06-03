package synchronize

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/bilibili-base/powermock/pkg/util/logger"
)

type Service struct {
	cfg *Config

	registerer prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
	Enable     bool
	StorageDir string
	Repository []*Repository
}

type Repository struct {
	Address string
	Branch  string
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Enable: true,
	}
}

// IsEnabled is used to return whether the current component is enabled
// This attribute is required in pluggable components
func (c *Config) IsEnabled() bool {
	return c.Enable
}

// RegisterFlags is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.BoolVar(&c.Enable, prefix+"enable", c.Enable, "define whether the component is enabled")
	f.StringVar(&c.StorageDir, prefix+"storageDir", c.StorageDir, "local storage dir of repositories")
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	if c.StorageDir == "" {
		return errors.New("storage dir is required")
	}
	return nil
}

// New is used to init service
func New(cfg *Config, logger logger.Logger, registerer prometheus.Registerer) (*Service, error) {
	service := &Service{
		cfg:        cfg,
		registerer: registerer,
		Logger:     logger.NewLogger("Service"),
	}
	return service, nil
}

func (s *Service) Synchronize(ctx context.Context, callback func(repository string, updated bool, err error) error) error {
	for _, repo := range s.cfg.Repository {
		repository := repo.Address
		branch := repo.Branch
		if branch == "" {
			branch = "master"
		}
		location := filepath.Join(s.cfg.StorageDir, getRepositoryPathFromUrl(repository))
		s.LogInfo(map[string]interface{}{
			"location":   location,
			"repository": repository,
			"branch":     branch,
		}, "start to synchronize repository")
		updated, err := s.SynchronizeRepository(ctx, repository, branch, location)
		if callback != nil {
			if err := callback(repository, updated, err); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) SynchronizeRepository(ctx context.Context, repository string, branch string, location string) (updated bool, err error) {
	_, err = os.Stat(filepath.Join(location, ".git"))
	if err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
		s.LogInfo(map[string]interface{}{
			"repository": repository,
			"branch":     branch,
			"location":   location,
		}, "start to clone repository")
		if err := cloneRepository(ctx, repository, branch, location); err != nil {
			return false, err
		}
		return true, nil
	}
	s.LogInfo(map[string]interface{}{
		"repository": repository,
		"branch":     branch,
		"location":   location,
	}, "start to get local commits")
	localCommits, err := listRepositoryCommits(ctx, "HEAD", location)
	if err != nil {
		return false, err
	}
	s.LogInfo(map[string]interface{}{
		"repository": repository,
		"location":   location,
		"branch":     branch,
	}, "start to get remote commits")
	remoteCommits, err := listRepositoryCommits(ctx, "origin/"+branch, location)
	if err != nil {
		return false, err
	}
	s.LogInfo(map[string]interface{}{
		"repository":    repository,
		"location":      location,
		"remoteCommits": len(remoteCommits),
		"localCommits":  len(localCommits),
		"branch":        branch,
	}, "start to compare local and remote commits")
	if len(localCommits) == len(remoteCommits) {
		s.LogInfo(map[string]interface{}{
			"repository": repository,
			"location":   location,
			"branch":     branch,
		}, "no need to update local repository")
		return false, nil
	}
	s.LogInfo(map[string]interface{}{
		"repository": repository,
		"location":   location,
		"branch":     branch,
	}, "start to pull repository")
	if err := pullRepository(ctx, repository, branch, location); err != nil {
		return false, err
	}
	s.LogInfo(map[string]interface{}{
		"repository": repository,
		"location":   location,
		"branch":     branch,
	}, "repository updated")
	return true, nil
}

// Start is used to start the service
func (s *Service) Start(ctx context.Context, cancelFunc context.CancelFunc) error {
	return nil
}

func execCommand(ctx context.Context, args []string, dir string, stdout io.Writer, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	cmd.Dir = dir
	fmt.Println(">", cmd.String())
	return cmd.Run()
}

func cloneRepository(ctx context.Context, address string, branch string, location string) error {
	if branch == "" {
		branch = "master"
	}
	return execCommand(ctx, []string{
		"clone",
		"-b",
		branch,
		address,
		location,
	}, "", os.Stdout, os.Stderr)
}

func pullRepository(ctx context.Context, address string, branch string, location string) error {
	return execCommand(ctx, []string{
		"pull",
		"origin",
		branch,
	}, location, os.Stdout, os.Stderr)
}

func listRepositoryCommits(ctx context.Context, branch string, location string) ([]string, error) {
	var buffer bytes.Buffer
	if err := execCommand(ctx, []string{
		"log",
		branch,
		"--pretty=format:%H",
	}, location, &buffer, os.Stderr); err != nil {
		return nil, err
	}
	commits := strings.Split(buffer.String(), "\n")
	return commits, nil
}

func getRepositoryPathFromUrl(s string) string {
	var projectPath string
	s = strings.TrimSuffix(s, ".git")
	if strings.HasPrefix(s, "git@") {
		parts := strings.Split(s, ":")
		if len(parts) == 2 {
			projectPath = parts[1]
		}
	} else {
		parsed, err := url.Parse(s)
		if err != nil {
			return ""
		}
		projectPath = parsed.Path
	}
	return strings.TrimPrefix(projectPath, "/")
}
