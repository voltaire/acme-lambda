package main

import (
	"context"
	"crypto"
	"crypto/ed25519"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

func secretsManagerKeyName(email string) string {
	return "ACME private key for " + email
}

func isKeyNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	awsErr, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	if awsErr.Code() == secretsmanager.ErrCodeResourceNotFoundException {
		return true
	}
	return false
}

func createPrivateKey(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string) (crypto.PrivateKey, error) {
	_, key, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	createSecretInput := &secretsmanager.CreateSecretInput{
		Description:  aws.String("acme private key for email: " + email),
		Name:         aws.String(secretsManagerKeyName(email)),
		SecretBinary: []byte(key),
		Tags: []*secretsmanager.Tag{
			{
				Key:   aws.String("service"),
				Value: aws.String("map-cert"),
			},
		},
	}

	_, err = sm.CreateSecretWithContext(ctx, createSecretInput)
	return key, err
}

func loadPrivateKey(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string) (crypto.PrivateKey, error) {
	getSecretInput := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerKeyName(email)),
	}
	out, err := sm.GetSecretValueWithContext(ctx, getSecretInput)
	if err != nil {
		return nil, err
	}

	return ed25519.PrivateKey(out.SecretBinary), nil
}

func loadOrCreatePrivateKey(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string) (crypto.PrivateKey, error) {
	key, err := loadPrivateKey(ctx, sm, email)
	if err == nil || !isKeyNotFoundError(err) {
		return key, err
	}

	return createPrivateKey(ctx, sm, email)
}
