# Obstor Azure Backend
Obstor Backend adds S3 and [other supported protocol](/docs/protocols) compatibility to Microsoft Azure Blob Storage.

## Run Obstor Backend for Microsoft Azure Blob Storage
### Using Docker
```bash
docker run -p 9000:9000 --name azure-s3 \
 -e "OBSTOR_ROOT_USER=azurestorageaccountname" \
 -e "OBSTOR_ROOT_PASSWORD=azurestorageaccountkey" \
 ghcr.io/obstor/obstor backend azure
```

### Using Binary
```bash
export OBSTOR_ROOT_USER=azureaccountname
export OBSTOR_ROOT_PASSWORD=azureaccountkey
obstor backend azure
```
## Test using Browser Dashboard
Obstor Backend comes with an embedded web based object browser. Point your web browser to http://127.0.0.1:9000 to ensure that your server has started successfully.

![Screenshot](https://raw.githubusercontent.com/obstor/obstor/main/docs/screenshots/dashboard.png)
## Test using an S3 client
You can interact with the backend using rclone or the AWS CLI. Both support filesystems and S3-compatible cloud storage services.

### Configure your client
Configure an rclone S3 remote once:

```bash
rclone config create obstor s3 provider=Other endpoint=http://backend-ip:9000 access_key_id=azureaccountname secret_access_key=azureaccountkey
```

### List containers on Microsoft Azure
```bash
rclone lsd myazure
[2026-05-22 01:50:43]     0B ferenginar/
[2026-05-26 21:43:51]     0B my-container/
[2026-05-26 22:10:11]     0B test-container1/
```

### Use custom access/secret keys

If you do not want to share the credentials of the Azure blob storage with your users/applications, you can set the original credentials in the shell environment using `AZURE_STORAGE_ACCOUNT` and `AZURE_STORAGE_KEY` variables and assign different access/secret keys to `OBSTOR_ROOT_USER` and `OBSTOR_ROOT_PASSWORD`.

### Known limitations
Backend inherits the following Azure limitations:

- Only read-only bucket policy supported at bucket level, all other variations will return API Notimplemented error.
- Bucket names with "." in the bucket name are not supported.
- Non-empty buckets get removed on a DeleteBucket() call.
- _List Multipart Uploads_ always returns empty list.

Other limitations:

- Bucket notification APIs are not supported.

## Explore Further
- [Supported Protocols](/docs/protocols) - S3, SFTP, and more
- `rclone` command-line interface
- `aws` command-line interface
- `obstor-go` Go SDK
