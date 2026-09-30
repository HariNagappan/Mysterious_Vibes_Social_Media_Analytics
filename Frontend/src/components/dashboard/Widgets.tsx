import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";

import { EMOTION_META, PLATFORM_META } from "@/app/constants";
import { withAlpha, Sparkline } from "@/components/charts/ChartKit";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import type { Emotion, Platform, Post, TrendTopic } from "@/types";
import { cn } from "@/utils/cn";
import { compactNumber, growthLabel, timeAgo } from "@/utils/format";

export function MetricCard({
  label,
  value,
  delta,
  hint,
  icon: Icon,
}: {
  label: string;
  value: string;
  delta?: number;
  hint?: string;
  icon?: LucideIcon;
}) {
  return (
    <Card className="px-4 py-3">
      <div className="flex items-start justify-between">
        <p className="label-caps">{label}</p>
        {Icon ? <Icon className="h-4 w-4 text-faint" /> : null}
      </div>
      <p className="mt-2 font-mono text-2xl tracking-tight text-ink">{value}</p>
      <div className="mt-1 flex items-center gap-2 text-[11px]">
        {typeof delta === "number" ? (
          <span
            className={cn(
              "font-mono",
              delta >= 0 ? "text-pos" : "text-neg",
            )}
          >
            {delta >= 0 ? "+" : ""}
            {Math.round(delta * 100)}%
          </span>
        ) : null}
        {hint ? <span className="text-muted">{hint}</span> : null}
      </div>
    </Card>
  );
}

export function EmotionChip({ emotion }: { emotion: Emotion }) {
  const meta = EMOTION_META[emotion];
  return (
    <span
      className="inline-flex items-center rounded-[4px] border px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide"
      style={{
        borderColor: withAlpha(meta.color, 0.35),
        backgroundColor: withAlpha(meta.color, 0.12),
        color: meta.color,
      }}
    >
      {meta.label}
    </span>
  );
}

export function PlatformBadge({ platform }: { platform: Platform }) {
  const meta = PLATFORM_META[platform];
  return (
    <span
      title={meta.label}
      className="inline-flex h-6 min-w-[26px] items-center justify-center rounded-[4px] border px-1 font-mono text-[10px]"
      style={{
        borderColor: withAlpha(meta.color, 0.35),
        backgroundColor: withAlpha(meta.color, 0.1),
        color: meta.color,
      }}
    >
      {meta.short}
    </span>
  );
}

export function LiveFeed({
  posts,
  connected,
  className,
}: {
  posts: Post[];
  connected: boolean;
  className?: string;
}) {
  return (
    <div className={cn("flex flex-col", className)}>
      {posts.map((post) => (
        <article
          key={post.id}
          className="flex gap-3 border-b border-line/60 px-1 py-2.5 last:border-0"
        >
          <div className="pt-0.5">
            <PlatformBadge platform={post.platform} />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <span className="truncate text-xs font-medium text-ink">
                {post.author}
              </span>
              <span className="shrink-0 font-mono text-[10px] text-faint">
                {timeAgo(post.timestamp)}
              </span>
            </div>
            <p className="mt-0.5 line-clamp-2 text-xs leading-relaxed text-muted">
              {post.content}
            </p>
            <div className="mt-1.5 flex items-center gap-2">
              <EmotionChip emotion={post.emotion} />
              <span className="font-mono text-[10px] text-faint">
                {compactNumber(post.engagement)} eng
              </span>
            </div>
          </div>
        </article>
      ))}
      {connected ? (
        <div className="mt-2 flex items-center gap-2 px-1 text-[10px] text-faint">
          <span className="h-1.5 w-1.5 animate-pulse-dot rounded-full bg-pos" />
          live stream connected
        </div>
      ) : null}
    </div>
  );
}

export function TrendCard({ topic }: { topic: TrendTopic }) {
  return (
    <div className="flex items-center gap-3 border-b border-line/60 py-2.5 last:border-0">
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate text-xs font-medium text-ink">
            #{topic.keyword}
          </span>
          <Badge
            className={
              topic.velocity === "high"
                ? "border-pos/30 text-pos"
                : topic.velocity === "medium"
                  ? "border-warn/30 text-warn"
                  : undefined
            }
          >
            {topic.velocity}
          </Badge>
        </div>
        <p className="mt-0.5 text-[11px] text-muted">
          {topic.count} mentions · {growthLabel(topic.growth)} growth
        </p>
      </div>
      <div className="w-24 shrink-0">
        <Sparkline values={topic.sparkline} />
      </div>
      <span className="w-10 shrink-0 text-right font-mono text-xs text-accent">
        {Math.round(topic.score)}
      </span>
    </div>
  );
}

export function InsightCard({
  title,
  body,
  icon: Icon,
}: {
  title: string;
  body: string;
  icon?: LucideIcon;
}) {
  return (
    <div className="border border-line bg-panel2/60 px-4 py-3">
      <div className="flex items-center gap-2">
        {Icon ? <Icon className="h-3.5 w-3.5 text-accent" /> : null}
        <p className="text-xs font-medium text-ink">{title}</p>
      </div>
      <p className="mt-1.5 text-xs leading-relaxed text-muted">{body}</p>
    </div>
  );
}

export function SectionHeading({
  kicker,
  title,
  action,
  className,
}: {
  kicker?: string;
  title: string;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "mb-4 flex flex-wrap items-end justify-between gap-3",
        className,
      )}
    >
      <div>
        {kicker ? <p className="label-caps mb-1.5">{kicker}</p> : null}
        <h2 className="text-lg font-medium text-ink">{title}</h2>
      </div>
      {action}
    </div>
  );
}
