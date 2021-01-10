package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/crypto/acme"
)

type Config struct {
	DataBucketName        string   `required:"true" split_words:"true"`
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
	//route53Client := route53.New(sess)

	lambda.Start(func(ctx context.Context) error {
		privateKey, err := loadOrCreateKey(ctx, sm, cfg.AcmeRegistrationEmail)
		if err != nil {
			return err
		}

		acmeClient := &acme.Client{
			Key:       privateKey,
			UserAgent: "https://github.com/voltaire/acme-lambda",
		}

		err = registerUser(ctx, acmeClient, cfg.AcmeRegistrationEmail)
		if err != nil {
			return err
		}

		return nil
	})
}
