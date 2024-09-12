package csv

/*
import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	zl "github.com/rs/zerolog"

	. "gopkg.in/check.v1"
)

const (
	Profile = "dev"
	Region  = "us-east-1"
	Bucket  = "stephen"
	Prefix  = "data"
)

type S3Suite struct {
	sess   *session.Session
	reader *S3Reader
}

var _ = Suite(&S3Suite{})

func (s *S3Suite) SetUpSuite(c *C) {
	s.sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           Profile,
		Config: aws.Config{
			Region: aws.String(Region),
		},
	}))

	logger := zl.New(zl.ConsoleWriter{Out: os.Stdout}).With().
		Timestamp().
		Logger()

	s.reader = NewS3Reader(s.sess, &logger)
}

func (s *S3Suite) Test_List(c *C) {

	files, err := s.reader.List(Bucket, Prefix)
	c.Assert(err, IsNil)

	fmt.Println("here")

	for file := range files {
		fmt.Println(file)
	}
}
*/
