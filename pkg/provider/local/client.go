package local

import (
	"fmt"
	"os"

	"github.com/alec-rabold/zipspy/pkg/zipspy"
)

var _ zipspy.Reader = (*Client)(nil)

// Client implements the <<SOMETHING>> interface.
type Client struct {
	filePath string
}

// NewClient creates a new local file reader.
func NewClient(filePath string) (zipspy.Reader, error) {
	return &Client{filePath: filePath}, nil
}

// GetContentLength returns the size of and object's response body in bytes.
func (c *Client) Size() (int64, error) {
	file, err := os.Stat(c.filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file (name: %s): %w", c.filePath, err)
	}
	return file.Size(), nil
}

// ReadAt implements the io.ReaderAt interface by downloading a byte range of the object.
func (c *Client) ReadAt(p []byte, off int64) (n int, err error) {
	file, err := os.Open(c.filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file (path: %s): %w", c.filePath, err)
	}
	defer file.Close()
	return file.ReadAt(p, off)
}
