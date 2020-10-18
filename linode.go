package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

func uploadCertToLinode(ctx context.Context, linodeToken, mapDomain string, cert, key []byte) error {
	linodeClient := configureLinodeClient(linodeToken)
	linodeClusterId, err := inferLinodeClusterId(ctx, linodeClient, mapDomain)
	if err != nil {
		return err
	}
	response, err := linodeClient.UploadObjectStorageBucketCert(ctx, linodeClusterId, mapDomain, linodego.ObjectStorageBucketCertUploadOptions{
		PrivateKey:  string(cert),
		Certificate: string(key),
	})
	if err != nil {
		return err
	}
	if !response.SSL {
		return errors.New("couldnt upload ssl or something idk linode bro what is this")
	}
	return nil
}

type LinodeClient interface {
	ListObjectStorageBuckets(ctx context.Context, opts *linodego.ListOptions) ([]linodego.ObjectStorageBucket, error)
	UploadObjectStorageBucketCert(ctx context.Context, clusterID, bucket string, uploadOpts linodego.ObjectStorageBucketCertUploadOptions) (*linodego.ObjectStorageBucketCert, error)
}

func configureLinodeClient(token string) LinodeClient {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	client := linodego.NewClient(oauth2Client)
	return &client
}

func inferLinodeClusterId(ctx context.Context, linode LinodeClient, domain string) (clusterID string, err error) {
	buckets, err := linode.ListObjectStorageBuckets(ctx, nil)
	if err != nil {
		return "", err
	}

	for _, bucket := range buckets {
		if bucket.Label == domain {
			return bucket.Cluster, nil
		}
	}
	return "", nil
}
