package librestic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/restic/restic/librestic"
)

func TestInit(t *testing.T) {
	dir := t.TempDir()

	options := librestic.InitOptions{
		Repo:     dir,
		Version:  librestic.StableRepoVersion,
		Password: "passwd",
	}

	err := librestic.Init(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}

	err = librestic.Init(context.Background(), options)

	if err == nil || !strings.Contains(err.Error(), "config file already exists") {
		t.Fatal(err)
	}
}
