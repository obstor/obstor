"use server";

import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { rpc } from "./rpc";

// Types
export interface NamedPolicy {
  name: string;
  policy: string;
}

export interface IAMUser {
  accessKey: string;
  status: "enabled" | "disabled";
  policies: string[];
  pendingSecretKey?: string;
}

export interface BucketSettings {
  name: string;
  publicAccess: "private" | "public-read" | "public-read-write";
  versioning: boolean;
  objectLocking: boolean;
  quotaEnabled: boolean;
  quotaType: "hard" | "fifo";
  quotaSize: string;
  quotaUnit: "GB" | "TB" | "PB";
  encryptionEnabled: boolean;
  encryptionType: "SSE-S3" | "SSE-KMS";
  kmsKeyId: string;
  tags: { key: string; value: string }[];
  policies: NamedPolicy[];
  users: IAMUser[];
  sftpEnabled: boolean;
  s3Enabled: boolean;
  placementStrategy: "smart" | "custom";
  regions: string[];
}

// Auth
export async function loginAction(formData: FormData) {
  const accessKey = formData.get("accessKey") as string;
  const secretKey = formData.get("secretKey") as string;

  try {
    const result = await rpcUnauthed<{ token: string }>("Login", {
      username: accessKey,
      password: secretKey,
    });

    const cookieStore = await cookies();
    cookieStore.set("obstor_token", result.token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: 60 * 60 * 24 * 7,
    });
  } catch {
    return { error: "Invalid access key or secret key" };
  }

  redirect("/");
}

async function rpcUnauthed<T>(method: string, params: Record<string, unknown> = {}): Promise<T> {
  const endpoint = process.env.OBSTOR_ENDPOINT || "http://localhost:9000";
  const res = await fetch(`${endpoint}/obstor/webrpc`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "User-Agent": "Mozilla/5.0 Obstor Dashboard" },
    body: JSON.stringify({ id: 1, jsonrpc: "2.0", method: `web.${method}`, params }),
    cache: "no-store",
  });
  const data = await res.json();
  if (data.error) throw new Error(data.error.message);
  return data.result as T;
}

export async function logoutAction() {
  const cookieStore = await cookies();
  cookieStore.delete("obstor_token");
  redirect("/login");
}

export async function changePasswordAction(formData: FormData) {
  const currentAccessKey = formData.get("currentAccessKey") as string;
  const currentSecretKey = formData.get("currentSecretKey") as string;
  const newAccessKey = formData.get("newAccessKey") as string;
  const newSecretKey = formData.get("newSecretKey") as string;

  try {
    await rpc("SetAuth", { currentAccessKey, currentSecretKey, newAccessKey, newSecretKey });
    return { success: true };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to change password" };
  }
}

// Buckets CRUD
export async function createBucketAction(formData: FormData) {
  const bucketName = formData.get("bucketName") as string;
  try {
    await rpc("MakeBucket", { bucketName });
    return { success: true, bucketName };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to create bucket" };
  }
}

export async function deleteBucketAction(bucketName: string) {
  try {
    // Delete user if it doesn't have policies attached
    const users = await listUsers(bucketName);
    const allPolicies = await listPolicies();
    const bucketScoped = new Set(
      allPolicies.filter((p) => policyTargetsBucket(p.policy, bucketName)).map((p) => p.name),
    );

    for (const user of users) {
      const remaining = user.policies.filter((pn) => !bucketScoped.has(pn));
      if (remaining.length === 0) {
        await rpc("RemoveIAMUser", { accessKey: user.accessKey });
      } else {
        await rpc("SetIAMUserPolicy", {
          accessKey: user.accessKey,
          policies: remaining.join(","),
        });
      }
    }

    // Delete bucket-scoped policies
    for (const pName of bucketScoped) {
      try {
        await rpc("DeleteCannedPolicy", { name: pName });
      } catch {
        // Todo: Add error handling
      }
    }

    await rpc("DeleteBucket", { bucketName });
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to delete bucket" };
  }
  redirect("/");
}

function policyTargetsBucket(policyJSON: string, bucket: string): boolean {
  try {
    const parsed = JSON.parse(policyJSON);
    const arnPrefix = `arn:aws:s3:::${bucket}`;
    const statements = Array.isArray(parsed.Statement) ? parsed.Statement : [parsed.Statement];
    for (const st of statements) {
      if (!st) continue;
      const resources = Array.isArray(st.Resource) ? st.Resource : [st.Resource];
      for (const r of resources) {
        if (typeof r !== "string") continue;
        if (r === arnPrefix || r.startsWith(`${arnPrefix}/`) || r.startsWith(`${arnPrefix}*`)) {
          return true;
        }
      }
    }
  } catch {
    // Todo: Add error handling
  }
  return false;
}

