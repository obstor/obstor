#!/bin/sh

HOST=localhost:9000
SRC="aws --endpoint-url http://$HOST s3api"

$SRC create-bucket --bucket bucket --object-lock-enabled-for-bucket
# $SRC put-bucket-versioning --bucket bucket --versioning-configuration Status=Enabled
cat > repladmin-policy-source.json <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": [
            "admin:SetBucketTarget",
            "admin:GetBucketTarget"
        ],
        "Effect": "Allow",
        "Sid": ""
     },
     {
      "Effect": "Allow",
      "Action": [
       "s3:GetReplicationConfiguration",
       "s3:PutReplicationConfiguration",
       "s3:ListBucket",
       "s3:ListBucketMultipartUploads",
       "s3:GetBucketLocation",
       "s3:GetBucketVersioning"
      ],
      "Resource": [
       "arn:aws:s3:::bucket"
      ]
     }
    ]
   }
EOF
cat ./repladmin-policy-source.json
cat > replpolicy.json <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
  {
   "Effect": "Allow",
   "Action": [
    "s3:GetReplicationConfiguration",
    "s3:ListBucket",
    "s3:ListBucketMultipartUploads",
    "s3:GetBucketLocation",
    "s3:GetBucketVersioning",
    "s3:GetBucketObjectLockConfiguration"
   ],
   "Resource": [
    "arn:aws:s3:::bucket"
   ]
  },
  {
   "Effect": "Allow",
   "Action": [
    "s3:GetReplicationConfiguration",
    "s3:ReplicateTags",
    "s3:AbortMultipartUpload",
    "s3:GetObject",
    "s3:GetObjectVersion",
    "s3:GetObjectVersionTagging",
    "s3:PutObject",
    "s3:DeleteObject",
    "s3:ReplicateObject",
    "s3:ReplicateDelete"
   ],
   "Resource": [
    "arn:aws:s3:::bucket/*"
   ]
  }
 ]
}
EOF
cat ./replpolicy.json
echo "Use this ARN to add replication rules on Obstor's dashboard or API"
