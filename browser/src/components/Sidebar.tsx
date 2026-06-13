"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";
import { deleteBucketAction, logoutAction } from "@/lib/actions";
import { safeDisplayName } from "@/lib/safe-name";
import { BucketModal } from "./BucketModal";

interface Props {
  buckets: { name: string; creationDate: string }[];
  storageUsed: string;
  _storageBytes: number;
  bucketCount: number;
  serverVersion: string;
  serverPlatform: string;
}

export function Sidebar({
  buckets,
  storageUsed,
  _storageBytes,
  bucketCount,
  serverVersion,
  serverPlatform,
}: Props) {
  const pathname = usePathname();
  const router = useRouter();
  const activeBucket = decodeURIComponent(pathname.split("/")[1] || "");
  const [filter, setFilter] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [editBucket, setEditBucket] = useState<string | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  const filtered = buckets.filter((b) => b.name.toLowerCase().includes(filter.toLowerCase()));

  const handleModalSuccess = () => {
    router.refresh();
  };

  const openCreate = () => {
    setEditBucket(null);
    setModalOpen(true);
  };

  const openEdit = (bucketName: string, e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setEditBucket(bucketName);
    setModalOpen(true);
  };

  const handleDelete = async (bucketName: string) => {
    await deleteBucketAction(bucketName);
    setDeleteConfirm(null);
    router.refresh();
  };

  return (
    <>
      <aside className="flex h-full w-64 shrink-0 flex-col border-border border-r bg-abyss">
        {/* Logo */}
        <Link
          href="/"
          className="flex items-center gap-2.5 border-border border-b px-4 py-4 transition-colors hover:bg-surface/30"
        >
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent">
            <span className="icon-[fluent-emoji-high-contrast--lobster] text-black text-lg" />
          </div>
          <div>
            <p className="font-display font-semibold text-sm leading-tight">Obstor</p>
            <p className="font-mono text-[10px] text-text-muted">{storageUsed} used</p>
          </div>
        </Link>

        {/* Search + create */}
        <div className="border-border border-b px-3 py-3">
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <span className="icon-[lucide--search] pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2 text-text-muted text-xs" />
              <input
                type="text"
                placeholder="Filter buckets..."
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                className="w-full rounded-md border border-border bg-surface py-1.5 pr-3 pl-8 font-mono text-xs outline-none transition-colors placeholder:text-text-muted focus:border-accent"
              />
            </div>
            <button
              type="button"
              onClick={openCreate}
              className="flex h-[30px] w-[30px] shrink-0 items-center justify-center rounded-md bg-accent text-black transition-colors hover:bg-accent-bright"
              title="Create bucket"
            >
              <span className="icon-[lucide--plus] text-sm" />
            </button>
          </div>
        </div>

        {/* Bucket list */}
        <nav className="flex-1 overflow-y-auto py-1">
          {filtered.map((b) => {
            const isActive = activeBucket === b.name;
            return (
              <div
                key={b.name}
                className={`group relative mx-1.5 mb-0.5 rounded-md ${
                  isActive ? "bg-surface" : "hover:bg-surface/50"
                }`}
              >
                <Link
                  href={`/${encodeURIComponent(b.name)}`}
                  className="flex items-center gap-2.5 px-3 py-2"
                >
                  <span
                    className={`icon-[lucide--hard-drive] text-xs ${
                      isActive ? "text-accent" : "text-text-muted"
                    }`}
                  />
                  <span
                    className={`truncate font-mono text-xs ${
                      isActive ? "text-text-primary" : "text-text-secondary"
                    }`}
                  >
                    <bdi>{safeDisplayName(b.name)}</bdi>
                  </span>
                </Link>

                {/* Action buttons (visible on hover or when active) */}
                <div
                  className={`absolute top-1/2 right-1.5 flex -translate-y-1/2 items-center gap-0.5 ${
                    isActive ? "opacity-100" : "opacity-0 group-hover:opacity-100"
                  } transition-opacity`}
                >
                  <button
                    type="button"
                    onClick={(e) => openEdit(b.name, e)}
                    className="flex h-6 w-6 items-center justify-center rounded text-text-muted transition-colors hover:bg-surface-overlay hover:text-accent"
                    title="Bucket settings"
                  >
                    <span className="icon-[lucide--settings] text-[11px]" />
                  </button>
                  <button
                    type="button"
                    onClick={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      setDeleteConfirm(b.name);
                    }}
                    className="flex h-6 w-6 items-center justify-center rounded text-text-muted transition-colors hover:bg-danger/10 hover:text-danger"
                    title="Delete bucket"
                  >
                    <span className="icon-[lucide--trash-2] text-[11px]" />
                  </button>
                </div>
              </div>
            );
          })}

          {filtered.length === 0 && (
            <p className="px-4 py-6 text-center font-body text-text-muted text-xs">
              {filter ? "No matching buckets" : "No buckets yet"}
            </p>
          )}
        </nav>

        {/* Footer */}
        <div className="border-border border-t px-4 py-3">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-mono text-[10px] text-text-muted">
                {bucketCount} bucket{bucketCount !== 1 ? "s" : ""} | v{serverVersion}
              </p>
              <p className="font-mono text-[10px] text-text-muted">{serverPlatform}</p>
            </div>
            <button
              type="button"
              onClick={() => logoutAction()}
              className="flex h-7 w-7 items-center justify-center rounded-md text-text-muted transition-colors hover:bg-surface hover:text-danger"
              title="Sign out"
            >
              <span className="icon-[lucide--log-out] text-xs" />
            </button>
          </div>
        </div>
      </aside>

      {/* Bucket Create/Edit Modal */}
      <BucketModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSuccess={handleModalSuccess}
        editBucket={editBucket}
      />

      {/* Delete Confirmation */}
      {deleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <button
            type="button"
            aria-label="Close dialog"
            className="absolute inset-0 bg-void/80 backdrop-blur-sm"
            onClick={() => setDeleteConfirm(null)}
          />
          <div className="relative w-full max-w-sm rounded-xl border border-border bg-abyss p-6 shadow-2xl shadow-black/50">
            <div className="mb-4 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-danger/10">
                <span className="icon-[lucide--alert-triangle] text-base text-danger" />
              </div>
              <div>
                <h3 className="font-display font-semibold text-sm">Delete Bucket</h3>
                <p className="font-mono text-[11px] text-text-muted">
                  <bdi>{safeDisplayName(deleteConfirm)}</bdi>
                </p>
              </div>
            </div>
            <p className="mb-5 font-body text-text-secondary text-xs leading-relaxed">
              This will permanently delete the bucket and all objects inside it. This action cannot
              be undone.
            </p>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => handleDelete(deleteConfirm)}
                className="flex-1 rounded-lg bg-danger px-4 py-2 font-body font-medium text-sm text-white transition-colors hover:bg-danger/90"
              >
                Delete
              </button>
              <button
                type="button"
                onClick={() => setDeleteConfirm(null)}
                className="rounded-lg border border-border px-4 py-2 font-body text-sm text-text-muted transition-colors hover:bg-surface-overlay"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
