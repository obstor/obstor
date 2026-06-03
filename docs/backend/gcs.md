# Obstor GCS Backend

Obstor GCS Backend allows you to access Google Cloud Storage (GCS) with S3-compatible APIs and [other supported protocols](/docs/protocols)

- [Run Obstor Backend for GCS](#run-obstor-backend-for-gcs)
- [Test Using Browser Dashboard](#test-using-obstor-browser)
- [Test Using Obstor Client](#test-using-obstor-client)

## <a name="run-obstor-backend-for-gcs"></a>1. Run Obstor Backend for GCS

### 1.1 Create a Service Account key for GCS and get the Credentials File
1. Navigate to the [API Console Credentials page](https://console.developers.google.com/project/_/apis/credentials).
2. Select a project or create a new project. Note the project ID.
3. Select the **Create credentials** dropdown on the **Credentials** page, and click **Service account key**.
4. Select **New service account** from the **Service account** dropdown.
5. Populate the **Service account name** and **Service account ID**.
6. Click the dropdown for the **Role** and choose **Storage** > **Storage Admin** *(Full control of GCS resources)*.
7. Click the **Create** button to download a credentials file and rename it to `credentials.json`.

**Note:** For alternate ways to set up *Application Default Credentials*, see [Setting Up Authentication for Server to Server Production Applications](https://developers.google.com/identity/protocols/application-default-credentials).

### 1.2 Run Obstor GCS Backend Using Docker
```bash
docker run -p 9000:9000 --name gcs-s3 \
 -v /path/to/credentials.json:/credentials.json \
 -e "GOOGLE_APPLICATION_CREDENTIALS=/credentials.json" \
 -e "OBSTOR_ROOT_USER=obstoraccountname" \
 -e "OBSTOR_ROOT_PASSWORD=obstoraccountkey" \
 ghcr.io/cloudment/obstor backend gcs yourprojectid
```

### 1.3 Run Obstor GCS Backend Using the Obstor Binary

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
export OBSTOR_ROOT_USER=obstoraccesskey
export OBSTOR_ROOT_PASSWORD=obstorsecretkey
obstor backend gcs yourprojectid
```

## <a name="test-using-obstor-browser"></a>2. Test Using Browser Dashboard

Obstor Backend comes with an embedded web-based object browser that outputs content to http://127.0.0.1:9000. To test that Obstor Backend is running, open a web browser, navigate to http://127.0.0.1:9000, and ensure that the object browser is displayed.

![Screenshot](https://raw.githubusercontent.com/cloudment/obstor/main/docs/screenshots/dashboard.png)

## <a name="test-using-obstor-client"></a>3. Test Using Obstor Client

Obstor Client is a command-line tool called `mc` that provides UNIX-like commands for interacting with the server (e.g. ls, cat, cp, mirror, diff, find, etc.).  `mc` supports file systems and S3-compatible cloud storage services (AWS Signature v2 and v4).

### 3.1 Configure the Backend using Obstor Client

Use the following command to configure the backend:

```bash
mc alias set mygcs http://backend-ip:9000 obstoraccesskey obstorsecretkey
```

### 3.2 List Containers on GCS

Use the following command to list the containers on GCS:

```bash
mc ls mygcs
```

A response similar to this one should be displayed:

```
[2026-05-22 01:50:43 PST]     0B ferenginar/
[2026-05-26 21:43:51 PST]     0B my-container/
[2026-05-26 22:10:11 PST]     0B test-container1/
```

### 3.3 Known limitations
Obstor Backend has the following limitations when used with GCS:

* It only supports read-only and write-only bucket policies at the bucket level; all other variations will return `API Not implemented`.
* The `List Multipart Uploads` and `List Object parts` commands always return empty lists. Therefore, the client must store all of the parts that it has uploaded and use that information when invoking the `_Complete Multipart Upload` command.

Other limitations:

* Bucket notification APIs are not supported.

## <a name="explore-further"></a>4. Explore Further
- [Supported Protocols](/docs/protocols) - S3, SFTP, and more
- `mc` command-line interface
- `aws` command-line interface
- `minio-go` Go SDK
