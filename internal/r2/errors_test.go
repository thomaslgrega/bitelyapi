package r2

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/thomaslgrega/bitelyapi/internal/models"
)

func TestAsStoreErrorNamesAMissingObject(t *testing.T) {
	missing := []error{
		&types.NotFound{},
		&types.NoSuchKey{},
		fmt.Errorf("operation error S3: HeadObject: %w", &types.NotFound{}),
		&awshttp.ResponseError{
			ResponseError: &smithyhttp.ResponseError{
				Response: &smithyhttp.Response{Response: &http.Response{StatusCode: http.StatusNotFound}},
				Err:      errors.New("not found"),
			},
		},
	}

	for _, err := range missing {
		if !errors.Is(asStoreError(err), models.ErrImageNotFound) {
			t.Fatalf("expected %v to read as a missing object", err)
		}
	}
}

func TestAsStoreErrorLeavesEverythingElseAlone(t *testing.T) {
	down := errors.New("dial tcp: connection refused")

	if !errors.Is(asStoreError(down), down) {
		t.Fatal("expected an unrelated failure to pass through unchanged")
	}
}
