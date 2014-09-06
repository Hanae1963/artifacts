package upload

import (
	"testing"
	"time"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/mitchellh/goamz/s3/s3test"
	"github.com/travis-ci/artifacts/artifact"
)

var (
	s3srv = &localS3Server{
		config: &s3test.Config{
			Send409Conflict: true,
		},
	}
	testS3 *s3.S3
)

func init() {
	s3srv.SetUp()
	testS3 = s3.New(s3srv.Auth, s3srv.Region)
}

type localS3Server struct {
	Auth   aws.Auth
	Region aws.Region
	srv    *s3test.Server
	config *s3test.Config
}

func (s *localS3Server) SetUp() {
	if s.srv != nil {
		return
	}

	srv, err := s3test.NewServer(s.config)
	if err != nil {
		panic(err)
	}

	s.srv = srv
	s.Region = aws.Region{
		Name:                 "faux-region-9000",
		S3Endpoint:           srv.URL(),
		S3LocationConstraint: true,
	}
}

func TestNewS3Provider(t *testing.T) {
	s3p := newS3Provider(NewOptions(), getPanicLogger())

	if s3p.RetryInterval != (3 * time.Second) {
		t.Fatalf("wrong default retry interval")
	}

	if s3p.Name() != "s3" {
		t.Fatalf("wrong name")
	}
}

func TestS3ProviderUpload(t *testing.T) {
	opts := NewOptions()
	s3p := newS3Provider(opts, getPanicLogger())
	s3p.overrideConn = testS3
	s3p.overrideAuth = aws.Auth{
		AccessKey: "whatever",
		SecretKey: "whatever",
		Token:     "whatever",
	}

	in := make(chan *artifact.Artifact)
	out := make(chan *artifact.Artifact)
	done := make(chan bool)

	go s3p.Upload("test-0", opts, in, out, done)

	go func() {
		in <- &artifact.Artifact{}
		close(in)
	}()

	accum := []*artifact.Artifact{}
	for {
		select {
		case <-time.After(5 * time.Second):
			t.Fatalf("took too long oh derp")
		case a := <-out:
			accum = append(accum, a)
		case <-done:
			if len(accum) == 0 {
				t.Fatalf("nothing uploaded")
			}
			return
		}
	}
}
