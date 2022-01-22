package zipspy

import (
	"fmt"
	"io"
	"regexp"

	"github.com/alec-rabold/zipspy/pkg/reader"
)

// Reader includes the necessary methods for a zip file.
type Reader interface {
	io.ReaderAt
	Size() (int64, error)
}

// Client is a zipspy client.
type Client struct {
	r *reader.Reader
}

// NewClient creates a new top-level zipspy client.
func NewClient(p Reader) (*Client, error) {
	size, err := p.Size()
	if err != nil {
		return nil, fmt.Errorf("failed to get size: %w", err)
	}
	zr, err := reader.NewReader(p, size)
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}
	return &Client{r: zr}, nil
}

// AllFiles returns a list of all files in the archive.
func (c *Client) AllFiles() []*reader.File {
	return c.r.File
}

// SearchFiles returns all files that match the given regex.
func (c *Client) GetFiles(searchFiles []string) []*reader.File {
	var matches []*reader.File
	for _, s := range searchFiles {
		for _, f := range c.r.File {
			if f.Name == s {
				matches = append(matches, f)
			}
		}
	}
	return matches
}

// SearchFiles returns all files that match the given regex.
func (c *Client) SearchFiles(re regexp.Regexp) []*reader.File {
	var matches []*reader.File
	for _, file := range c.r.File {
		if len(re.FindStringSubmatch(file.Name)) > 0 {
			matches = append(matches, file)
		}
	}
	return matches
}