// Bucket Settings
function publicAccessToBackend(policy: BucketSettings["publicAccess"]): string {
  switch (policy) {
    case "public-read":
      return "readonly";
    case "public-read-write":
      return "readwrite";
    default:
      return "none";
  }
}

export async function getBucketSettingsAction(
  bucketName: string,
): Promise<BucketSettings | { error: string }> {
  try {
    let cannedPublic: BucketSettings["publicAccess"] = "private";
    try {
      const res = await rpc<{ policy: string }>("GetBucketPolicy", {
        bucketName,
        prefix: "",
      });
      if (res.policy === "readonly") cannedPublic = "public-read";
      else if (res.policy === "readwrite" || res.policy === "writeonly")
        cannedPublic = "public-read-write";
    } catch {
      // Todo: Add error handling
    }

    const policies = await listPolicies(bucketName);
    const users = await listUsers(bucketName);

    let s3Enabled = true;
    let sftpEnabled = true;
    try {
      const res = await rpc<{ s3Enabled: boolean; sftpEnabled: boolean }>("GetBucketToggles", {
        bucketName,
      });
      s3Enabled = res.s3Enabled;
      sftpEnabled = res.sftpEnabled;
    } catch {
      // Todo: Add error handling
    }

    return {
      name: bucketName,
      publicAccess: cannedPublic,
      versioning: false,
      objectLocking: false,
      quotaEnabled: false,
      quotaType: "hard",
      quotaSize: "",
      quotaUnit: "GB",
      encryptionEnabled: false,
      encryptionType: "SSE-S3",
      kmsKeyId: "",
      tags: [],
      policies,
      users,
      sftpEnabled,
      s3Enabled,
      placementStrategy: "smart",
      regions: [],
    };
  } catch (err) {
    return {
      error: err instanceof Error ? err.message : "Failed to load bucket settings",
    };
  }
}

export async function createBucketWithSettingsAction(
  settings: BucketSettings,
): Promise<{ success: true; bucketName: string } | { error: string }> {
  try {
    await rpc("MakeBucket", { bucketName: settings.name });
    await applyBucketSettings(settings, true);
    revalidatePath("/");
    revalidatePath(`/${settings.name}`);
    return { success: true, bucketName: settings.name };
  } catch (err) {
    return {
      error: err instanceof Error ? err.message : "Failed to create bucket",
    };
  }
}

export async function updateBucketSettingsAction(
  settings: BucketSettings,
): Promise<{ success: true } | { error: string }> {
  try {
    await applyBucketSettings(settings, false);
    revalidatePath("/");
    revalidatePath(`/${settings.name}`);
    return { success: true };
  } catch (err) {
    return {
      error: err instanceof Error ? err.message : "Failed to update bucket settings",
    };
  }
}

async function applyBucketSettings(settings: BucketSettings, _isCreate: boolean) {
  await rpc("SetBucketPolicy", {
    bucketName: settings.name,
    prefix: "",
    policy: publicAccessToBackend(settings.publicAccess),
  });

  // Feature Toggles
  await rpc("SetBucketToggles", {
    bucketName: settings.name,
    s3Enabled: settings.s3Enabled,
    sftpEnabled: settings.sftpEnabled,
  });

  // List IAM policies
  const existing = await listPolicies(settings.name);
  const desiredNames = new Set(settings.policies.map((p) => p.name));
  for (const ex of existing) {
    if (!desiredNames.has(ex.name)) {
      try {
        await rpc("DeleteCannedPolicy", { name: ex.name });
      } catch {
        // Todo: Add error handling
      }
    }
  }
  for (const p of settings.policies) {
    if (!p.name.trim() || !p.policy.trim()) continue;
    // Use BUCKET_NAME for policy placeholder
    const resolved = p.policy.split("BUCKET_NAME").join(settings.name);
    await rpc("SetCannedPolicy", { name: p.name, policy: resolved });
  }

  // Create users that were added on bucket modal
  for (const user of settings.users) {
    if (!user.pendingSecretKey) continue;
    await rpc("AddIAMUser", {
      accessKey: user.accessKey,
      secretKey: user.pendingSecretKey,
      policy: "", // Policies attached later
    });
  }

  // Attach policies
  const allGlobal = await listPolicies();
  const bucketScoped = new Set(
    allGlobal.filter((p) => policyTargetsBucket(p.policy, settings.name)).map((p) => p.name),
  );

  for (const user of settings.users) {
    let current: IAMUser | undefined;
    try {
      const fresh = await listUsersAll();
      current = fresh.find((u) => u.accessKey === user.accessKey);
    } catch {
      // Todo: Add error handling
    }
    const others = current ? current.policies.filter((pn) => !bucketScoped.has(pn)) : [];
    const finalPolicies = Array.from(new Set([...others, ...user.policies]));
    await rpc("SetIAMUserPolicy", {
      accessKey: user.accessKey,
      policies: finalPolicies.join(","),
    });
  }
}

