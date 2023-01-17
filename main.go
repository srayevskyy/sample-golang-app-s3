package main

import (
	"log"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type AppConfigType struct {
	S3Config S3ConfigType `mapstructure:"S3_CONFIG" validate:"required,dive,required"`
}

type S3ConfigType struct {
	AccessKeyId     string `mapstructure:"ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"SECRET_ACCESS_KEY"`
	Region          string `mapstructure:"REGION" validate:"required"`
	BucketName      string `mapstructure:"BUCKET_NAME" validate:"required"`
	PathPrefix      string `mapstructure:"PATH_PREFIX" validate:"required"`
}

var (
	StructValidator = validator.New()
	GitCommitHash   = "unknown"
)

func (s3Config S3ConfigType) UploadFileToS3(
	localTmpFileDir string,
	localTmpFileName string,
) (TgtS3KeyName string, err error) {
	var (
		localTmpFilePath  string = path.Join(localTmpFileDir, localTmpFileName)
		file              *os.File
		TgtS3SessUpload   *session.Session
		TgtS3UploadResult *s3manager.UploadOutput
		TgtAWSConfig      = aws.Config{
			Region:                        &s3Config.Region,
			CredentialsChainVerboseErrors: aws.Bool(true),
		}
	)

	if s3Config.AccessKeyId != "" && s3Config.SecretAccessKey != "" {
		log.Printf("AWS access key and secret key have been provided, connecting with static credentials")
		TgtAWSConfig.Credentials = credentials.NewStaticCredentials(
			s3Config.AccessKeyId,
			s3Config.SecretAccessKey,
			"",
		)
	} else {
		log.Printf("AWS access key and/or secret key have not been provided, connecting with IAM role/dynamic credentials")
	}

	if TgtS3SessUpload, err = session.NewSession(&TgtAWSConfig); err != nil {
		return TgtS3KeyName, err
	}

	TgtS3KeyName = path.Join("/", s3Config.PathPrefix, localTmpFileName)
	if file, err = os.Open(localTmpFilePath); err != nil {
		return TgtS3KeyName, err
	}
	defer file.Close()

	TgtS3Uploader := s3manager.NewUploader(TgtS3SessUpload)

	if TgtS3UploadResult, err = TgtS3Uploader.Upload(&s3manager.UploadInput{
		Body:   file,
		Bucket: aws.String(s3Config.BucketName),
		Key:    aws.String(TgtS3KeyName),
	}); err != nil {
		return TgtS3KeyName, err
	}
	if err = os.Remove(localTmpFilePath); err != nil {
		return TgtS3KeyName, err
	}
	log.Println("Successfully uploaded to", TgtS3UploadResult.Location)

	return TgtS3KeyName, nil
}

func (appConfig *AppConfigType) Run() (err error) {
	var (
		tmpFilePath = "/tmp"
		tmpFileName = "out.txt"
	)
	if err = os.WriteFile(path.Join(tmpFilePath, tmpFileName), []byte("hello\ngo\n"), 0644); err != nil {
		return err
	}
	if _, err = appConfig.S3Config.UploadFileToS3(tmpFilePath, tmpFileName); err != nil {
		return err
	}
	return err
}

func (appConfig *AppConfigType) ReadConfig() (err error) {
	v := viper.NewWithOptions(viper.KeyDelimiter("__"))
	v.AddConfigPath("/config")
	v.AddConfigPath("config")
	v.SetConfigName("app")
	v.SetConfigType("env")

	v.AutomaticEnv()

	if err = v.ReadInConfig(); err != nil {
		return err
	}
	if err = v.Unmarshal(&appConfig); err != nil {
		// log.Fatalf("Unable to decode viper config into app config struct, %v", err)
		return err
	}

	if err = StructValidator.Struct(appConfig); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		// TODO: collapse validationErrors into string, return as 'err'
		log.Println(validationErrors)
	}

	return err
}

func main() {
	var (
		err       error
		appConfig AppConfigType
	)
	log.Printf("Pipeline source code git commit hash: %s", GitCommitHash)
	if err = appConfig.ReadConfig(); err != nil {
		log.Fatalf("Error reading app config: %s", err.Error())
	}
	if err = appConfig.Run(); err != nil {
		log.Fatalf(err.Error())
	}
}
