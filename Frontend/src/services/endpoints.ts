/**
 * Typed API clients for every PulseGraph domain.
 *
 * Each client talks to the real backend when VITE_API_URL is configured and
 * transparently falls back to the deterministic mock engine otherwise, so the
 * UI is fully functional in both modes. Real-mode mappings are best-effort
 * adapters over the backend's REST shapes.
 */

import { http, isMockMode } from "@/services/api";
import { mock } from "@/services/mock/engine";
import type {
  AudiencePayload,
  AuthResponse,
  CreateProjectInput,
  DateRange,
  Post,
  NetworkPayload,
  OverviewPayload,
  Project,
  Report,
  SentimentPayload,
  TimelineEvent,
  TimelinePayload,
  Topic,
  TrendPayload,
  TrendVelocity,
} from "@/types";

const mockDelay = () =>
  new Promise<void>((resolve) => window.setTimeout(resolve, 140 + Math.random() * 200));

async function maybeMock<T>(producer: () => T): Promise<T> {
  await mockDelay();
  return producer();
}

function velocityFrom(value: number): TrendVelocity {
  if (value >= 0.9) return "high";
  if (value >= 0.25) return "medium";
  return "low";
}

export const authAPI = {
  async login(input: { email: string; password: string }): Promise<AuthResponse> {
    if (isMockMode()) return maybeMock(() => mock.login(input));
    const { data } = await http.post("/auth/login", input);
    const payload = data as {
      token: string;
      expires_at?: string;
      expiresAt?: string;
      user: AuthResponse["user"];
    };
    return {
      token: payload.token,
      expiresAt: payload.expires_at ?? payload.expiresAt ?? new Date().toISOString(),
      user: payload.user,
    };
  },

  async register(input: {
    name: string;
    email: string;
    password: string;
  }): Promise<AuthResponse> {
    if (isMockMode()) return maybeMock(() => mock.register(input));
    const { data } = await http.post("/auth/register", input);
    const payload = data as {
      token: string;
      expires_at?: string;
      expiresAt?: string;
      user: AuthResponse["user"];
    };
    return {
      token: payload.token,
      expiresAt: payload.expires_at ?? payload.expiresAt ?? new Date().toISOString(),
      user: payload.user,
    };
  },
};

interface BackendTopic {
  id: number | string;
  name?: string;
  description?: string;
  keywords?: string[];
  created_at?: string;
  updated_at?: string;
}

function mapTopicToProject(topic: BackendTopic): Project {
  const id = String(topic.id);
  return {
    id,
    name: topic.name ?? "Untitled project",
    description: topic.description ?? "",
    platforms: ["twitter", "telegram"],
    keywords: topic.keywords ?? [],
    frequency: "hourly",
    status: "active",
    topicIds: [id],
    lastUpdated: topic.updated_at ?? new Date().toISOString(),
    createdAt: topic.created_at ?? new Date().toISOString(),
  };
}

export const projectAPI = {
  async list(): Promise<Project[]> {
    if (isMockMode()) return maybeMock(() => mock.listProjects());
    const { data } = await http.get<{ topics: BackendTopic[] }>("/topics");
    return (data.topics ?? []).map(mapTopicToProject);
  },

  async get(id: string): Promise<Project> {
    if (isMockMode()) return maybeMock(() => mock.getProject(id));
    const { data } = await http.get<BackendTopic>(`/topics/${id}`);
    return mapTopicToProject(data);
  },

  async create(input: CreateProjectInput): Promise<Project> {
    if (isMockMode()) return maybeMock(() => mock.createProject(input));
    const { data } = await http.post<BackendTopic>("/topics", {
      name: input.name,
      description: input.description,
      keywords: input.keywords,
    });
    return mapTopicToProject(data);
  },
};

export const topicAPI = {
  async list(): Promise<Topic[]> {
    if (isMockMode()) return maybeMock(() => mock.listTopics());
    const { data } = await http.get<{ topics: BackendTopic[] }>("/topics");
    return (data.topics ?? []).map((topic) => ({
      id: String(topic.id),
      projectId: String(topic.id),
      name: topic.name ?? "Untitled topic",
      keywords: topic.keywords ?? [],
      description: topic.description ?? "",
      scenario: "event" as const,
    }));
  },

  async get(id: string): Promise<Topic> {
    if (isMockMode()) return maybeMock(() => mock.getTopic(id));
    const { data } = await http.get<BackendTopic>(`/topics/${id}`);
    return {
      id: String(data.id),
      projectId: String(data.id),
      name: data.name ?? "Untitled topic",
      keywords: data.keywords ?? [],
      description: data.description ?? "",
      scenario: "event" as const,
    };
  },
};

interface BackendPost {
  id: number | string;
  platform?: string;
  author_reference?: string;
  content?: string;
  language?: string;
  region_hint?: string;
  timestamp?: string;
  engagement_count?: number;
}

