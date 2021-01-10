# acme-lambda

## How it works

* TBD something that manages what domains we want to manage and uploads to s3

  * maybe cli that invokes acme with new domain and removes unwanted domains from s3?

* when renewing certificates, acme should find domains it needs to renew by listing all domains in s3

* acme should check the expiry / revocation status of these certificates. If they are revoked, renew (?). If they have expired already or will expire within a configured time duration, then renew.

* obtaining a cert:

  * generate ecdsa private key, store in s3, encrypted with kms

  * generate a `*x509.CertificateRequest`, store alongside the private key

  * initiate DNS-01 challenge using route53

  * send message with delay timer, containing acme order information to poll for authorization, the path to the csr and private key, and include time for backoff

    * trigger acme-verify lambda which will call the authorization endpoint for status on the cert request

    * when authorization is confirmed, call `CreateOrderCert` with the CSR

    * store cert alongside the privatekey
