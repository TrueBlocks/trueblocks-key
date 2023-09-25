package main

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type LambdaRunEnv struct{}

func (c *LambdaRunEnv) Blooms(chain string) (map[string]string, error) {
	return map[string]string{
		"004039898-004053179": "004039898-004053179.bloom",
		"005911512-005917403": "005911512-005917403.bloom",
		"012145996-012148690": "012145996-012148690.bloom",
		"012159745-012162467": "012159745-012162467.bloom",
		// "012548760-012551586": "012548760-012551586.bloom",
		// "012770525-012773128": "012770525-012773128.bloom",
		// "012829036-012831489": "012829036-012831489.bloom",
		// "012925312-012927827": "012925312-012927827.bloom",
		// "012983024-012985641": "012983024-012985641.bloom",
		// "013061630-013064215": "013061630-013064215.bloom",
		// "013069535-013072315": "013069535-013072315.bloom",
		// "013085505-013088077": "013085505-013088077.bloom",
		// "013485902-013488560": "013485902-013488560.bloom",
		// "013570582-013573316": "013570582-013573316.bloom",
		// "013575972-013578610": "013575972-013578610.bloom",
		// "013702893-013705896": "013702893-013705896.bloom",
		// "013705897-013708637": "013705897-013708637.bloom",
		// "013716714-013719729": "013716714-013719729.bloom",
		// "013727881-013730562": "013727881-013730562.bloom",
		// "013757905-013760669": "013757905-013760669.bloom",
		// "013777100-013779944": "013777100-013779944.bloom",
		// "013794793-013797575": "013794793-013797575.bloom",
		// "013820278-013823184": "013820278-013823184.bloom",
		// "013823185-013825978": "013823185-013825978.bloom",
		// "013834857-013837833": "013834857-013837833.bloom",
		// "014028932-014031737": "014028932-014031737.bloom",
		// "014031738-014034567": "014031738-014034567.bloom",
		// "014034568-014037390": "014034568-014037390.bloom",
		// "014091407-014094384": "014091407-014094384.bloom",
		// "014216863-014219639": "014216863-014219639.bloom",
		// "014246559-014249311": "014246559-014249311.bloom",
		// "014375257-014378295": "014375257-014378295.bloom",
		// "014381182-014384105": "014381182-014384105.bloom",
		// "014435826-014438546": "014435826-014438546.bloom",
		// "014438547-014441376": "014438547-014441376.bloom",
		// "014441377-014444006": "014441377-014444006.bloom",
		// "014469359-014472078": "014469359-014472078.bloom",
	}, nil
}

func (c *LambdaRunEnv) ReadBloom(fileName string) (io.ReadSeekCloser, error) {
	return c.readObject("blooms/" + fileName)
}

func (c *LambdaRunEnv) ReadChunk(chain string, blockRange string) (io.ReadSeekCloser, error) {
	return c.readObject("finalized/" + blockRange + ".bin")
}

func (c *LambdaRunEnv) readObject(filePath string) (io.ReadSeekCloser, error) {
	bucket := aws.String(BUCKET_NAME)
	key := aws.String(filePath)
	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		return nil, fmt.Errorf("get %s/%s: %w", *bucket, *key, err)
	}

	obj, err := NewS3Object(result)
	return obj, err
}
