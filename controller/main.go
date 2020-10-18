package main

import (
	"context"
	"log"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/providers/dns/linode"

	"github.com/go-acme/lego/v4/certcrypto"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/go-acme/lego/v4/lego"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MapDomain         string `envconfig:"MAP_DOMAIN" default:"map.tonkat.su"`
	ACMEServer        string `evnconfig:"ACME_SERVER" default:"https://acme-v02.api.letsencrypt.org/directory"`
	RegistrationEmail string `envconfig:"ACME_REGISTRATION_EMAIL" default:"bsd@voltaire.sh"`

	LinodeToken string `envconfig:"LINODE_TOKEN" required:"true"`

	AWSRegion string `envconfig:"AWS_REGION" default:"us-west-2"`
	_         string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	_         string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx := context.Background()
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err.Error())
	}
	sm := secretsmanager.New(sess)

	pkey, err := loadOrCreatePrivateKey(ctx, sm, cfg.RegistrationEmail)
	if err != nil {
		log.Fatal(err.Error())
	}

	legoConfig := lego.NewConfig(&user{
		Email:      cfg.RegistrationEmail,
		PrivateKey: pkey,
	})

	legoConfig.CADirURL = cfg.ACMEServer
	legoConfig.Certificate.KeyType = certcrypto.EC384

	legoClient, err := lego.NewClient(legoConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	linodeDNSProvider, err := linode.NewDNSProvider()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = legoClient.Challenge.SetDNS01Provider(linodeDNSProvider)
	if err != nil {
		log.Fatal(err.Error())
	}

	request := certificate.ObtainRequest{
		Domains: []string{cfg.MapDomain},
		Bundle:  true,
	}

	cert, err := legoClient.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = uploadCertToLinode(ctx, cfg.LinodeToken, cfg.MapDomain, cert.Certificate, cert.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
}
