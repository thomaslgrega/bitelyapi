package r2

import (
	"errors"
	"net/http"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/thomaslgrega/bitelyapi/internal/models"
)

// asStoreError separates a key that names no object from a bucket that could
// not answer. Without the split every failure would read as the client's stale
// claim ticket, and R2 being down would be reported as its fault.
//
// The 404 case is here because S3 models a missing object for HeadObject as a
// bare status: there is no body to decode a typed error out of.
func asStoreError(err error) error {
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return models.ErrImageNotFound
	}

	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return models.ErrImageNotFound
	}

	var response *awshttp.ResponseError
	if errors.As(err, &response) && response.HTTPStatusCode() == http.StatusNotFound {
		return models.ErrImageNotFound
	}

	return err
}
