package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
)

func secretsManagerCertificateName(email, domain string) string {
	return fmt.Sprintf("Certificate for %s:%s", email, domain)
}

func getPreviousCertificate(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string, domain string) (*certificate.Resource, error) {
	getCertificateInput := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerCertificateName(email, domain)),
	}
	out, err := sm.GetSecretValueWithContext(ctx, getCertificateInput)
	if err != nil {
		return nil, err
	}

	resource := new(certificate.Resource)
	err = json.Unmarshal(out.SecretBinary, resource)
	return resource, err
}

func storeCertificate(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string, resource *certificate.Resource) error {
	bs, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	createSecretInput := &secretsmanager.CreateSecretInput{
		Description:  aws.String(""),
		Name:         aws.String(secretsManagerCertificateName(email, resource.Domain)),
		SecretBinary: bs,
		Tags: []*secretsmanager.Tag{
			{
				Key:   aws.String("service"),
				Value: aws.String("map-cert"),
			},
		},
	}
	_, err = sm.CreateSecretWithContext(ctx, createSecretInput)
	return err
}

func obtainOrRenewCertificate(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, legoClient *lego.Client, email, domain string) (*certificate.Resource, error) {
	resource, err := getPreviousCertificate(ctx, sm, email, domain)
	if err != nil && !isKeyNotFoundError(err) {
		return nil, err
	}

	if resource == nil {
		request := certificate.ObtainRequest{
			Domains: []string{domain},
		}

		resource, err = legoClient.Certificate.Obtain(request)
		if err != nil {
			return nil, err
		}
	} else {
		resource, err = legoClient.Certificate.Renew(*resource, false, false, "")
		if err != nil {
			return nil, err
		}
	}

	err = storeCertificate(ctx, sm, email, resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}
