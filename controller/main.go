package main

import (
	"context"
	"log"

	"github.com/docker/docker/client"
	"github.com/kelseyhightower/envconfig"
	"github.com/voltaire/map-cert/lego"
)

type Config struct {
	LinodeToken string `envconfig:"LINODE_TOKEN" required:"true"`

	MapDomain         string `envconfig:"MAP_DOMAIN" default:"map.tonkat.su"`
	ACMEServer        string `evnconfig:"ACME_SERVER" default:"https://acme-v02.api.letsencrypt.org/directory"`
	RegistrationEmail string `envconfig:"ACME_REGISTRATION_EMAIL" default:"bsd@voltaire.sh"`

	AWSRegion          string `envconfig:"AWS_REGION" default:"us-west-2"`
	AWSAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	AWSSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	docker, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("error starting docker client: %s", err.Error())
	}

	ctx := context.Background()
	err = lego.LetsEncryptUsingDNS(ctx, lego.Config{
		AWSRegion:          cfg.AWSRegion,
		AWSAccessKeyId:     cfg.AWSAccessKeyId,
		AWSSecretAccessKey: cfg.AWSSecretAccessKey,
		MapDomain:          cfg.MapDomain,
		ACMEServer:         cfg.ACMEServer,
		RegistrationEmail:  cfg.RegistrationEmail,
	}, docker)
	if err != nil {
		log.Fatalf("error fetching certs: %s", err.Error())
	}
}
