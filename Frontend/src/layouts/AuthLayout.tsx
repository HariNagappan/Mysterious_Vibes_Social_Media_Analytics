import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Radar } from "lucide-react";

import { APP_NAME, APP_TAGLINE } from "@/app/constants";

export function AuthLayout({
  title,
  subtitle,
  children,
  footer,
}: {
  title: string;
  subtitle: string;
  children: ReactNode;
  footer: ReactNode;
}) {
  return (
    <div className="grid min-h-screen grid-cols-1 lg:grid-cols-2">
      <div className="hidden flex-col justify-between border-r border-line bg-panel/40 p-10 lg:flex">
        <Link to="/" className="focus-ring flex w-fit items-center gap-2.5 rounded-md">
          <span className="flex h-8 w-8 items-center justify-center rounded-md bg-accent/15">
            <Radar className="h-4 w-4 text-accent" />
          </span>
          <span className="text-sm font-medium text-ink">{APP_NAME}</span>
        </Link>
        <div className="max-w-md">
          <p className="label-caps mb-3">{APP_TAGLINE}</p>
          <h2 className="text-2xl leading-snug text-ink">
            Turn fragmented conversations into decisions.
          </h2>
          <ul className="mt-6 space-y-2 text-sm text-muted">
            <li>Sentiment and emotion intelligence</li>
            <li>Audience and language composition</li>
            <li>Trend velocity and narrative clusters</li>
            <li>Influence networks and propagation</li>
          </ul>
        </div>
        <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-faint">
          SIH 2026 - Problem Statement 26152
        </p>
      </div>
      <div className="flex items-center justify-center p-6">
        <div className="w-full max-w-sm">
          <h1 className="text-xl font-medium text-ink">{title}</h1>
          <p className="mt-1 text-xs text-muted">{subtitle}</p>
          <div className="mt-6">{children}</div>
          <div className="mt-6 text-xs text-muted">{footer}</div>
        </div>
      </div>
    </div>
  );
}
