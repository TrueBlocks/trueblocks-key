package main

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Object struct {
	s3Output *s3.GetObjectOutput
	body     []byte
	*bytes.Reader
}

func NewS3Object(s3Output *s3.GetObjectOutput) (o *S3Object, err error) {
	body, err := io.ReadAll(s3Output.Body)
	if err != nil {
		return
	}

	o = &S3Object{
		s3Output: s3Output,
		body:     body,
		Reader:   bytes.NewReader(body),
	}
	return
}

func (o *S3Object) Close() error {
	return o.s3Output.Body.Close()
}
