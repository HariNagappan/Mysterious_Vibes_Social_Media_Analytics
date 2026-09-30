import type { HTMLAttributes } from "react";

import { cn } from "@/utils/cn";

export function Badge({ className, ...props }: HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-[4px] border border-line bg-white/5 px-2 py-0.5 font-mono text-[10px] uppercase tracking-[0.08em] text-muted",
        className,
      )}
      {...props}
    />
  );
}

export function StatusBadge({ status }: { status: string }) {
  const palette: Record<string, string> = {
    active: "border-pos/30 bg-pos/10 text-pos",
    paused: "border-warn/30 bg-warn/10 text-warn",
    archived: "border-line bg-white/5 text-muted",
  };
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-[4px] border px-2 py-0.5 font-mono text-[10px] uppercase tracking-[0.08em]",
        palette[status] ?? palette.archived,
      )}
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current" aria-hidden />
      {status}
    </span>
  );
}
