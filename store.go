package daybook

import "fmt"
import "os"
import "strings"
import "github.com/mitchellh/goamz/aws"
import "github.com/mitchellh/goamz/s3"

type Store interface {
	GetAll(serviceName string) ([]*Service, error)
	Get(service *Service) (Archive, error)
	Put(service *Service, filePath string) error
}

type S3Store struct {
	client *s3.S3
	bucket string
}

func NewS3Store(auth aws.Auth, region aws.Region, bucket string) *S3Store {
	return &S3Store{
		client: s3.New(auth, region),
		bucket: bucket,
	}
}

func (s *S3Store) GetAll(serviceName string) ([]*Service, error) {
	services := []*Service{}

	b := s.client.Bucket(s.bucket)
	resp, err := b.List(serviceName, "", "", 0)
	if err != nil {
		return nil, err
	}

	for _, k := range resp.Contents {
		cleaned := strings.TrimSuffix(k.Key, ".tar.gz")
		parts := strings.Split(cleaned, "-")

		services = append(services, &Service{Name: serviceName, Version: parts[1]})
	}

	return services, nil
}

func (s *S3Store) Get(service *Service) (Archive, error) {
	b := s.client.Bucket(s.bucket)

	rc, err := b.GetReader(fmt.Sprintf("%s-%s.tar.gz", service.Name, service.Version))
	if err != nil {
		s3err, ok := err.(*s3.Error)
		if ok {
			if s3err.Code == "NoSuchKey" {
				return nil, fmt.Errorf("couldn't find asset for %s/%s", service.Name, service.Version)
			}
		}

		return nil, err
	}

	return NewTarGzArchive(rc), nil
}

func (s *S3Store) Put(service *Service, filePath string) error {
	if !strings.HasSuffix(filePath, ".tar.gz") {
		return fmt.Errorf("only .tar.gz archives are supported with this store")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	b := s.client.Bucket(s.bucket)
	return b.PutReader(fmt.Sprintf("%s-%s.tar.gz", service.Name, service.Version), f, fi.Size(), "application/x-gtar", s3.Private)
}
