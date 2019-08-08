# Set up Object Lifecycle Configuration [![Slack](https://slack.min.io/slack?type=svg)](https://slack.min.io) [![Go Report Card](https://goreportcard.com/badge/minio/minio)](https://goreportcard.com/report/minio/minio) [![Docker Pulls](https://img.shields.io/docker/pulls/minio/minio.svg?maxAge=604800)](https://hub.docker.com/r/minio/minio/)

Using object lifecycle configuration on buckets to setup automatic objects deletion after a specified number of days or a specified date.

## 1. Prerequisites
- Install MinIO - [MinIO Quickstart Guide](https://docs.min.io/docs/minio-quickstart-guide).
- Install AWS Cli - [Installing the AWS CLI - AWS Command Line Interface](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html)


## 2. Setup a bucket lifecycle using AWS

1. Create the following bucket lifecycle configuration expires the objects under the prefix `uploads/2015` on 2020-01-01T00:00:00.000Z since upload and the objects under `temporary-uploads/` after 7 days after upload, as shown below:

```sh
$ cat >bucket-lifecycle.json
cat bucket-lifecycle.json
{
    "Rules": [
        {
            "Expiration": {
                "Date": "2020-01-01T00:00:00.000Z"
            },
            "ID": "Delete very old messenger pictures",
            "Filter": {
                "Prefix": "uploads/2015/"
            },
            "Status": "Enabled"
        },
        {
            "Expiration": {
                "Days": 7
            },
            "ID": "Delete temporary uploads",
            "Filter": {
                "Prefix": "temporary-uploads/"
            },
            "Status": "Enabled"
        }
    ]
}
```

2. Upload the bucket lifecycle configuration using AWS cli:

```sh
$ export AWS_ACCESS_KEY_ID="your-access-key"
$ export AWS_SECRET_ACCESS_KEY="your-secret-key"
$ aws s3api put-bucket-lifecycle-configuration --bucket your-bucket --endpoint-url http://minio-server-address:port --lifecycle-configuration file://bucket-lifecycle.json
```

## Explore Further
- [Object Lifecycle Management](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lifecycle-mgmt.html)
