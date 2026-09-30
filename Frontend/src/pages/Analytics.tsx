import { useParams } from "react-router-dom";
import { FileText, Smile, TrendingUp, Users } from "lucide-react";

import { SentimentPanel, TrendsPanel } from "@/components/analytics/panels";
import { ChartCard, GrowthLineChart } from "@/components/charts/ChartKit";
import { MetricCard, SectionHeading } from "@/components/dashboard/Widgets";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useOverview, useTimeline } from "@/hooks/useIntelligence";
import { useUiStore } from "@/stores/ui";
import { compactNumber } from "@/utils/format";

export function AnalyticsPage() {
  const { projectId = "" } = useParams();
  const range = useUiStore((state) => state.range);
  const overview = useOverview(projectId, range);
  const timeline = useTimeline(projectId, range);
  const metrics = overview.data;

  return (
    <div className="space-y-6">
      <SectionHeading
        kicker="Combined view"
        title="Analytics"
        action={
          metrics ? (
            <Badge>
              {compactNumber(metrics.totalPosts)} posts ·{" "}
              {compactNumber(metrics.engagementTotal)} engagements
            </Badge>
          ) : null
        }
      />

      {metrics ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <MetricCard
            label="Total posts"
            value={compactNumber(metrics.totalPosts)}
            delta={metrics.postsDelta}
            icon={FileText}
          />
          <MetricCard
            label="Active users"
            value={compactNumber(metrics.activeUsers)}
            delta={metrics.usersDelta}
            icon={Users}
          />
          <MetricCard
            label="Current sentiment"
            value={metrics.avgSentiment.toFixed(2)}
            hint={metrics.sentimentLabel}
            icon={Smile}
          />
          <MetricCard
            label="Trend velocity"
            value={metrics.trendVelocity.toUpperCase()}
            hint={`score ${Math.round(metrics.velocityScore)}`}
            icon={TrendingUp}
          />
        </div>
      ) : (
        <Skeleton className="h-[92px]" />
      )}

      <ChartCard title="Discussion growth" kicker="Posts per window" height={280}>
        {timeline.data ? (
          <GrowthLineChart
            data={timeline.data.growth.map((bucket) => ({
              bucket: bucket.bucket,
              posts: bucket.posts,
            }))}
          />
        ) : (
          <div className="flex h-full items-center justify-center text-xs text-faint">
            Loading series...
          </div>
        )}
      </ChartCard>

      <SentimentPanel topicId={projectId} range={range} />
      <TrendsPanel topicId={projectId} range={range} />
    </div>
  );
}
