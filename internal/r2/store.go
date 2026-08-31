// Package r2 stores Recipe Images in a Cloudflare R2 bucket. It signs uploads
// and moves objects; the bytes go straight between the client and R2
// (ADR-0006).
package r2

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thomaslgrega/bitelyapi/internal/models"
)

// uploadTTL is how long a presigned URL lives. It is short because the URL is
// reusable for the whole of its life: whoever holds it can write that object
// again until it expires.
const uploadTTL = 5 * time.Minute

type Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

type Store struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

// ConfigFromEnv reads the bucket's credentials, refusing anything missing.
// Booting without them would give a server that answers every upload with a
// 500 and every share with a rejected image.
func ConfigFromEnv() (Config, error) {
	cfg := Config{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		Bucket:          os.Getenv("R2_BUCKET"),
	}

	missing := make([]string, 0, 4)
	for name, value := range map[string]string{
		"R2_ACCOUNT_ID":        cfg.AccountID,
		"R2_ACCESS_KEY_ID":     cfg.AccessKeyID,
		"R2_SECRET_ACCESS_KEY": cfg.SecretAccessKey,
		"R2_BUCKET":            cfg.Bucket,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing R2 configuration: %v", missing)
	}

	return cfg, nil
}

func NewStore(cfg Config) (*Store, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		// R2 is single-region and wants the literal "auto"; SigV4 still needs
		// a region to sign with.
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://" + cfg.AccountID + ".r2.cloudflarestorage.com")
		o.UsePathStyle = true

		// R2 answered a checksum-carrying request with 501 for a while. The
		// server side appears fixed, but nothing here needs a checksum, and
		// asking for none costs nothing (ADR-0006).
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	return &Store{client: client, presign: s3.NewPresignClient(client), bucket: cfg.Bucket}, nil
}

// PresignUpload signs one PUT into the staging prefix. Both the declared type
// and the declared length are signed, so the upload R2 accepts is the one the
// client described. No checksum is configured: signing one covers an empty
// body and every real upload fails (ADR-0006).
func (s *Store) PresignUpload(ctx context.Context, contentType string, contentLength int64) (models.PresignedUpload, error) {
	key, err := models.NewStagedImageKey()
	if err != nil {
		return models.PresignedUpload{}, err
	}

	signedAt := time.Now()
	request, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	}, s3.WithPresignExpires(uploadTTL))
	if err != nil {
		return models.PresignedUpload{}, err
	}

	return models.PresignedUpload{
		UploadURL: request.URL,
		Key:       key,
		ExpiresAt: signedAt.Add(uploadTTL).UTC(),
	}, nil
}

// Head reports what an object turned out to be, which is the only place the
// size and type of an upload can be checked.
func (s *Store) Head(ctx context.Context, key string) (models.StagedImage, error) {
	head, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return models.StagedImage{}, err
	}

	staged := models.StagedImage{}
	if head.ContentType != nil {
		staged.ContentType = *head.ContentType
	}
	if head.ContentLength != nil {
		staged.ContentLength = *head.ContentLength
	}

	return staged, nil
}

// Promote copies a staged object to the key the server derived for it. The
// copy is server-side: the bytes never reach this process.
func (s *Store) Promote(ctx context.Context, stagedKey string, key string) error {
	if stagedKey == "" || key == "" {
		return errors.New("promote needs both a source and a destination key")
	}

	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		// CopySource is a `bucket/key` path that must be URL-encoded. Both keys
		// are server-shaped — a prefix, a UUID, a fixed filename — so encoding
		// them changes nothing.
		CopySource: aws.String(s.bucket + "/" + stagedKey),
	})

	return err
}

func (s *Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	return err
}
