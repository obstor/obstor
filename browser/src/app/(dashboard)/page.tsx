import { humanSize, rpc } from "@/lib/rpc";

interface StorageResult {
  used: number;
}

interface BucketResult {
  buckets: { name: string; creationDate: string }[] | null;
}

interface ServerResult {
  ObstorVersion: string;
  ObstorPlatform: string;
  ObstorRuntime: string;
  ObstorGlobalInfo: { nodes?: number; drives?: number };
}

export default async function DashboardHome() {
  let used = 0;
  let bucketCount = 0;
  let version = "";
  let platform = "";
  let runtime = "";
  let nodes = 0;
  let drives = 0;

  try {
    const s = await rpc<StorageResult>("StorageInfo");
    used = s.used;
  } catch {}

  try {
    const b = await rpc<BucketResult>("ListBuckets");
    bucketCount = b.buckets?.length || 0;
  } catch {}

  try {
    const sv = await rpc<ServerResult>("ServerInfo");
    version = sv.ObstorVersion;
    platform = sv.ObstorPlatform;
    runtime = sv.ObstorRuntime;
    nodes = sv.ObstorGlobalInfo?.nodes || 1;
    drives = sv.ObstorGlobalInfo?.drives || 0;
  } catch {}

  return (
    <div>
      <h1 className="mb-6 font-display font-semibold text-2xl text-text-primary">
        Cluster Overview
      </h1>

      {/* Stats grid */}
      <div className="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <div className="rounded-xl border border-border bg-abyss p-5">
          <div className="mb-2 flex items-center gap-2">
            <span className="icon-[lucide--database] text-accent text-sm" />
            <span className="font-mono text-[10px] text-text-muted uppercase tracking-wider">
              Storage Used
            </span>
          </div>
          <p className="font-bold font-display text-2xl text-text-primary">{humanSize(used)}</p>
          <p className="mt-1 font-mono text-[10px] text-text-muted">
            Across all nodes | Replicated
          </p>
        </div>

        <div className="rounded-xl border border-border bg-abyss p-5">
          <div className="mb-2 flex items-center gap-2">
            <span className="icon-[lucide--hard-drive] text-accent text-sm" />
            <span className="font-mono text-[10px] text-text-muted uppercase tracking-wider">
              Buckets
            </span>
          </div>
          <p className="font-bold font-display text-2xl text-text-primary">{bucketCount}</p>
          <p className="mt-1 font-mono text-[10px] text-text-muted">Sharded across cluster</p>
        </div>

        <div className="rounded-xl border border-border bg-abyss p-5">
          <div className="mb-2 flex items-center gap-2">
            <span className="icon-[lucide--server] text-accent text-sm" />
            <span className="font-mono text-[10px] text-text-muted uppercase tracking-wider">
              Nodes
            </span>
          </div>
          <p className="font-bold font-display text-2xl text-text-primary">{nodes}</p>
          <p className="mt-1 font-mono text-[10px] text-text-muted">
            {drives} drive{drives !== 1 ? "s" : ""} across cluster
          </p>
        </div>

        <div className="rounded-xl border border-border bg-abyss p-5">
          <div className="mb-2 flex items-center gap-2">
            <span className="icon-[lucide--server] text-accent text-sm" />
            <span className="font-mono text-[10px] text-text-muted uppercase tracking-wider">
              Replication
            </span>
          </div>
          <p className="font-bold font-display text-2xl text-text-primary">2x min</p>
          <p className="mt-1 font-mono text-[10px] text-text-muted">
            Objects stored in 2+ locations
          </p>
        </div>
      </div>

      {/* Server info */}
      <div className="mb-6 rounded-xl border border-border bg-abyss">
        <div className="flex items-center gap-2 border-border border-b px-5 py-3">
          <span className="icon-[lucide--monitor] text-sm text-text-muted" />
          <span className="font-display font-semibold text-sm text-text-primary">Server Info</span>
        </div>
        <div className="grid gap-px bg-border sm:grid-cols-3">
          <div className="bg-abyss px-5 py-4">
            <span className="font-mono text-[10px] text-text-muted uppercase tracking-wider">
              Version
            </span>
            <p className="mt-1 font-mono text-text-secondary text-xs">{version || "-"}</p>
          </div>
          <div className="bg-abyss px-5 py-4">
            <span className="font-mono text-[10px] text-text-muted uppercase tracking-wider">
              Platform
            </span>
            <p className="mt-1 font-mono text-text-secondary text-xs">{platform || "-"}</p>
          </div>
          <div className="bg-abyss px-5 py-4">
            <span className="font-mono text-[10px] text-text-muted uppercase tracking-wider">
              Runtime
            </span>
            <p className="mt-1 font-mono text-text-secondary text-xs">{runtime || "-"}</p>
          </div>
        </div>
      </div>

      {/* Distribution note */}
      <div className="rounded-xl border border-border bg-abyss p-5">
        <div className="flex items-start gap-3">
          <span className="icon-[lucide--globe] mt-0.5 shrink-0 text-accent text-base" />
          <div>
            <h3 className="mb-1 font-display font-semibold text-sm text-text-primary">
              Distributed Storage
            </h3>
            <p className="font-body text-text-muted text-xs leading-relaxed">
              Objects are automatically sharded and replicated across connected nodes. Each object
              is stored in at least 2 global locations. Node capacity varies, the cluster aggregates
              all available storage and balances data transparently.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
