# Obstor Admin Multi-user Quickstart Guide
Obstor supports multiple admin users in addition to default operator credential created during server startup. New admins can be added after server starts up, and server can be configured to deny or allow access to different admin operations for these users. This document explains how to add/remove admin users and modify their access rights.

## Get started
In this document we will explain in detail on how to configure admin users.

### 1. Prerequisites
- Install mc - Obstor Client Quickstart Guide
- Install Obstor - Obstor Quickstart Guide

### 2. Create a new admin user with CreateUser, DeleteUser and ConfigUpdate permissions
Use [`mc admin policy`](/docs/multi-user/admin) to create custom admin policies.

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

Create new canned policy by name `userManager` using `userManager.json` policy file.
```bash
mc admin policy add myobstor userManager adminManageUser.json
```

Create a new admin user `admin1` on Obstor use `mc admin user`.
```bash
mc admin user add myobstor admin1 admin123
```

Once the user is successfully created you can now apply the `userManage` policy for this user.

```bash
mc admin policy set myobstor userManager user=admin1
```

This admin user will then be allowed to perform create/delete user operations via `mc admin user`

### 3. Configure `mc` and create another user user1 with attached policy user1policy
```bash
mc alias set myobstor-admin1 http://localhost:9000 admin1 admin123 --api s3v4

mc admin user add myobstor-admin1 user1 user123
mc admin policy add myobstor-admin1 user1policy ~/user1policy.json
mc admin policy set myobstor-admin1 user1policy user=user1
```

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
- Obstor Client Complete Guide
- [Obstor STS Quickstart Guide](/docs/sts)
- Obstor Admin Complete Guide
- [The Obstor documentation website](/docs)
