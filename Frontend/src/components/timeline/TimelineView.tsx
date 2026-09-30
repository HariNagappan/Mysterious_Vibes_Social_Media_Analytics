import { useMemo, useState } from "react";
import { Flag, TrendingUp, Users, Zap } from "lucide-react";

import { DATE_RANGES } from "@/app/constants";
import { ChartCard, GrowthLineChart } from "@/components/charts/ChartKit";
import { ErrorState, Loader } from "@/components/common/States";
import { PlatformBadge, SectionHeading } from "@/components/dashboard/Widgets";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { useTimeline } from "@/hooks/useIntelligence";
import { useUiStore } from "@/stores/ui";
import type { TimelineEventKind } from "@/types";
import { cn } from "@/utils/cn";
import { compactNumber, formatDateTime, formatDay } from "@/utils/format";

const KIND_FILTERS: { value: "all" | TimelineEventKind; label: string }[] = [
  { value: "all", label: "All events" },
  { value: "first-post", label: "First post" },
  { value: "growth", label: "Growth" },
  { value: "spike", label: "Spikes" },
  { value: "community", label: "Communities" },
];

const KIND_ICON: Record<TimelineEventKind, typeof Flag> = {
  "first-post": Flag,
  growth: TrendingUp,
  spike: Zap,
  community: Users,
};

export function TimelineView({ topicId }: { topicId: string }) {
  const range = useUiStore((state) => state.range);
  const setRange = useUiStore((state) => state.setRange);
  const [kindFilter, setKindFilter] = useState<"all" | TimelineEventKind>("all");

  const { data, isPending, isError, refetch } = useTimeline(topicId, range);

  const events = useMemo(() => {
    if (!data) return [];
    return kindFilter === "all"
      ? data.events
      : data.events.filter((event) => event.kind === kindFilter);
  }, [data, kindFilter]);

  if (isPending) return <Loader label="Reconstructing the conversation..." />;
  if (isError || !data) {
    return (
      <ErrorState
        description="The timeline could not be loaded."
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <div>
      <SectionHeading
        kicker="Conversation replay"
        title="Timeline"
        action={
          <div className="flex items-center gap-2">
            <span className="label-caps">Zoom</span>
            <div className="flex overflow-hidden rounded-md border border-line">
              {DATE_RANGES.map((option) => (
                <button
                  key={option.value}
                  onClick={() => setRange(option.value)}
                  className={cn(
                    "focus-ring px-3 py-1.5 font-mono text-[10px] uppercase tracking-wider transition-colors",
                    range === option.value
                      ? "bg-accent/15 text-accent"
                      : "text-muted hover:bg-white/5",
                  )}
                >
                  {option.value}
                </button>
              ))}
            </div>
          </div>
        }
      />

      <div className="mb-4 grid grid-cols-2 gap-3 md:grid-cols-4">
        <Card className="px-4 py-3">
          <p className="label-caps">Posts</p>
          <p className="mt-1 font-mono text-xl text-ink">
            {compactNumber(data.totals.posts)}
          </p>
        </Card>
        <Card className="px-4 py-3">
          <p className="label-caps">Engagements</p>
          <p className="mt-1 font-mono text-xl text-ink">
            {compactNumber(data.totals.engagements)}
          </p>
        </Card>
        <Card className="px-4 py-3">
          <p className="label-caps">Accounts</p>
          <p className="mt-1 font-mono text-xl text-ink">
            {compactNumber(data.totals.authors)}
          </p>
        </Card>
        <Card className="px-4 py-3">
          <p className="label-caps">Events</p>
          <p className="mt-1 font-mono text-xl text-ink">{data.events.length}</p>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <ChartCard
          title="Discussion growth"
          kicker="Posts per window"
          className="xl:col-span-2"
          height={280}
        >
          <GrowthLineChart
            data={data.growth.map((bucket) => ({
              bucket: bucket.bucket,
              posts: bucket.posts,
            }))}
          />
        </ChartCard>

        <Card>
          <CardHeader>
            <CardTitle>Filters</CardTitle>
          </CardHeader>
          <div className="flex flex-wrap gap-2 px-4 py-4">
            {KIND_FILTERS.map((filter) => (
              <Button
                key={filter.value}
                size="sm"
                variant={kindFilter === filter.value ? "primary" : "secondary"}
                onClick={() => setKindFilter(filter.value)}
              >
                {filter.label}
              </Button>
            ))}
          </div>
          <div className="border-t border-line px-4 py-3">
            <p className="text-[11px] leading-relaxed text-muted">
              Zoom alters the bucket size: hourly for 24h windows, 4-hourly for
              7 days, 12-hourly for 14 days.
            </p>
          </div>
        </Card>
      </div>

      <Card className="mt-4">
        <CardHeader>
          <CardTitle>Event stream</CardTitle>
          <Badge>{events.length} shown</Badge>
        </CardHeader>
        <div className="px-4 py-4">
          <ol className="relative border-l border-line pl-6">
            {events.map((event) => {
              const Icon = KIND_ICON[event.kind];
              return (
                <li key={event.id} className="relative pb-6 last:pb-0">
                  <span className="absolute -left-[31px] top-0.5 flex h-6 w-6 items-center justify-center rounded-full border border-line bg-panel2">
                    <Icon className="h-3 w-3 text-accent" />
                  </span>
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-mono text-[10px] text-faint">
                      {formatDateTime(event.timestamp)} · {formatDay(event.timestamp)}
                    </span>
                    {event.platform ? (
                      <PlatformBadge platform={event.platform} />
                    ) : null}
                    {event.engagement ? (
                      <span className="font-mono text-[10px] text-faint">
                        {compactNumber(event.engagement)} eng
                      </span>
                    ) : null}
                  </div>
                  <p className="mt-1 text-xs font-medium text-ink">
                    {event.title}
                  </p>
                  <p className="mt-0.5 text-xs leading-relaxed text-muted">
                    {event.detail}
                  </p>
                </li>
              );
            })}
          </ol>
        </div>
      </Card>
    </div>
  );
}
