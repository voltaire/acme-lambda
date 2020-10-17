package lego

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func buildLegoEnv(cfg Config) []string {
	return []string{
		"AWS_REGION=" + cfg.AWSRegion,
		"AWS_ACCESS_KEY_ID=" + cfg.AWSAccessKeyId,
		"AWS_SECRET_ACCESS_KEY=" + cfg.AWSSecretAccessKey,
	}
}

func buildLegoCmd(cfg Config) []string {
	return []string{
		"lego",
		"--domains " + cfg.MapDomain,
		"--server " + cfg.ACMEServer,
		"--accept-os",
		"--email " + cfg.RegistrationEmail,
		"--path /output",
		"--dns route53",
	}
}

// Config holds required configuration options
type Config struct {
	AWSRegion          string
	AWSAccessKeyId     string
	AWSSecretAccessKey string

	ACMEServer        string
	MapDomain         string
	RegistrationEmail string
}

func LetsEncryptUsingDNS(ctx context.Context, cfg Config, docker client.APIClient) error {
	container, err := docker.ContainerCreate(ctx, &container.Config{
		Image: "docker.io/goacme/lego",
		Env:   buildLegoEnv(cfg),
		Labels: map[string]string{
			"service": "map-cert",
		},
		Cmd: buildLegoCmd(cfg),
	}, &container.HostConfig{
		AutoRemove: true,
		LogConfig: container.LogConfig{
			Type: "awslogs",
			Config: map[string]string{
				"awslogs-group":        "map-cert-lego",
				"awslogs-create-group": "true",
				"awslogs-region":       cfg.AWSRegion,
			},
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: "map_cert",
				Target: "/output",
			},
		},
	}, nil, "")
	if err != nil {
		return err
	}

	return docker.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
}
