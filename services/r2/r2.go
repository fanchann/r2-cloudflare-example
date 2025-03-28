package r2

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"
	"github.com/zakirkun/dy"

	"r2_example/dto"
	"r2_example/utils"
)

type IR2Services interface {
	UploadSingleFile(ctx context.Context, file io.Reader, fileName string) (*dto.FileSuccessUploadResponse, error)
	GetListFile(ctx context.Context) (*dto.FileListsResponse, error)
	GenSignedURL(ctx context.Context, fileName string, duration time.Duration) (*dto.MakeFilePublicResponse, error)
	GetFileByKey(ctx context.Context, key string) (*dto.R2ListFile, error)

	init() (*s3.Client, error)
}

type r2Services struct {
	v   *viper.Viper
	log *dy.Logger
}

func NewR2Services(v *viper.Viper, l *dy.Logger) IR2Services {
	return &r2Services{v: v, log: l}
}

func (r2 *r2Services) init() (*s3.Client, error) {
	account := r2.v.GetString("R2_ACCOUNT_ID")
	accessKey := r2.v.GetString("R2_ACCESS_KEY_ID")
	secretKey := r2.v.GetString("R2_SECRET_ACCESS_KEY")

	// Custom resolver
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", account),
		}, nil
	})

	// Load AWS config with custom resolver
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	// Create a new S3 client
	s3Client := s3.NewFromConfig(cfg)

	return s3Client, nil
}

func (r2 *r2Services) UploadSingleFile(ctx context.Context, file io.Reader, fileName string) (*dto.FileSuccessUploadResponse, error) {
	s3Client, err := r2.init()
	if err != nil {
		r2.log.Debug(err.Error())
		return nil, fmt.Errorf("failed to initialize R2 client: %w", err)
	}

	bucketName := r2.v.GetString("R2_BUCKET_NAME")

	var contentType string
	var contentLength int64
	var body io.ReadSeeker

	if seeker, ok := file.(io.ReadSeeker); ok {
		contentType, err = getContentType(seeker)
		if err != nil {
			r2.log.Debug(err.Error())
			return nil, fmt.Errorf("failed to get content type: %w", err)
		}

		// count file size
		currentPos, _ := seeker.Seek(0, io.SeekCurrent)
		endPos, err := seeker.Seek(0, io.SeekEnd)
		if err != nil {
			r2.log.Debug(err.Error())
			return nil, fmt.Errorf("failed to determine file size: %w", err)
		}
		_, _ = seeker.Seek(currentPos, io.SeekStart) // reset
		contentLength = endPos - currentPos

		body = seeker
	} else {
		// if file can't read, read all
		data, err := io.ReadAll(file)
		if err != nil {
			r2.log.Debug(err.Error())
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		contentLength = int64(len(data))
		contentType = http.DetectContentType(data)
		body = bytes.NewReader(data)
	}

	if contentLength == 0 {
		r2.log.Debug(err.Error())
		return nil, fmt.Errorf("file is empty, upload aborted")
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(fileName),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	}

	if _, err := s3Client.PutObject(ctx, input); err != nil {
		r2.log.Debug(err.Error())
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	r2.log.Info("file uploaded successfully: %s", fileName)
	return &dto.FileSuccessUploadResponse{
		FileName: fileName,
		FileSize: contentLength,
	}, nil
}

func (r2 *r2Services) GetListFile(ctx context.Context) (*dto.FileListsResponse, error) {
	s3Client, err := r2.init()
	if err != nil {
		r2.log.Debug(err.Error())
		return nil, fmt.Errorf("failed to initialize R2 client: %v", err)
	}

	bucketName := r2.v.GetString("R2_BUCKET_NAME")

	output, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	if err != nil {
		r2.log.Debug(err.Error())
		return nil, fmt.Errorf("failed to list objects: %v", err)
	}

	filesResponses := utils.R2TypesObjectToListsFileDto(output.Contents)

	return &dto.FileListsResponse{Files: filesResponses}, nil
}

func (r2 *r2Services) GetFileByKey(ctx context.Context, key string) (*dto.R2ListFile, error) {
	s3Client, err := r2.init()
	if err != nil {
		r2.log.Debug(err.Error())
		return nil, fmt.Errorf("failed to initialize R2 client: %w", err)
	}

	bucketName := r2.v.GetString("R2_BUCKET_NAME")

	output, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	if err != nil {
		r2.log.Debug(err.Error())
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	for _, content := range output.Contents {
		if key == *content.Key {
			return &dto.R2ListFile{
				FileName: *content.Key,
				Size:     *content.Size,
			}, nil
		}
	}

	errMsg := fmt.Sprintf("file with key '%s' not found", key)
	r2.log.Debug(errMsg)
	return nil, fmt.Errorf(errMsg)
}

func (r2 *r2Services) GenSignedURL(ctx context.Context, fileName string, duration time.Duration) (*dto.MakeFilePublicResponse, error) {
	s3Client, err := r2.init()
	if err != nil {
		r2.log.Debug(err.Error())
		return nil, fmt.Errorf("failed to initialize R2 client: %v", err)
	}

	bucketName := r2.v.GetString("R2_BUCKET_NAME")

	presignClient := s3.NewPresignClient(s3Client)

	presignResult, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(fileName),
		},
		s3.WithPresignExpires(duration),
	)

	if err != nil {
		r2.log.Debug(err.Error())
	}

	return &dto.MakeFilePublicResponse{URL: presignResult.URL, Duration: duration.String()}, nil

}

func getContentType(seeker io.ReadSeeker) (string, error) {
	buff := make([]byte, 512)

	startPos, err := seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return "", fmt.Errorf("failed to get current position: %w", err)
	}

	_, err = seeker.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("failed to seek start: %w", err)
	}

	bytesRead, err := seeker.Read(buff)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file header: %w", err)
	}

	_, err = seeker.Seek(startPos, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("failed to reset position: %w", err)
	}

	return http.DetectContentType(buff[:bytesRead]), nil
}
