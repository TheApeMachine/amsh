package datalake

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"
	"github.com/theapemachine/errnie"
)

// Conn implements the io.ReadWriteCloser interface such that datalake (S3)
// operations fit transparently in the common Go data transfer methods.
type Conn struct {
	bucket string
	client *s3.Client
	key    string
	wg     *sync.WaitGroup
}

func getClient() *s3.Client {
	return s3.NewFromConfig(aws.Config{
		Region: "weur",
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
		EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:       "aws",
				URL:               os.Getenv("AWS_ENDPOINT"),
				SigningRegion:     region,
				HostnameImmutable: true,
			}, nil
		}),
	})
}

var (
	downloaderPool = sync.Pool{
		New: func() interface{} {
			return manager.NewDownloader(getClient())
		},
	}
	uploaderPool = sync.Pool{
		New: func() interface{} {
			return manager.NewUploader(getClient())
		},
	}
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

func NewConn(prefix string) *Conn {
	return &Conn{
		bucket: viper.GetViper().GetString("datalake.bucket"), // Add bucket parameter to initialize the bucket
		client: getClient(),                                   // Initialize the S3 client here
		key:    prefix,
	}
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	var page *s3.ListObjectsV2Output

	downloader := downloaderPool.Get().(*manager.Downloader)
	defer downloaderPool.Put(downloader)
	downloader.Concurrency = 10
	downloader.PartSize = 1024 * 1024 * 64

	paginator := s3.NewListObjectsV2Paginator(conn.client, &s3.ListObjectsV2Input{
		Bucket: aws.String("datalake"),
		Prefix: aws.String(conn.key),
	})

	totalBytesRead := 0
	buf := bytes.NewBuffer(p[:0]) // Buffer to write data into (initially empty)

	for paginator.HasMorePages() {
		if page, err = paginator.NextPage(context.TODO()); err != nil {
			return 0, err
		}

		for _, obj := range page.Contents {
			// Get the S3 object
			s3Object, err := conn.client.GetObject(context.TODO(), &s3.GetObjectInput{
				Bucket: aws.String(conn.bucket),
				Key:    aws.String(*obj.Key),
			})
			if err != nil {
				return 0, fmt.Errorf("failed to get object: %w", err)
			}
			defer s3Object.Body.Close() // Ensure we close the object stream

			// Use io.Copy to copy data from the S3 object stream (Reader) to our buffer (Writer)
			bytesCopied, err := io.Copy(buf, s3Object.Body)
			if err != nil {
				return int(totalBytesRead), fmt.Errorf("failed to copy data: %w", err)
			}

			totalBytesRead += int(bytesCopied)

			// Check if `p` is large enough to hold the copied data
			if totalBytesRead > len(p) {
				return totalBytesRead, fmt.Errorf("buffer size exceeded")
			}
		}
	}

	// Copy the data from the buffer `buf` to `p`
	copy(p, buf.Bytes())

	// print p

	return totalBytesRead, nil
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	uploader := uploaderPool.Get().(*manager.Uploader)
	defer uploaderPool.Put(uploader)
	defer conn.wg.Done()

	// Perform the S3 upload
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("datalake"),
		Key:    aws.String(conn.key),
		Body:   bytes.NewReader([]byte(string(p))), // Upload the provided byte slice
	})

	if err != nil {
		return 0, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return the number of bytes written
	return len(p), nil
}

// ListFiles lists all files under the current key as a prefix recursively
func (conn *Conn) ListFiles() []byte {
	var (
		page *s3.ListObjectsV2Output
		keys = bufferPool.Get().(*bytes.Buffer)
		err  error
	)

	defer bufferPool.Put(keys)
	keys.Reset()
	keys.WriteString("[")

	paginator := s3.NewListObjectsV2Paginator(conn.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(conn.bucket),
		Prefix: aws.String(conn.key),
	})

	for paginator.HasMorePages() {
		if page, err = paginator.NextPage(context.TODO()); err != nil {
			errnie.Error(err)
			continue
		}

		for _, obj := range page.Contents {
			keys.WriteString(`{"prefix": "`)
			keys.WriteString(*obj.Key)
			keys.WriteString(`"},`)
		}
	}

	if keys.Len() > 1 {
		keys.Truncate(keys.Len() - 1) // Remove trailing comma
	}
	keys.WriteString("]")

	return keys.Bytes()
}

func (conn *Conn) SetKey(key string) {
	conn.key = key
}

func (conn *Conn) Close() error {
	return nil
}
