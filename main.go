package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kelseyhightower/envconfig"
)

var conf config

func main() {
	err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func _main() error {
	err := parse()
	if err != nil {
		return err
	}

	sess, err := newAwsSession()
	if err != nil {
		return err
	}

	var iErr error
	iter := new(s3manager.DownloadObjectsIterator)
	svc := s3.New(sess)
	err = svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(conf.Bucket),
		Prefix: aws.String(conf.Prefix),
	}, func(output *s3.ListObjectsV2Output, last bool) bool {
		for _, object := range output.Contents {
			key := *object.Key
			path := filepath.Join(conf.Destination, strings.TrimPrefix(key, conf.Prefix))

			// NOTE: folder objects
			if *object.Size == 0 && strings.HasSuffix(key, "/") {
				continue
			}

			err := os.MkdirAll(filepath.Dir(path), 0o755)
			if err != nil {
				iErr = err
				return false
			}

			if conf.Debug {
				log.Printf("DEBUG: download s3://%s/%s -> %s\n", conf.Bucket, key, path)
			}

			f, err := os.Create(path)
			if err != nil {
				iErr = err
				return false
			}

			iter.Objects = append(iter.Objects, s3manager.BatchDownloadObject{
				Object: &s3.GetObjectInput{
					Bucket: aws.String(conf.Bucket),
					Key:    object.Key,
				},
				Writer: f,
				After: func() error {
					return f.Close()
				},
			})
		}
		return true
	})
	if err != nil {
		return err
	}
	if iErr != nil {
		return iErr
	}

	downloader := s3manager.NewDownloader(sess)
	err = downloader.DownloadWithIterator(context.Background(), iter)

	return err
}

func newAwsSession() (*session.Session, error) {
	c := aws.NewConfig()
	if conf.Debug {
		c.WithLogLevel(aws.LogDebugWithHTTPBody)
	}
	return session.NewSession(c)
}

func parse() error {
	err := envconfig.Process("fountain", &conf)
	if err != nil {
		return err
	}

	u, err := url.Parse(conf.Source)
	if err != nil {
		return err
	}

	if u.Scheme != "s3" && u.Host != "" && u.Path != "" {
		return fmt.Errorf("source must `s3://your_bucket/prefix/` format")
	}

	conf.Bucket = u.Host
	conf.Prefix = strings.TrimPrefix(u.Path, "/")
	if !strings.HasSuffix(conf.Prefix, "/") {
		conf.Prefix = conf.Prefix + "/"
	}

	if conf.Debug {
		log.Printf("DEBUG: config %+v\n", conf)
	}

	return nil
}
