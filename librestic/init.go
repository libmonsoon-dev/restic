package librestic

import (
	"context"

	"github.com/restic/restic/internal/backend"
	"github.com/restic/restic/internal/backend/location"
	"github.com/restic/restic/internal/backend/logger"
	"github.com/restic/restic/internal/backend/sema"
	"github.com/restic/restic/internal/debug"
	"github.com/restic/restic/internal/errors"
	"github.com/restic/restic/internal/options"
	"github.com/restic/restic/internal/repository"
	"github.com/restic/restic/internal/restic"
)

// TODO: default options
type InitOptions struct {
	Repo             string
	Password         string
	Version          uint
	TransportOptions backend.TransportOptions
	Compression      repository.CompressionMode
	PackSize         uint
	Extended         options.Options
}

func Init(ctx context.Context, opts InitOptions) error {
	be, err := create(ctx, opts)
	if err != nil {
		return errors.Fatalf("create repository at %s failed: %v\n", location.StripPassword(backends, opts.Repo), err)
	}

	s, err := repository.New(be, repository.Options{
		Compression: opts.Compression,
		PackSize:    opts.PackSize * 1024 * 1024,
	})
	if err != nil {
		return errors.Fatal(err.Error())
	}

	err = s.Init(ctx, opts.Version, opts.Password, nil)
	if err != nil {
		return errors.Fatalf("create key in repository at %s failed: %v\n", location.StripPassword(backends, opts.Repo), err)
	}

	return nil
}

// Create the backend specified by URI.
func create(ctx context.Context, opts InitOptions) (restic.Backend, error) {
	loc, err := location.Parse(backends, opts.Repo)
	if err != nil {
		return nil, err
	}

	cfg, err := parseConfig(loc, opts.Extended)
	if err != nil {
		return nil, err
	}

	rt, err := backend.Transport(opts.TransportOptions)
	if err != nil {
		return nil, err
	}

	factory := backends.Lookup(loc.Scheme)
	if factory == nil {
		return nil, errors.Fatalf("invalid backend: %q", loc.Scheme)
	}

	be, err := factory.Create(ctx, cfg, rt, nil)
	if err != nil {
		return nil, err
	}

	return logger.New(sema.NewBackend(be)), nil
}

func parseConfig(loc location.Location, opts options.Options) (interface{}, error) {
	cfg := loc.Config
	if cfg, ok := cfg.(restic.ApplyEnvironmenter); ok {
		cfg.ApplyEnvironment("")
	}

	// only apply options for a particular backend here
	opts = opts.Extract(loc.Scheme)
	if err := opts.Apply(loc.Scheme, cfg); err != nil {
		return nil, err
	}

	debug.Log("opening %v repository at %#v", loc.Scheme, cfg)
	return cfg, nil
}
