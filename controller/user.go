package main

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/go-acme/lego/v4/registration"
)

type user struct {
	Email        string
	Registration *registration.Resource
	PrivateKey   []byte
}

func (u *user) GetEmail() string {
	return u.Email
}

func (u *user) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *user) GetPrivateKey() crypto.PrivateKey {
	return ed25519.PrivateKey(u.PrivateKey)
}

func loadOrCreateUser(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string) (*user, error) {
	u, err := loadUser(ctx, sm, email)
	if isKeyNotFoundError(err) {
		return createUser(ctx, email)
	}

	return u, err
}

func secretsManagerUserName(email string) string {
	return "goacme/lego user for " + email
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

func createUser(ctx context.Context, email string) (*user, error) {
	_, key, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	return &user{
		Email:      email,
		PrivateKey: []byte(key),
	}, nil
}

func saveUser(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, u *user) error {
	bs, err := json.Marshal(u)
	if err != nil {
		return err
	}

	createSecretInput := &secretsmanager.CreateSecretInput{
		Description:  aws.String("goacme/lego user for: " + u.Email),
		Name:         aws.String(secretsManagerUserName(u.Email)),
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

func loadUser(ctx context.Context, sm secretsmanageriface.SecretsManagerAPI, email string) (*user, error) {
	getSecretInput := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretsManagerUserName(email)),
	}
	out, err := sm.GetSecretValueWithContext(ctx, getSecretInput)
	if err != nil {
		return nil, err
	}

	u := new(user)
	err = json.Unmarshal(out.SecretBinary, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}
