package s3

import (
	"fmt"
	"io/ioutil"

	"github.com/alec-rabold/zipspy/pkg/zipspy"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// TODO: add a cache plugin?

var _ zipspy.Reader = (*Client)(nil)

// Client implements the <<SOMETHING>> interface.
type Client struct {
	bucket string
	key    string
	s3     S3API
}

// S3API contains the S3 API endpoints we care about in this package.
type S3API interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	HeadObject(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error)
}

// NewClient creates a new AWS S3 file reader.
func NewClient(bucket, key string) *Client {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &Client{
		bucket: bucket,
		key:    key,
		s3:     s3.New(sess),
	}
}

// Size returns the size of the object.
func (c *Client) Size() (int64, error) {
	output, err := c.s3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(c.key),
	})
	if err != nil {
		return 0, fmt.Errorf("failed getting head object (bucket: %s) (key: %s): %w", c.bucket, c.key, err)
	}
	return *output.ContentLength, nil
}

// ReadAt implements the io.ReaderAt interface by downloading a byte range of the object.
func (c *Client) ReadAt(p []byte, off int64) (n int, err error) {
	byteRange := fmt.Sprintf("bytes=%v-%v", off, off+int64(len(p)-1))
	output, err := c.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(c.key),
		Range:  aws.String(byteRange),
	})
	if err != nil {
		return 0, fmt.Errorf("failed getting object (bucket: %s) (key: %s) (range: %s): %w", c.bucket, c.key, byteRange, err)
	}
	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}
	return copy(p, body), nil
}
