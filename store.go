package daybook

import "io"
import "fmt"
import "os"
import "strings"
import "github.com/mitchellh/goamz/aws"
import "github.com/mitchellh/goamz/s3"

var SUPPORTED_EXTENSIONS map[string]string = map[string]string{
	".tar":    "application/x-tar",
	".tar.gz": "application/x-gtar",
	".zip":    "application/zip",
	".xz":     "application/x-xz",
}

type Store interface {
	Get(serviceName, version string) (io.ReadCloser, error)
	Put(serviceName, version, filePath string) error
}

type S3Store struct {
	client *s3.S3
	bucket string
}

func NewS3Store(auth aws.Auth, region aws.Region, bucket string) *S3Store {
	return &S3Store{
		client: s3.New(auth, region),
	}
}

func (s *S3Store) Get(serviceName, version string) (io.ReadCloser, error) {
	b := s.client.Bucket(s.bucket)

	for e, _ := range SUPPORTED_EXTENSIONS {
		rc, err := b.GetReader(fmt.Sprintf("%s-%s.%s", serviceName, version, e))
		if err != nil {
			s3err, ok := err.(*s3.Error)
			if ok {
				if s3err.Code == "NoSuchKey" {
					// skip to the next file extenson
					continue
				}
			}

			return nil, err
		}

		return rc, nil
	}

	return nil, fmt.Errorf("couldn't find object for %s-%s with known extensions", serviceName, version)
}

func (s *S3Store) Put(serviceName, version, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	extension, contentType, err := getFileInfo(filePath)
	if err != nil {
		return err
	}

	b := s.client.Bucket(s.bucket)
	return b.PutReader(fmt.Sprintf("%s-%s.%s", serviceName, version, extension), f, fi.Size(), contentType, s3.Private)
}

func getFileInfo(filePath string) (string, string, error) {
	for e, c := range SUPPORTED_EXTENSIONS {
		if strings.HasSuffix(filePath, e) {
			return e, c, nil
		}
	}

	return "", "", fmt.Errorf("unsupported file type (%s)", filePath)
}
