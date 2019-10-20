package main

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Config struct {
	session    *session.Session
	bucket     string
	expiryMins time.Duration
}

var s3Config *S3Config

type Secret struct {
	Key  string
	Data string
}

var port = "9090"

var usage = `
Usage examples:
# from stdin
- echo 'somesecret' | curl --data-binary @- https://<secrets-server-url>
# from file: aws_user_secrets
- curl --data-binary @aws_user_secrets https://<secrets-server-url>
`

func (s *Secret) keyExists(config *S3Config) (bool, error) {
	svc := s3.New(config.session)
	results, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(config.bucket), Prefix: aws.String(s.Key)})
	if err != nil {
		return false, err
	}

	return len(results.Contents) > 0, nil
}

func (s *Secret) storeSecret(config *S3Config) error {

	if exists, err := s.keyExists(config); err != nil {
		return err
	} else if exists {
		return nil
	}

	uploader := s3manager.NewUploader(config.session)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(config.bucket),
		Key:         aws.String(s.Key),
		ContentType: aws.String("text/plain"),
		Body:        strings.NewReader(s.Data),
	})

	if err != nil {
		return fmt.Errorf("failed to upload file %s to s3://%s", s.Key, config.bucket)
	}
	return nil
}

func (s *Secret) generateExpiryURL(config *S3Config) (string, error) {
	svc := s3.New(config.session)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(config.bucket),
		Key:    aws.String(s.Key),
	})

	url, _, err := req.PresignRequest(config.expiryMins * time.Minute)

	if err != nil {
		return "", fmt.Errorf("failed to sign request for obj s3://%s/%s", config.bucket, s.Key)
	}

	return url + "\n", nil
}

func extractDataFromRequest(r io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return buf.String()
}

func hash(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return strconv.FormatUint(h.Sum64(), 10)
}

func writeError(w io.Writer, msg string) {
	fmt.Fprintf(w, msg)
	fmt.Fprintf(w, usage)
}

func handler(w http.ResponseWriter, r *http.Request) {
	data := extractDataFromRequest(r.Body)

	if len(data) == 0 {
		writeError(w, "error: received empty request")
		return
	}

	var s = &Secret{
		Key:  hash(data) + ".txt",
		Data: data,
	}

	if err := s.storeSecret(s3Config); err != nil {
		writeError(w, "error: "+err.Error())
		return
	}

	url, err := s.generateExpiryURL(s3Config)
	if err != nil {
		writeError(w, "error: "+err.Error())
		return
	}

	w.Write([]byte(url))
}

func main() {
	// TODO add flags for --bucket --expiry-time-mins --port
	region, exists := os.LookupEnv("AWS_REGION")

	if !exists {
		log.Fatalf("missing env AWS_REGION")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		log.Fatalf("failed to get aws session", err)
	}

	s3Config = &S3Config{
		session:    sess,
		bucket:     "users-shared-secrets",
		expiryMins: time.Duration(30),
	}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
