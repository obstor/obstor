# Obstor Multi-user Quickstart Guide
Obstor supports multiple long term users in addition to default user created during server startup. New users can be added after server starts up, and server can be configured to deny or allow access to buckets and resources to each of these users. This document explains how to add/remove users and modify their access rights.

## Get started
In this document we will explain in detail on how to configure multiple users.

### 1. Prerequisites
- Install Obstor - Obstor Quickstart Guide
- Configure etcd (optional needed only in backend or federation mode) - [Etcd V3 Quickstart Guide](/docs/sts/etcd)

### 2. Create a new user with a canned policy
User, group, and policy management is done through Obstor's Admin API or the dashboard. The steps below refer to those operations without repeating the tooling each time.

The server provides default canned policies `writeonly`, `readonly` and `readwrite` *(these apply to all resources on the server)*. You can also define custom policies.

Create a canned policy file `getonly.json`. This policy lets users download all objects under `my-bucketname`.
```json
cat > getonly.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:GetObject"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:aws:s3:::my-bucketname/*"
      ],
      "Sid": ""
    }
  ]
}
EOF
```

Then:
- Create a canned policy named `getonly` from `getonly.json`.
- Create a new user `newuser`.
- Apply the `getonly` policy to `newuser`.

### 3. Create a new group
- Create a group `newgroup` and add `newuser` to it.
- Apply the `getonly` policy to `newgroup`.

### 4. Disable user
- Disable user `newuser`.
- Disable group `newgroup`.

### 5. Remove user or group
- Remove user `newuser`.
- Remove `newuser` from a group.
- Remove group `newgroup`.

### 6. Change user or group policy
- Change `newuser`'s policy to the `putonly` canned policy.
- Change `newgroup`'s policy to the `putonly` canned policy.

### 7. List users or groups
- List all enabled and disabled users.
- List all enabled and disabled groups.

### 8. Configure your client
Configure an S3 client such as rclone with the new user's credentials, then read an object. With rclone, create the remote once:
```bash
rclone config create myobstor-newuser s3 provider=Other endpoint=http://localhost:9000 access_key_id=newuser secret_access_key=newuser123
rclone cat myobstor-newuser:my-bucketname/my-objectname
```

The same read with the AWS CLI:
```bash
aws --endpoint-url http://localhost:9000 s3 cp s3://my-bucketname/my-objectname -
```

### Policy Variables
You can use policy variables in the *Resource* element and in string comparisons in the *Condition* element.

You can use a policy variable in the Resource element, but only in the resource portion of the ARN. This portion of the ARN appears after the 5th colon (:). You can't use a variable to replace parts of the ARN before the 5th colon, such as the service or account. The following policy might be attached to a group. It gives each of the users in the group full programmatic access to a user-specific object (their own "home directory") in Obstor.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": ["s3:ListBucket"],
      "Effect": "Allow",
      "Resource": ["arn:aws:s3:::mybucket"],
      "Condition": {"StringLike": {"s3:prefix": ["${aws:username}/*"]}}
    },
    {
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Effect": "Allow",
      "Resource": ["arn:aws:s3:::mybucket/${aws:username}/*"]
    }
  ]
}
```

If the user is authenticating using an STS credential which was authorized from OpenID connect we allow all `jwt:*` variables specified in the JWT specification, custom `jwt:*` or extensions are not supported.

List of policy variables for OpenID based STS.
```
"jwt:sub"
"jwt:iss"
"jwt:aud"
"jwt:jti"
"jwt:upn"
"jwt:name"
"jwt:groups"
"jwt:given_name"
"jwt:family_name"
"jwt:middle_name"
"jwt:nickname"
"jwt:preferred_username"
"jwt:profile"
"jwt:picture"
"jwt:website"
"jwt:email"
"jwt:gender"
"jwt:birthdate"
"jwt:phone_number"
"jwt:address"
"jwt:scope"
"jwt:client_id"
```

Following example shows OpenID users with full programmatic access to a OpenID user-specific directory (their own "home directory") in Obstor.
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": ["s3:ListBucket"],
      "Effect": "Allow",
      "Resource": ["arn:aws:s3:::mybucket"],
      "Condition": {"StringLike": {"s3:prefix": ["${jwt:preferred_username}/*"]}}
    },
    {
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Effect": "Allow",
      "Resource": ["arn:aws:s3:::mybucket/${jwt:preferred_username}/*"]
    }
  ]
}
```

If the user is authenticating using an STS credential which was authorized from AD/LDAP we allow `ldap:*` variables, currently only supports `ldap:user`. Following example shows LDAP users full programmatic access to a LDAP user-specific directory (their own "home directory") in Obstor.
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": ["s3:ListBucket"],
      "Effect": "Allow",
      "Resource": ["arn:aws:s3:::mybucket"],
      "Condition": {"StringLike": {"s3:prefix": ["${ldap:user}/*"]}}
    },
    {
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Effect": "Allow",
      "Resource": ["arn:aws:s3:::mybucket/${ldap:user}/*"]
    }
  ]
}
```

#### Common information available in all requests

- *aws:CurrentTime* - This can be used for conditions that check the date and time.
- *aws:EpochTime* - This is the date in epoch or Unix time, for use with date/time conditions.
- *aws:PrincipalType* - This value indicates whether the principal is an account (Root credential), user (Obstor user), or assumed role (STS)
- *aws:SecureTransport* - This is a Boolean value that represents whether the request was sent over TLS.
- *aws:SourceIp* - This is the requester's IP address, for use with IP address conditions. If running behind Nginx like proxies, Obstor preserve's the source IP.

```json
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "s3:ListBucket*",
    "Resource": "arn:aws:s3:::mybucket",
    "Condition": {"IpAddress": {"aws:SourceIp": "203.0.113.0/24"}}
  }
}
```

- *aws:UserAgent* - This value is a string that contains information about the requester's client application. This string is generated by the client and can be unreliable. You can only use this context key from SDKs which standardize the User-Agent string.
- *aws:username* - This is a string containing the friendly name of the current user, this value would point to STS temporary credential in `AssumeRole`ed requests, instead use `jwt:preferred_username` in case of OpenID connect and `ldap:user` in case of AD/LDAP connect. *aws:userid* is an alias to *aws:username* in Obstor.


## Explore Further
- [Obstor STS Quickstart Guide](/docs/sts)
- Obstor Admin Complete Guide
- [The Obstor documentation website](/docs)
