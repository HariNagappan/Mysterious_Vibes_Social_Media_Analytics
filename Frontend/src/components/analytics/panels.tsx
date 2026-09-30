import { EMOTION_META } from "@/app/constants";
import {
  ChartCard,
  DistributionBarChart,
  EmotionPieChart,
  HeatMap,
  KeywordTrendChart,
  SentimentAreaChart,
} from "@/components/charts/ChartKit";
import { ErrorState, Loader } from "@/components/common/States";
import { TrendCard } from "@/components/dashboard/Widgets";
import { COMMUNITY_COLORS } from "@/components/graphs/NetworkGraph";
import { Badge } from "@/components/ui/badge";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { useAudience, useSentiment, useTrends } from "@/hooks/useIntelligence";
import type { DateRange, NetworkPayload } from "@/types";
import { cn } from "@/utils/cn";
import {
  compactNumber,
  formatDay,
  formatTime,
  growthLabel,
} from "@/utils/format";

// ── Sentiment ───────────────────────────────────────────────────────────────

export function SentimentPanel({
  topicId,
  range,
  compact = false,
}: {
  topicId: string;
  range: DateRange;
  compact?: boolean;
}) {
  const { data, isPending, isError, refetch } = useSentiment(topicId, range);

  if (isPending) return <Loader label="Scoring emotional tone..." />;
  if (isError || !data) {
    return (
      <ErrorState
        description="Sentiment analytics are unavailable."
        onRetry={() => void refetch()}
      />
    );
  }

  const pieSlices = data.distribution.map((item) => ({
    emotion: item.emotion,
    share: item.share,
  }));
  const heatRows = data.heatmap.emotions.map((emotion, rowIndex) => ({
    label: EMOTION_META[emotion].label,
    color: EMOTION_META[emotion].color,
    values: data.heatmap.values[rowIndex] ?? [],
  }));
  const shortRange = range === "24h";
  const heatColumns = data.heatmap.columns.map((column) =>
    shortRange ? formatTime(column) : formatDay(column),
  );

  return (
    <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
      <ChartCard
        title="Emotion distribution"
        kicker={`${data.analyzed} posts scored · overall ${data.overallScore}`}
      >
        <EmotionPieChart slices={pieSlices} />
      </ChartCard>
      <ChartCard title="Sentiment over time" kicker="Average score per window">
        <SentimentAreaChart
          data={data.timeline.map((point) => ({
            bucket: point.bucket,
            avgScore: point.avgScore,
          }))}
        />
      </ChartCard>
      {!compact ? (
        <>
          <Card className="xl:col-span-2">
            <CardHeader>
              <div>
                <p className="label-caps mb-1">Emotion x time</p>
                <CardTitle>Emotion heatmap</CardTitle>
              </div>
              <Badge>{heatRows.length} emotions</Badge>
            </CardHeader>
            <div className="px-4 py-4">
              <HeatMap columns={heatColumns} rows={heatRows} />
            </div>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Emotion breakdown</CardTitle>
            </CardHeader>
            <div className="space-y-3 px-4 py-4">
              {data.distribution.map((item) => {
                const meta = EMOTION_META[item.emotion];
                return (
                  <div key={item.emotion}>
                    <div className="mb-1 flex items-center justify-between text-[11px]">
                      <span className="text-muted">{meta.label}</span>
                      <span className="font-mono text-faint">
                        {item.count} · {Math.round(item.share * 100)}%
                      </span>
                    </div>
                    <div className="h-1.5 rounded-full bg-white/5">
                      <div
                        className="h-full rounded-full"
                        style={{
                          width: `${Math.max(2, item.share * 100)}%`,
                          backgroundColor: meta.color,
                        }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Top emotional keywords</CardTitle>
            </CardHeader>
            <div className="flex flex-wrap gap-2 px-4 py-4">
              {data.keywords.slice(0, 10).map((keyword) => (
                <span
                  key={keyword.keyword}
                  className="inline-flex items-center gap-1.5 rounded-[4px] border border-line bg-white/5 px-2 py-1 font-mono text-[10px] text-muted"
                >
                  {keyword.keyword}
                  <span className="text-faint">{keyword.count}</span>
                </span>
              ))}
            </div>
          </Card>
        </>
      ) : null}
    </div>
  );
}

// ── Trends ──────────────────────────────────────────────────────────────────

export function TrendsPanel({
  topicId,
  range,
  compact = false,
}: {
  topicId: string;
  range: DateRange;
  compact?: boolean;
}) {
  const { data, isPending, isError, refetch } = useTrends(topicId, range);

  if (isPending) return <Loader label="Detecting rising narratives..." />;
  if (isError || !data) {
    return (
      <ErrorState
        description="Trend analytics are unavailable."
        onRetry={() => void refetch()}
      />
    );
  }

  const compare = data.rising.slice(0, 6);
  const maxScore = Math.max(1, ...compare.map((item) => item.score));

  return (
    <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
      <Card>
        <CardHeader>
          <div>
            <p className="label-caps mb-1">
              Overall growth {growthLabel(data.overallGrowth)}
            </p>
            <CardTitle>Rising topics</CardTitle>
          </div>
        </CardHeader>
        <div className="px-4 py-2">
          {data.rising.slice(0, compact ? 5 : 8).map((topic) => (
            <TrendCard key={topic.keyword} topic={topic} />
          ))}
        </div>
      </Card>
      {!compact ? (
        <>
          <ChartCard title="Keyword growth" kicker="Mentions per window" height={300}>
            <KeywordTrendChart series={data.series} />
          </ChartCard>
          <Card>
            <CardHeader>
              <CardTitle>Trend comparison</CardTitle>
              <Badge>trend score</Badge>
            </CardHeader>
            <div className="space-y-3 px-4 py-4">
              {compare.map((item) => (
                <div key={item.keyword}>
                  <div className="mb-1 flex items-center justify-between text-[11px]">
                    <span className="text-muted">#{item.keyword}</span>
                    <span className="font-mono text-faint">
                      {Math.round(item.score)} · {growthLabel(item.growth)}
                    </span>
                  </div>
                  <div className="h-1.5 rounded-full bg-white/5">
                    <div
                      className="h-full rounded-full bg-accent"
                      style={{
                        width: `${Math.max(3, (item.score / maxScore) * 100)}%`,
                      }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Topic clusters</CardTitle>
              <Badge>{data.clusters.length} clusters</Badge>
            </CardHeader>
            <div className="space-y-4 px-4 py-4">
              {data.clusters.map((cluster) => (
                <div key={cluster.label}>
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-medium text-ink">{cluster.label}</span>
                    <span className="font-mono text-[10px] text-faint">
                      {Math.round(cluster.share * 100)}%
                    </span>
                  </div>
                  <div className="mt-1.5 flex flex-wrap gap-1.5">
                    {cluster.keywords.map((keyword) => (
                      <Badge key={keyword}>{keyword}</Badge>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </>
      ) : null}
    </div>
  );
}

// ── Audience ────────────────────────────────────────────────────────────────

export function AudiencePanel({ topicId }: { topicId: string }) {
  const { data, isPending, isError, refetch } = useAudience(topicId);

  if (isPending) return <Loader label="Estimating audience composition..." />;
  if (isError || !data) {
    return (
      <ErrorState
        description="Audience analytics are unavailable."
        onRetry={() => void refetch()}
      />
    );
  }

  const toBars = (record: Record<string, number>, limit = 6) =>
    Object.entries(record)
      .slice(0, limit)
      .map(([name, value]) => ({ name, value }));

  return (
    <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
      <ChartCard
        title="Language distribution"
        kicker={`Sample: ${compactNumber(data.sampleSize)} posts`}
      >
        <DistributionBarChart data={toBars(data.languages)} />
      </ChartCard>
      <ChartCard title="Regional patterns" kicker="Geographic spread">
        <DistributionBarChart data={toBars(data.regions)} color="#60A5FA" />
      </ChartCard>
      <ChartCard
        title="Interest categories"
        kicker="Content signals"
        height={280}
      >
        <DistributionBarChart data={toBars(data.interests)} color="#FBBF24" />
      </ChartCard>
      <Card>
        <CardHeader>
          <div>
            <p className="label-caps mb-1">Aggregated groups only</p>
            <CardTitle>Community segments</CardTitle>
          </div>
        </CardHeader>
        <div className="space-y-3 px-4 py-4">
          {data.segments.map((segment) => (
            <div
              key={segment.label}
              className="border border-line bg-panel2/60 px-3 py-2.5"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-ink">
                  {segment.label}
                </span>
                <span className="font-mono text-[10px] text-accent">
                  {Math.round(segment.share * 100)}%
                </span>
              </div>
              <p className="mt-1 text-[11px] leading-relaxed text-muted">
                {segment.description}
              </p>
            </div>
          ))}
          <p className="pt-1 text-[10px] leading-relaxed text-faint">
            Audience analytics use aggregated groups only - no individual user
            profiling.
          </p>
        </div>
      </Card>
    </div>
  );
}

// ── Network side panel ──────────────────────────────────────────────────────

export function NetworkInsightsSide({
  network,
  selectedId,
  onSelect,
}: {
  network: NetworkPayload;
  selectedId?: string;
  onSelect: (nodeId?: string) => void;
}) {
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Influence ranking</CardTitle>
          <Badge>pagerank</Badge>
        </CardHeader>
        <div className="px-4 py-2">
          {network.influencers.map((influencer, index) => (
            <button
              key={influencer.reference}
              onClick={() =>
                onSelect(
                  influencer.reference === selectedId
                    ? undefined
                    : influencer.reference,
                )
              }
              className={cn(
                "focus-ring flex w-full items-center gap-3 border-b border-line/60 py-2 text-left transition-colors last:border-0",
                selectedId === influencer.reference && "bg-accent/5",
              )}
            >
              <span className="w-5 shrink-0 font-mono text-[10px] text-faint">
                {index + 1}
              </span>
              <span className="min-w-0 flex-1 truncate text-xs text-ink">
                {influencer.displayName}
              </span>
              <span className="font-mono text-[10px] text-accent">
                {influencer.influence.toFixed(2)}
              </span>
            </button>
          ))}
        </div>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Communities</CardTitle>
          <Badge>{network.communities.length}</Badge>
        </CardHeader>
        <div className="space-y-2.5 px-4 py-4">
          {network.communities.map((community) => (
            <div key={community.id} className="flex items-center gap-2">
              <span
                className="h-2.5 w-2.5 shrink-0 rounded-full"
                style={{
                  backgroundColor:
                    COMMUNITY_COLORS[community.id % COMMUNITY_COLORS.length],
                }}
              />
              <span className="min-w-0 flex-1 truncate text-xs text-muted">
                {community.label}
              </span>
              <span className="font-mono text-[10px] text-faint">
                {community.size} accounts
              </span>
            </div>
          ))}
        </div>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Information spread path</CardTitle>
        </CardHeader>
        <div className="px-4 py-4">
          <ol className="space-y-2.5">
            {network.spreadPath.map((step, index) => (
              <li
                key={index}
                className="flex gap-2.5 text-[11px] leading-relaxed text-muted"
              >
                <span className="mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full border border-line bg-panel2 font-mono text-[9px] text-accent">
                  {index + 1}
                </span>
                {step}
              </li>
            ))}
          </ol>
        </div>
      </Card>
    </div>
  );
}
