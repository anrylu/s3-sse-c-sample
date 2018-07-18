package config

import (
	"flag"

	"github.com/anrylu/s3-sse-c-sample/pkg/service"
)

func GetS3Config() *service.QS3Config {
	qs3config := new(service.QS3Config)
	{
		flag.BoolVar(&qs3config.UseAccessKey, "use_access_key", true, "")
		flag.StringVar(&qs3config.AccessKeyID, "access_key", "", "S3 access key.")
		flag.StringVar(&qs3config.SecretAccessKey, "secret_key", "", "S3 secret key.")
		flag.StringVar(&qs3config.Region, "region", "us-east-1", "S3 region.")
		flag.StringVar(&qs3config.BucketName, "bucket_name", "s3-sse-c-sample", "S3 Bucket name.")
		flag.StringVar(&qs3config.SSECustomerKeyBase64, "sse_customer_key_base64", "pWLFFPtkS6tdlfYdPLB7VnTNUynI+xWjWPU+3uUKub4=", "s3 sse customer key in base64")
		flag.StringVar(&qs3config.SSECustomerKeyMD5Base64, "sse_customer_key_md5_base64", "kGbzQ9GzucxsX+i6c8nK6A==", "s3 sse customer key in base64")
	}
	return qs3config
}
