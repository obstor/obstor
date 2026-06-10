"use client";

import { usePathname } from "next/navigation";

export function Header() {
  const pathname = usePathname();
  const parts = pathname.split("/").filter(Boolean);
  const bucket = parts[0] ? decodeURIComponent(parts[0]) : null;

  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-border border-b bg-abyss px-6">
      <div className="flex items-center gap-2 font-mono text-sm text-text-secondary">
        {bucket ? (
          <>
            <span className="icon-[lucide--hard-drive] text-text-muted" />
            <span className="text-text-primary">{bucket}</span>
          </>
        ) : null}
      </div>

      <div className="flex items-center gap-2">
        <a
          href="https://github.com/obstor/obstor"
          target="_blank"
          rel="noopener noreferrer"
          className="flex h-8 w-8 items-center justify-center rounded-md text-text-muted transition-colors hover:bg-surface hover:text-text-secondary"
          title="GitHub"
        >
          <span className="icon-[lucide--github] text-sm" />
        </a>
        <a
          href="https://obstor.net/docs"
          target="_blank"
          rel="noopener noreferrer"
          className="flex h-8 w-8 items-center justify-center rounded-md text-text-muted transition-colors hover:bg-surface hover:text-text-secondary"
          title="Documentation"
        >
          <span className="icon-[lucide--book-open] text-sm" />
        </a>
      </div>
    </header>
  );
}