// List IAM canned policies
async function listPolicies(bucketName = ""): Promise<NamedPolicy[]> {
  try {
    const res = await rpc<{ policies?: NamedPolicy[] }>("ListCannedPolicies", { bucketName });
    return res.policies || [];
  } catch {
    return [];
  }
}

export async function listPoliciesAction(bucketName: string): Promise<NamedPolicy[]> {
  return listPolicies(bucketName);
}

export async function savePolicyAction(name: string, policyJSON: string) {
  try {
    await rpc("SetCannedPolicy", { name, policy: policyJSON });
    return { success: true };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to save policy" };
  }
}

export async function deletePolicyAction(name: string) {
  try {
    await rpc("DeleteCannedPolicy", { name });
    return { success: true };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to delete policy" };
  }
}

// List IAM users
async function listUsers(bucketName = ""): Promise<IAMUser[]> {
  try {
    const res = await rpc<{ users?: IAMUser[] }>("ListIAMUsers", { bucketName });
    return (res.users || []).map((u) => ({ ...u, policies: u.policies ?? [] }));
  } catch {
    return [];
  }
}

async function listUsersAll(): Promise<IAMUser[]> {
  return listUsers("");
}

export async function listUsersAction(bucketName: string): Promise<IAMUser[]> {
  return listUsers(bucketName);
}

export async function addUserAction(
  accessKey: string,
  secretKey: string,
  policy: string,
): Promise<{ accessKey: string; secretKey: string } | { error: string }> {
  try {
    const res = await rpc<{ accessKey: string; secretKey: string }>("AddIAMUser", {
      accessKey,
      secretKey,
      policy,
    });
    return { accessKey: res.accessKey, secretKey: res.secretKey };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to create user" };
  }
}

export async function removeUserAction(accessKey: string) {
  try {
    await rpc("RemoveIAMUser", { accessKey });
    return { success: true };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to remove user" };
  }
}

export async function setUserStatusAction(accessKey: string, enabled: boolean) {
  try {
    await rpc("SetIAMUserStatus", { accessKey, enabled });
    return { success: true };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to set user status" };
  }
}

export async function setUserPolicyAction(accessKey: string, policiesCSV: string) {
  try {
    await rpc("SetIAMUserPolicy", { accessKey, policies: policiesCSV });
    return { success: true };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to attach policy" };
  }
}

// Detach bucket access for user
export async function detachUserFromBucketAction(
  bucketName: string,
  accessKey: string,
): Promise<{ deleted: boolean } | { error: string }> {
  try {
    const all = await listPolicies();
    const bucketScoped = new Set(
      all.filter((p) => policyTargetsBucket(p.policy, bucketName)).map((p) => p.name),
    );
    const users = await listUsers();
    const user = users.find((u) => u.accessKey === accessKey);
    if (!user) return { error: "User not found" };

    const remaining = user.policies.filter((pn) => !bucketScoped.has(pn));
    if (remaining.length === 0) {
      await rpc("RemoveIAMUser", { accessKey });
      return { deleted: true };
    }
    await rpc("SetIAMUserPolicy", {
      accessKey,
      policies: remaining.join(","),
    });
    return { deleted: false };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to detach user" };
  }
}

// Objects
export async function deleteObjectAction(bucketName: string, objectName: string) {
  try {
    await rpc("RemoveObject", { bucketname: bucketName, objects: [objectName] });
    return { success: true };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to delete object" };
  }
}

export async function getShareLink(bucketName: string, objectName: string, expiry = 300) {
  try {
    const result = await rpc<{ url: string }>("PresignedGet", {
      host: process.env.OBSTOR_HOST || "localhost:9000",
      bucket: bucketName,
      object: objectName,
      expiry,
    });
    return { url: result.url };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to generate link" };
  }
}

export async function getUploadURL(bucketName: string, prefix: string, objectName: string) {
  try {
    const result = await rpc<{ url: string }>("PresignedPut", {
      host: process.env.OBSTOR_HOST || "localhost:9000",
      bucket: bucketName,
      prefix,
      object: objectName,
    });
    return { url: result.url };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to get upload URL" };
  }
}

export async function getObjectChecksums(bucketName: string, objectName: string) {
  try {
    const result = await rpc<{ md5: string; sha1: string; sha256: string; sha512: string }>(
      "GetObjectChecksums",
      { bucketName, objectName },
    );
    return result;
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to compute checksums" };
  }
}

export async function getDownloadURL(bucketName: string, objectName: string) {
  try {
    const result = await rpc<{ url: string }>("PresignedGet", {
      host: process.env.OBSTOR_HOST || "localhost:9000",
      bucket: bucketName,
      object: objectName,
      expiry: 300,
    });
    return { url: result.url };
  } catch (err) {
    return { error: err instanceof Error ? err.message : "Failed to get download URL" };
  }
}
