# S3 SSE-C Sample Code (Go)

This is a sample code to demo how to implement S3 SSE-C in Go.

## Contents
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Test](#test)

## Installation

To install this package, you need to install Go and set your Go workspace first.

### Download and install it:

```sh
$ go get -u github.com/anrylu/s3-sse-c-sample
```

### Use dep as the vendor tool

```sh
$ cd s3-sse-c-sample
$ go get -u github.com/golang/dep/cmd/dep
$ dep ensure
```

## Quick start

To execute this project, you need to have access & secret key to access S3. Or you can setup an instance with IAM role permission.

### Setup S3 Region & Bucket

```sh
$ export S3_SSE_C_SAMPLE_REGION="us-east-1"
$ export S3_SSE_C_SAMPLE_BUCKET_NAME="s3-sse-c-sample"
```

### Setup Access & Secret Keys

```sh
$ export S3_SSE_C_SAMPLE_ACCESS_KEY=<YOU_ACCESS_KEY>
$ export S3_SSE_C_SAMPLE_SECRET_KEY=<YOU_SECRET_KEY>
```

### Generate Customer Encryption Key

```sh
$ openssl rand -out PlaintextKeyMaterial.bin 32
$ cat PlaintextKeyMaterial.bin | base64
```

### Get Customer Encrytion Key MD5

```sh
$ cat PlaintextKeyMaterial.bin | openssl dgst -md5 -binary | base64
```

### Setup Customer Encryption Key

```sh
$ export S3_SSE_C_SAMPLE_SSE_CUSTOMER_KEY_BASE64=<YOU_ENCRYPTION_KEY_IN_BASE64>
$ export S3_SSE_C_SAMPLE_SSE_CUSTOMER_KEY_MD5_BASE64=<YOU_ENCRYPTION_KEY_MD5_IN_BASE64>
```

### Start Server

```sh
$ go run cmd/server/main.go
```

## Test

To test, you need to have curl installed.

### Upload

```sh
curl -vvv -X POST http://localhost:8080/upload -F "file=@./test.jpg" -H "Content-Type: multipart/form-data"
```

### Check on S3 Management Console

Just check if the file is successfully uploaded to S3 and see if it has "Server-side encryption".

### Download

```sh
curl -vvv http://localhost:8080/download/test.jpg -o test-download.jpg
```

### Compare files

Just check if test.jpg & test-download.jpg are the same.