function mapPost(post: BackendPost, topicId: string): Post {
  return {
    id: String(post.id),
    topicId,
    platform: (post.platform as Post["platform"]) ?? "twitter",
    author: post.author_reference ?? "unknown",
    authorHandle: post.author_reference ?? "unknown",
    content: post.content ?? "",
    language: post.language ?? "en",
    region: post.region_hint ?? "Unknown",
    timestamp: post.timestamp ?? new Date().toISOString(),
    engagement: post.engagement_count ?? 0,
    emotion: "neutral",
    sentimentScore: 0,
    confidence: 0.5,
  };
}

export const postsAPI = {
  async list(
    topicId: string,
    options?: { limit?: number; offset?: number; search?: string; range?: DateRange },
  ): Promise<{ items: Post[]; total: number }> {
    if (isMockMode()) return maybeMock(() => mock.getPosts(topicId, options));
    const { data } = await http.get<{ posts: BackendPost[] }>(`/topics/${topicId}/timeline`, {
      params: { limit: options?.limit ?? 40, offset: options?.offset ?? 0 },
    });
    const items = (data.posts ?? []).map((post) => mapPost(post, topicId));
    return { items, total: items.length };
  },
};

export const timelineAPI = {
  async forTopic(topicId: string, range: DateRange): Promise<TimelinePayload> {
    if (isMockMode()) return maybeMock(() => mock.getTimeline(topicId, range));
    const { data } = await http.get(`/topics/${topicId}/timeline`);
    const payload = data as {
      growth?: { bucket: string; post_count: number }[];
      important_events?: { post_id: number | string; author?: string; content?: string; timestamp?: string; engagement_count?: number }[];
      posts?: BackendPost[];
    };
    const events: TimelineEvent[] = (payload.important_events ?? []).map((event, index) => ({
      id: `ev-${index}`,
      timestamp: event.timestamp ?? new Date().toISOString(),
      kind: "spike",
      title: "Viral activity spike",
      detail: `${event.author ?? "unknown"} drove ${event.engagement_count ?? 0} engagements.`,
    }));
    return {
      topicId,
      range,
      growth: (payload.growth ?? []).map((bucket) => ({ bucket: bucket.bucket, posts: bucket.post_count })),
      events,
      totals: {
        posts: (payload.posts ?? []).length,
        engagements: (payload.posts ?? []).reduce((sum, post) => sum + (post.engagement_count ?? 0), 0),
        authors: new Set((payload.posts ?? []).map((post) => post.author_reference)).size,
      },
    };
  },
};

export const sentimentAPI = {
  async forTopic(topicId: string, range: DateRange): Promise<SentimentPayload> {
    if (isMockMode()) return maybeMock(() => mock.getSentiment(topicId, range));
    const { data } = await http.get(`/topics/${topicId}/sentiment`);
    const payload = data as {
      overall_score?: number;
      analyzed_posts?: number;
      distribution?: { emotion: string; count: number; share: number }[];
      timeline?: { bucket: string; avg_score: number; post_count: number }[];
    };
    return {
      topicId,
      range,
      overallScore: payload.overall_score ?? 0,
      analyzed: payload.analyzed_posts ?? 0,
      distribution: (payload.distribution ?? []).map((item) => ({
        emotion: item.emotion as SentimentPayload["distribution"][number]["emotion"],
        count: item.count,
        share: item.share,
      })),
      timeline: (payload.timeline ?? []).map((item) => ({
        bucket: item.bucket,
        avgScore: item.avg_score,
        posts: item.post_count,
      })),
      keywords: [],
      heatmap: { emotions: [], columns: [], values: [] },
    };
  },
};

export const demographicAPI = {
  async forTopic(topicId: string): Promise<AudiencePayload> {
    if (isMockMode()) return maybeMock(() => mock.getAudience(topicId));
    const { data } = await http.get(`/topics/${topicId}/demographics`);
    const payload = data as {
      sample_size?: number;
      language_distribution?: Record<string, number>;
      region_distribution?: Record<string, number>;
      interest_distribution?: Record<string, number>;
      audience_groups?: { label: string; share: number; description: string }[];
    };
    return {
      topicId,
      sampleSize: payload.sample_size ?? 0,
      languages: payload.language_distribution ?? {},
      regions: payload.region_distribution ?? {},
      interests: payload.interest_distribution ?? {},
      segments: payload.audience_groups ?? [],
    };
  },
};

