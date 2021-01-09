package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/go-acme/lego/v4/lego"
	route53provider "github.com/go-acme/lego/v4/providers/dns/route53"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DataBucketName        string   `required:"true" split_words:"true"`
	LegoDataPrefix        string   `required:"true" split_words:"true"`
	RenewIfWithin         string   `default:"336h" split_words:"true"`
	AcmeRegistrationEmail string   `required:"true" split_words:"true"`
	Domains               []string `required:"true"`
	HostedZoneId          string   `required:"true" split_words:"true"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	sm := secretsmanager.New(sess)
	route53Client := route53.New(sess)
	dnsProvider, err := route53provider.NewDNSProviderConfig(&route53provider.Config{
		Client:       route53Client,
		HostedZoneID: cfg.HostedZoneId,
	})
	if err != nil {
		log.Fatal(err)
	}

	lambda.Start(func(ctx context.Context) error {
		acmeUser, err := loadOrCreateUser(ctx, sm, cfg.AcmeRegistrationEmail)
		if err != nil {
			return err
		}

		legoClient, err := lego.NewClient(lego.NewConfig(acmeUser))
		if err != nil {
			return err
		}

		err = legoClient.Challenge.SetDNS01Provider(dnsProvider)
		if err != nil {
			return err
		}
		return nil
	})
}
