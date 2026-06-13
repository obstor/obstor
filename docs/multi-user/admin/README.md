# Obstor Admin Multi-user Quickstart Guide
Obstor supports multiple admin users in addition to default operator credential created during server startup. New admins can be added after server starts up, and server can be configured to deny or allow access to different admin operations for these users. This document explains how to add/remove admin users and modify their access rights.

## Get started
In this document we will explain in detail on how to configure admin users.

### 1. Prerequisites
- Install Obstor - Obstor Quickstart Guide

### 2. Create a new admin user with CreateUser, DeleteUser and ConfigUpdate permissions
Create new canned policy file `adminManageUser.json`. This policy enables admin user to
manage other users.
```json
cat > adminManageUser.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "admin:CreateUser",
        "admin:DeleteUser",
        "admin:ConfigUpdate"
      ],
      "Effect": "Allow",
      "Sid": ""
    },
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:aws:s3:::*"
      ],
      "Sid": ""
    }
  ]
}
EOF
```

Create a new canned policy by name `userManager` using the `adminManageUser.json` policy file through Obstor's dashboard or the API.

Create a new admin user `admin1` on Obstor.

Once the user is successfully created you can now apply the `userManager` policy for this user.

This admin user will then be allowed to perform create/delete user operations.

### 3. Create another user user1 with attached policy user1policy
Authenticating as `admin1` against Obstor's dashboard or API, create the user `user1`, create the `user1policy` policy from `~/user1policy.json`, and attach that policy to `user1`.

### 4. List of permissions defined for admin operations
#### Config management permissions
- admin:ConfigUpdate

#### User management permissions
- admin:CreateUser
- admin:DeleteUser
- admin:ListUsers
- admin:EnableUser
- admin:DisableUser
- admin:GetUser

#### Service management permissions
- admin:ServerInfo
- admin:ServerUpdate
- admin:StorageInfo
- admin:DataUsageInfo
- admin:TopLocks
- admin:OBDInfo
- admin:Profiling,
- admin:ServerTrace
- admin:ConsoleLog
- admin:KMSKeyStatus

#### User/Group management permissions
- admin:AddUserToGroup
- admin:RemoveUserFromGroup
- admin:GetGroup
- admin:ListGroups
- admin:EnableGroup
- admin:DisableGroup

#### Policy management permissions
- admin:CreatePolicy
- admin:DeletePolicy
- admin:GetPolicy
- admin:AttachUserOrGroupPolicy
- admin:ListUserPolicies

#### Give full admin permissions
- admin:*

### 5. Using an external IDP for admin users
Admin users can also be externally managed by an IDP by configuring admin policy with
special permissions listed above. Follow [Obstor STS Quickstart Guide](/docs/sts) to manage users with an IDP.

## Explore Further
- [Obstor STS Quickstart Guide](/docs/sts)
- Obstor Admin Complete Guide
- [The Obstor documentation website](/docs)