export const trendAPI = {
  async forTopic(topicId: string, range: DateRange): Promise<TrendPayload> {
    if (isMockMode()) return maybeMock(() => mock.getTrends(topicId, range));
    const { data } = await http.get(`/topics/${topicId}/trends`);
    const payload = data as {
      overall?: { growth_rate?: number };
      rising_keywords?: { keyword: string; trend_score: number; growth_rate: number; velocity: number }[];
    };
    return {
      topicId,
      range,
      overallGrowth: payload.overall?.growth_rate ?? 0,
      rising: (payload.rising_keywords ?? []).map((item) => ({
        keyword: item.keyword,
        count: 0,
        growth: item.growth_rate,
        velocity: velocityFrom(item.growth_rate),
        score: item.trend_score,
        sparkline: [],
      })),
      series: [],
      clusters: [],
    };
  },
};

export const networkAPI = {
  async forTopic(topicId: string): Promise<NetworkPayload> {
    if (isMockMode()) return maybeMock(() => mock.getNetwork(topicId));
    const { data } = await http.get(`/topics/${topicId}/network`);
    const payload = data as {
      nodes?: { id: number; external_user_reference: string; display_name?: string; platform?: string; influence_score?: number; community_id?: number }[];
      edges?: { source_node: number; target_node: number; interaction_type?: string; weight?: number }[];
      top_influencers?: { reference: string; display_name?: string; influence_score?: number; community_id?: number }[];
    };
    const refById = new Map<number, string>();
    const nodes = (payload.nodes ?? []).map((node) => {
      refById.set(node.id, node.external_user_reference);
      return {
        id: node.external_user_reference,
        label: node.display_name || node.external_user_reference,
        handle: node.external_user_reference,
        community: node.community_id ?? 0,
        influence: node.influence_score ?? 0,
        platform: (node.platform as NetworkPayload["nodes"][number]["platform"]) ?? "twitter",
        posts: 0,
      };
    });
    const edges = (payload.edges ?? []).map((edge, index) => ({
      id: `e${index}`,
      source: refById.get(edge.source_node) ?? String(edge.source_node),
      target: refById.get(edge.target_node) ?? String(edge.target_node),
      weight: edge.weight ?? 1,
      type: (edge.interaction_type === "reply" ? "reply" : "mention") as "reply" | "mention",
    }));
    const communityIds = [...new Set(nodes.map((node) => node.community))];
    return {
      topicId,
      nodes,
      edges,
      influencers: (payload.top_influencers ?? []).map((item) => ({
        reference: item.reference,
        displayName: item.display_name || item.reference,
        influence: item.influence_score ?? 0,
        community: item.community_id ?? 0,
      })),
      communities: communityIds.map((id) => ({
        id,
        label: `Community ${id + 1}`,
        size: nodes.filter((node) => node.community === id).length,
      })),
      spreadPath: ["Propagation path is computed by the backend graph service."],
    };
  },
};

export const reportAPI = {
  async forTopic(topicId: string): Promise<Report> {
    if (isMockMode()) return maybeMock(() => mock.getReport(topicId));
    const { data } = await http.get(`/topics/${topicId}/dashboard`);
    const summary = (data as { summary?: Record<string, unknown> }).summary as
      | {
          title?: string;
          executive_summary?: string;
          key_findings?: string[];
          sections?: { heading: string; body: string }[];
          model_version?: string;
          generated_at?: string;
        }
      | undefined;
    if (!summary) {
      throw new Error("AI summary is not available from the backend yet.");
    }
    return {
      topicId,
      title: summary.title ?? "PulseGraph intelligence report",
      executiveSummary: summary.executive_summary ?? "",
      keyObservations: summary.key_findings ?? [],
      sections: summary.sections ?? [],
      recommendedActions: [],
      modelVersion: summary.model_version ?? "backend",
      generatedAt: summary.generated_at ?? new Date().toISOString(),
    };
  },
};

export const overviewAPI = {
  async forTopic(topicId: string, range: DateRange): Promise<OverviewPayload> {
    if (isMockMode()) return maybeMock(() => mock.getOverview(topicId, range));
    const { data } = await http.get(`/topics/${topicId}/dashboard`);
    const dashboard = data as {
      statistics?: { overview?: { total_posts?: number; total_engagement?: number; distinct_authors?: number; avg_sentiment?: number } };
      timeline?: { posts?: BackendPost[] };
    };
    const overview = dashboard.statistics?.overview ?? {};
    const avgSentiment = overview.avg_sentiment ?? 0;
    return {
      topicId,
      range,
      totalPosts: overview.total_posts ?? 0,
      postsDelta: 0,
      activeUsers: overview.distinct_authors ?? 0,
      usersDelta: 0,
      avgSentiment,
      sentimentLabel:
        avgSentiment >= 0.1
          ? "leaning positive"
          : avgSentiment <= -0.1
            ? "leaning negative"
            : "mixed",
      trendVelocity: "medium",
      velocityScore: 0,
      engagementTotal: overview.total_engagement ?? 0,
      latestPosts: (dashboard.timeline?.posts ?? []).map((post) => mapPost(post, topicId)),
    };
  },
};
