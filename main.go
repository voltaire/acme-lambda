package main

import (
	"context"
	"log"

	"github.com/go-acme/lego/v4/providers/dns/linode"
	"github.com/go-acme/lego/v4/registration"

	"github.com/go-acme/lego/v4/certcrypto"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/go-acme/lego/v4/lego"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MapDomains        []string `envconfig:"MAP_DOMAINS" default:"map.tonkat.su,oldmap.tonkat.su"`
	ACMEServer        string   `evnconfig:"ACME_SERVER" default:"https://acme-v02.api.letsencrypt.org/directory"`
	RegistrationEmail string   `envconfig:"ACME_REGISTRATION_EMAIL" default:"bsd@voltaire.sh"`

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

	user, err := loadOrCreateUser(ctx, sm, cfg.RegistrationEmail)
	if err != nil {
		log.Fatal(err.Error())
	}

	legoConfig := lego.NewConfig(user)

	legoConfig.CADirURL = cfg.ACMEServer
	legoConfig.Certificate.KeyType = certcrypto.EC384

	legoClient, err := lego.NewClient(legoConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	if user.Registration == nil {
		user.Registration, err = legoClient.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			log.Fatal(err.Error())
		}

		err = saveUser(ctx, sm, user)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	linodeDNSProvider, err := linode.NewDNSProvider()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = legoClient.Challenge.SetDNS01Provider(linodeDNSProvider)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, mapDomain := range cfg.MapDomains {
		cert, err := obtainOrRenewCertificate(ctx, sm, legoClient, cfg.RegistrationEmail, mapDomain)
		if err != nil {
			log.Fatal(err.Error())
		}

		err = uploadCertToLinode(ctx, cfg.LinodeToken, mapDomain, cert.Certificate, cert.PrivateKey)
		if err != nil {
			log.Fatal(err)
		}
	}
}
