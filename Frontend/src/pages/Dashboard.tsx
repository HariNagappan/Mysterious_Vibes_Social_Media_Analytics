import { useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import {
  ArrowRight,
  FileText,
  Smile,
  Sparkles,
  TrendingUp,
  Users,
} from "lucide-react";

import { SentimentPanel, TrendsPanel } from "@/components/analytics/panels";
import { ErrorState, Loader } from "@/components/common/States";
import {
  LiveFeed,
  MetricCard,
  SectionHeading,
} from "@/components/dashboard/Widgets";
import { Badge } from "@/components/ui/badge";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useLiveFeed } from "@/hooks/useLiveFeed";
import { useNetwork, useOverview, useReport } from "@/hooks/useIntelligence";
import { useUiStore } from "@/stores/ui";
import { compactNumber } from "@/utils/format";

export function DashboardPage() {
  const { projectId = "" } = useParams();
  const range = useUiStore((state) => state.range);
  const search = useUiStore((state) => state.search);

  const overview = useOverview(projectId, range);
  const report = useReport(projectId);
  const network = useNetwork(projectId);
  const { posts, connected, isLoading: feedLoading } = useLiveFeed(projectId, 30);

  const filteredFeed = useMemo(() => {
    const query = search.trim().toLowerCase();
    if (!query) return posts;
    return posts.filter((post) =>
      `${post.content} ${post.author} ${post.authorHandle}`
        .toLowerCase()
        .includes(query),
    );
  }, [posts, search]);

  const metrics = overview.data;

  return (
    <div className="space-y-6">
      {overview.isPending || !metrics ? (
        overview.isError ? (
          <ErrorState
            description="Overview metrics are unavailable."
            onRetry={() => void overview.refetch()}
          />
        ) : (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton key={index} className="h-[92px]" />
            ))}
          </div>
        )
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <MetricCard
            label="Total posts"
            value={compactNumber(metrics.totalPosts)}
            delta={metrics.postsDelta}
            hint={`vs previous ${range}`}
            icon={FileText}
          />
          <MetricCard
            label="Active users"
            value={compactNumber(metrics.activeUsers)}
            delta={metrics.usersDelta}
            hint="distinct authors"
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
            hint={`velocity score ${Math.round(metrics.velocityScore)}`}
            icon={TrendingUp}
          />
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Card className="xl:col-span-2">
          <CardHeader>
            <div>
              <p className="label-caps mb-1">Realtime</p>
              <CardTitle>Live activity feed</CardTitle>
            </div>
            <Badge
              className={connected ? "border-pos/30 text-pos" : undefined}
            >
              {connected ? "streaming" : "offline"}
            </Badge>
          </CardHeader>
          <div className="max-h-[430px] overflow-y-auto px-4 py-2">
            {feedLoading ? (
              <Loader label="Warming up the stream..." />
            ) : filteredFeed.length === 0 ? (
              <p className="py-10 text-center text-xs text-faint">
                No posts match the current search.
              </p>
            ) : (
              <LiveFeed posts={filteredFeed} connected={connected} />
            )}
          </div>
        </Card>

        <div className="space-y-4">
          <Card>
            <CardHeader>
              <div>
                <p className="label-caps mb-1">Grounded AI report</p>
                <CardTitle>Executive summary</CardTitle>
              </div>
              <Sparkles className="h-3.5 w-3.5 text-accent" />
            </CardHeader>
            <div className="px-4 py-3">
              {report.isPending ? (
                <Skeleton className="h-20 w-full" />
              ) : report.data ? (
                <>
                  <p className="line-clamp-4 text-xs leading-relaxed text-muted">
                    {report.data.executiveSummary}
                  </p>
                  <Link
                    to={`/projects/${projectId}/reports`}
                    className="mt-3 inline-flex items-center gap-1.5 text-xs text-accent hover:underline"
                  >
                    Read the full report
                    <ArrowRight className="h-3 w-3" />
                  </Link>
                </>
              ) : (
                <p className="text-xs text-muted">
                  Report unavailable right now.
                </p>
              )}
            </div>
          </Card>

          <Card>
            <CardHeader>
              <div>
                <p className="label-caps mb-1">Network overview</p>
                <CardTitle>Top influencers</CardTitle>
              </div>
              <Badge>
                {network.data
                  ? `${network.data.communities.length} communities`
                  : "—"}
              </Badge>
            </CardHeader>
            <div className="px-4 py-2">
              {network.isPending ? (
                <div className="space-y-2 py-2">
                  {Array.from({ length: 4 }).map((_, index) => (
                    <Skeleton key={index} className="h-7 w-full" />
                  ))}
                </div>
              ) : (
                (network.data?.influencers ?? [])
                  .slice(0, 4)
                  .map((influencer, index) => (
                    <div
                      key={influencer.reference}
                      className="flex items-center gap-3 border-b border-line/60 py-2 last:border-0"
                    >
                      <span className="w-4 font-mono text-[10px] text-faint">
                        {index + 1}
                      </span>
                      <span className="min-w-0 flex-1 truncate text-xs text-ink">
                        {influencer.displayName}
                      </span>
                      <span className="font-mono text-[10px] text-accent">
                        {influencer.influence.toFixed(2)}
                      </span>
                    </div>
                  ))
              )}
              <Link
                to={`/projects/${projectId}/network`}
                className="mt-2 inline-flex items-center gap-1.5 pb-2 text-xs text-accent hover:underline"
              >
                Open the network view
                <ArrowRight className="h-3 w-3" />
              </Link>
            </div>
          </Card>
        </div>
      </div>

      <section>
        <SectionHeading
          kicker="Emotion & stance"
          title="Sentiment overview"
          action={
            <Link
              to={`/projects/${projectId}/sentiment`}
              className="inline-flex items-center gap-1.5 text-xs text-accent hover:underline"
            >
              Open sentiment view
              <ArrowRight className="h-3 w-3" />
            </Link>
          }
        />
        <SentimentPanel topicId={projectId} range={range} compact />
      </section>

      <section>
        <SectionHeading
          kicker="Rising narratives"
          title="Trend overview"
          action={
            <Link
              to={`/projects/${projectId}/trends`}
              className="inline-flex items-center gap-1.5 text-xs text-accent hover:underline"
            >
              Open trend view
              <ArrowRight className="h-3 w-3" />
            </Link>
          }
        />
        <TrendsPanel topicId={projectId} range={range} compact />
      </section>
    </div>
  );
}
