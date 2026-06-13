import Link from "next/link";
import { ObjectBrowser } from "@/components/ObjectBrowser";
import { formatDate, humanSize, rpc } from "@/lib/rpc";
import { safeDisplayName } from "@/lib/safe-name";

interface Props {
  params: Promise<{ bucket: string }>;
  searchParams: Promise<{ prefix?: string }>;
}

export default async function BucketPage({ params, searchParams }: Props) {
  const { bucket } = await params;
  const { prefix = "" } = await searchParams;
  const bucketName = decodeURIComponent(bucket);

  let objects: {
    name: string;
    size: number;
    lastModified: string;
    contentType: string;
    etag: string;
  }[] = [];
  let error = "";
  let creationDate = "";
  let policyType = "none";
  const objectLocations: Record<string, string[]> = {};

  try {
    const result = await rpc<{
      objects:
        | { name: string; size: number; lastModified: string; contentType: string; etag: string }[]
        | null;
    }>("ListObjects", { bucketName, prefix, marker: "" });
    objects = result.objects || [];
  } catch (err) {
    error = err instanceof Error ? err.message : "Failed to load objects";
  }

  try {
    const bucketsRes = await rpc<{
      buckets: { name: string; creationDate: string }[] | null;
    }>("ListBuckets");
    const match = bucketsRes.buckets?.find((b) => b.name === bucketName);
    if (match) creationDate = match.creationDate;
  } catch {}

  try {
    const policyRes = await rpc<{ policy: string }>("GetBucketPolicy", {
      bucketName,
      prefix: "",
    });
    policyType = policyRes.policy || "none";
  } catch {}

  // Fetch per-object node placements
  try {
    const locRes = await rpc<{
      objects: { name: string; endpoints: string[] }[] | null;
    }>("GetObjectLocations", { bucketName, prefix });
    for (const obj of locRes.objects || []) {
      objectLocations[obj.name] = obj.endpoints || [];
    }
  } catch {}

  const isPublic = policyType !== "none";
  const policyLabel =
    policyType === "readonly"
      ? "Public Read"
      : policyType === "readwrite"
        ? "Public Read/Write"
        : policyType === "writeonly"
          ? "Public Write"
          : "Private";

  const endpoint = process.env.OBSTOR_HOST || "localhost:9000";
  const httpUrl = `http://${endpoint}/${bucketName}`;

  const folders = objects.filter((o) => o.name.endsWith("/"));
  const files = objects.filter((o) => !o.name.endsWith("/"));

  return (
    <div>
      {/* Bucket info bar */}
      <div className="mb-4 flex items-center gap-5 rounded-lg border border-border bg-abyss px-4 py-3">
        <div className="flex items-center gap-2">
          <span className="icon-[lucide--hard-drive] text-accent text-xs" />
          <span className="font-display font-semibold text-sm">
            <bdi>{safeDisplayName(bucketName)}</bdi>
          </span>
        </div>

        {creationDate && (
          <>
            <div className="h-4 w-px bg-border" />
            <span className="font-mono text-[11px] text-text-muted">
              Created {formatDate(creationDate)}
            </span>
          </>
        )}

        {/* Visibility + URL - right aligned */}
        <div className="ml-auto flex items-center gap-2">
          <span className={`h-1.5 w-1.5 rounded-full ${isPublic ? "bg-up" : "bg-text-muted"}`} />
          <span className="font-mono text-[11px] text-text-secondary">{policyLabel}</span>
          <span className="text-text-muted">|</span>
          <span className="font-mono text-[11px] text-text-muted">{httpUrl}</span>
        </div>
      </div>

      {/* Breadcrumb */}
      {prefix && (
        <div className="mb-4 flex items-center gap-2">
          <Link
            href={`/${encodeURIComponent(bucketName)}`}
            className="font-mono text-accent text-sm transition-colors hover:text-accent-bright"
          >
            <bdi>{safeDisplayName(bucketName)}</bdi>
          </Link>
          {prefix
            .split("/")
            .filter(Boolean)
            .map((part, i, arr) => {
              const path = `${arr.slice(0, i + 1).join("/")}/`;
              return (
                <span key={path} className="flex items-center gap-2">
                  <span className="text-text-muted">/</span>
                  <Link
                    href={`/${encodeURIComponent(bucketName)}?prefix=${encodeURIComponent(path)}`}
                    className="font-mono text-sm text-text-secondary transition-colors hover:text-text-primary"
                  >
                    <bdi>{safeDisplayName(part)}</bdi>
                  </Link>
                </span>
              );
            })}
        </div>
      )}

      {error ? (
        <div className="flex items-center gap-2 rounded-lg border border-danger/20 bg-danger/5 px-4 py-3">
          <span className="icon-[lucide--alert-circle] text-danger text-sm" />
          <span className="font-body text-danger text-sm">{error}</span>
        </div>
      ) : (
        <ObjectBrowser
          bucketName={bucketName}
          prefix={prefix}
          folders={folders.map((f) => f.name)}
          files={files.map((f) => ({
            name: f.name,
            size: humanSize(f.size),
            sizeBytes: f.size,
            lastModified: formatDate(f.lastModified),
            contentType: f.contentType || "",
            etag: f.etag || "",
            locations: objectLocations[f.name] || [],
          }))}
        />
      )}
    </div>
  );
}
