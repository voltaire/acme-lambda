package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

func secretsManagerPrivateKey(email string) string {
	return "mapcert-privkey-" + email
}

func createAndSaveKey(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string) (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	keyDer, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyDer,
	})

	_, err = sm.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
		Description:  aws.String("acme private key for: " + email),
		Name:         aws.String(secretsManagerPrivateKey(email)),
		SecretBinary: privateKeyPem,
	})
	if err != nil {
		return nil, err
	}

	return key, nil
}

func loadOrCreateKey(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string) (*ecdsa.PrivateKey, error) {
	privateKeyOut, err := sm.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerPrivateKey(email)),
	})
	if err != nil {
		if isKeyNotFoundError(err) {
			return createAndSaveKey(ctx, sm, email)
		}
		return nil, err
	}

	block, _ := pem.Decode(privateKeyOut.SecretBinary)
	return x509.ParseECPrivateKey(block.Bytes)
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
