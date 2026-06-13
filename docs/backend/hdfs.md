# Obstor HDFS Backend
Obstor HDFS backend adds S3 API and [other supported protocol](/docs/protocols) support to Hadoop HDFS filesystem. Applications can use both the S3 and file APIs concurrently without requiring any data migration. Since the backend is stateless and shared-nothing, you may elastically provision as many Obstor instances as needed to distribute the load.

> NOTE: Intention of this backend implementation it to make it easy to migrate your existing data on HDFS clusters to Obstor clusters using standard tools like `rclone` or `aws-cli`, if the goal is to use HDFS perpetually we recommend that HDFS should be used directly for all write operations.

## Run Obstor Backend for HDFS Storage

### Using Binary
Namenode information is obtained by reading `core-site.xml` automatically from your hadoop environment variables *$HADOOP_HOME*
```bash
export OBSTOR_ROOT_USER=obstor
export OBSTOR_ROOT_PASSWORD=obstor123
obstor backend hdfs
```

You can also override the namenode endpoint as shown below.
```bash
export OBSTOR_ROOT_USER=obstor
export OBSTOR_ROOT_PASSWORD=obstor123
obstor backend hdfs hdfs://namenode:8200
```

### Using Docker
Using docker is experimental, most Hadoop environments are not dockerized and may require additional steps in getting this to work properly. You are better off just using the binary in this situation.
```bash
docker run -p 9000:9000 \
 --name hdfs-s3 \
 -e "OBSTOR_ROOT_USER=obstor" \
 -e "OBSTOR_ROOT_PASSWORD=obstor123" \
 ghcr.io/obstor/obstor backend hdfs hdfs://namenode:8200
```

### Setup Kerberos

Obstor supports two kerberos authentication methods, keytab and ccache.

To enable kerberos authentication, you need to set `hadoop.security.authentication=kerberos` in the HDFS config file.

```xml
<property>
  <name>hadoop.security.authentication</name>
  <value>kerberos</value>
</property>
```

Obstor will load `krb5.conf` from environment variable `KRB5_CONFIG` or default location `/etc/krb5.conf`.
```bash
export KRB5_CONFIG=/path/to/krb5.conf
```

If you want Obstor to use ccache for authentication, set environment variable `KRB5CCNAME` to the credential cache file path,
or Obstor will use the default location `/tmp/krb5cc_%{uid}`.
```bash
export KRB5CCNAME=/path/to/krb5cc
```

If you prefer to use keytab, with automatically renewal, you need to config three environment variables:

- `KRB5KEYTAB`: the location of keytab file
- `KRB5USERNAME`: the username
- `KRB5REALM`: the realm

Please note that the username is not principal name.

```bash
export KRB5KEYTAB=/path/to/keytab
export KRB5USERNAME=hdfs
export KRB5REALM=REALM.COM
```

## Test using Browser Dashboard
*Obstor backend* comes with an embedded web based object browser. Point your web browser to http://127.0.0.1:9000 to ensure that your server has started successfully.

![Screenshot](https://raw.githubusercontent.com/obstor/obstor/main/docs/screenshots/dashboard.png)

## Test using an S3 client

You can interact with the backend using rclone or the AWS CLI. Both support filesystems and S3-compatible cloud storage services.

### Configure your client

Configure an rclone S3 remote once:

```bash
rclone config create obstor s3 provider=Other endpoint=http://backend-ip:9000 access_key_id=access_key secret_access_key=secret_key
```

### List buckets on hdfs

```bash
rclone lsd myhdfs
[2026-05-22 01:50:43]     0B user/
[2026-05-26 21:43:51]     0B datasets/
[2026-05-26 22:10:11]     0B assets/
```

### Known limitations
Backend inherits the following limitations of HDFS storage layer:
- No bucket policy support (HDFS has no such concept)
- No bucket notification APIs are not supported (HDFS has no support for fsnotify)
- No server side encryption support (Intentionally not implemented)
- No server side compression support (Intentionally not implemented)
- Concurrent multipart operations are not supported (HDFS lacks safe locking support, or poorly implemented)

## Explore Further
- [Supported Protocols](/docs/protocols) - S3, SFTP, and more
- `rclone` command-line interface
- `aws` command-line interface
- `obstor-go` Go SDK
