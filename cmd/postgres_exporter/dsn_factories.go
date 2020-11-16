package main

import (
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"
)

// DsnFactory creates DSNs
type DsnFactory interface {
	// Creates a new DSN
	NewDsn() (string, error)

	// Return a stable, valid, & parsable DSN value for logging, map keys, etc.
	DsnID() string
}

// RawDsnFactory is a no-op factory that returns the original 'raw' DSN.
type RawDsnFactory struct {
	dsn string
}

// DsnID returns the original DSN
func (f *RawDsnFactory) DsnID() string {
	return f.dsn
}

// NewDsn returns the original DSN. No mutations are required DSNs created by the RawDsnFactory
func (f *RawDsnFactory) NewDsn() (string, error) {
	return f.dsn, nil
}

// AwsRdsIamFactory generates a DSN configured with AWS RDS IAM authentication credentials
type AwsRdsIamFactory struct {
	dsn string
}

// DsnID returns the original DSN. The password is _not_ replaced with an AWS RDS IAM generated password token
func (f *AwsRdsIamFactory) DsnID() string {
	return f.dsn
}

// NewDsn builds a new DSN with the password set to one obtained from the AWS RDS IAM token API
func (f *AwsRdsIamFactory) NewDsn() (string, error) {
	var err error
	var sess *session.Session
	sess, err = session.NewSession()
	if err != nil {
		return "", err
	}

	var u *url.URL
	u, err = url.Parse(f.dsn)
	if err != nil {
		return "", err
	}

	region := *sess.Config.Region

	if len(region) == 0 {
		if len(os.Getenv("AWS_REGION")) > 0 {
			region = os.Getenv("AWS_REGION")
		} else {
			var r string
			r, err = ec2metadata.New(sess).Region()

			if err != nil {
				return "", err
			}
			region = r
		}
	}

	var token string
	token, err = rdsutils.BuildAuthToken(u.Host, region, u.User.Username(), sess.Config.Credentials)
	if err != nil {
		return "", err
	}

	u.User = url.UserPassword(u.User.Username(), token)

	return u.String(), nil
}
