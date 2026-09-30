export type Platform =
  | "twitter"
  | "telegram"
  | "instagram"
  | "facebook"
  | "reddit"
  | "youtube";

export type Emotion =
  | "joy"
  | "surprise"
  | "neutral"
  | "fear"
  | "anger"
  | "sadness"
  | "disgust";

export type Scenario = "misinformation" | "event" | "launch";
export type ProjectStatus = "active" | "paused" | "archived";
export type MonitoringFrequency = "realtime" | "hourly" | "daily";
export type TrendVelocity = "high" | "medium" | "low";
export type DateRange = "24h" | "7d" | "14d";

export interface User {
  id: number;
  name: string;
  email: string;
  role: "admin" | "analyst" | "viewer";
}

export interface AuthResponse {
  token: string;
  expiresAt: string;
  user: User;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  platforms: Platform[];
  keywords: string[];
  frequency: MonitoringFrequency;
  status: ProjectStatus;
  topicIds: string[];
  lastUpdated: string;
  createdAt: string;
}

export interface CreateProjectInput {
  name: string;
  description: string;
  platforms: Platform[];
  keywords: string[];
  frequency: MonitoringFrequency;
}

export interface Topic {
  id: string;
  projectId: string;
  name: string;
  keywords: string[];
  description: string;
  scenario: Scenario;
}

export interface Post {
  id: string;
  topicId: string;
  platform: Platform;
  author: string;
  authorHandle: string;
  content: string;
  language: string;
  region: string;
  timestamp: string;
  engagement: number;
  emotion: Emotion;
  sentimentScore: number;
  confidence: number;
}

export interface GrowthBucket {
  bucket: string;
  posts: number;
}

export type TimelineEventKind = "first-post" | "growth" | "community" | "spike";

export interface TimelineEvent {
  id: string;
  timestamp: string;
  kind: TimelineEventKind;
  title: string;
  detail: string;
  platform?: Platform;
  engagement?: number;
}

export interface TimelinePayload {
  topicId: string;
  range: DateRange;
  growth: GrowthBucket[];
  events: TimelineEvent[];
  totals: { posts: number; engagements: number; authors: number };
}

export interface EmotionSlice {
  emotion: Emotion;
  count: number;
  share: number;
}

export interface SentimentPoint {
  bucket: string;
  avgScore: number;
  posts: number;
}

export interface KeywordMetric {
  keyword: string;
  count: number;
  growth: number;
  velocity: number;
  score: number;
}

export interface SentimentHeatmap {
  emotions: Emotion[];
  columns: string[];
  values: number[][];
}

export interface SentimentPayload {
  topicId: string;
  range: DateRange;
  overallScore: number;
  analyzed: number;
  distribution: EmotionSlice[];
  timeline: SentimentPoint[];
  keywords: KeywordMetric[];
  heatmap: SentimentHeatmap;
}

export interface AudienceGroup {
  label: string;
  share: number;
  description: string;
}

export interface AudiencePayload {
  topicId: string;
  sampleSize: number;
  languages: Record<string, number>;
  regions: Record<string, number>;
  interests: Record<string, number>;
  segments: AudienceGroup[];
}

export interface TrendPoint {
  bucket: string;
  value: number;
}

export interface TrendTopic {
  keyword: string;
  count: number;
  growth: number;
  velocity: TrendVelocity;
  score: number;
  sparkline: number[];
}

export interface TrendCluster {
  label: string;
  keywords: string[];
  share: number;
}

export interface TrendPayload {
  topicId: string;
  range: DateRange;
  overallGrowth: number;
  rising: TrendTopic[];
  series: { keyword: string; points: TrendPoint[] }[];
  clusters: TrendCluster[];
}

export interface NetworkNodeDatum {
  id: string;
  label: string;
  handle: string;
  community: number;
  influence: number;
  platform: Platform;
  posts: number;
}

export interface NetworkEdgeDatum {
  id: string;
  source: string;
  target: string;
  weight: number;
  type: "mention" | "reply";
}

export interface NetworkPayload {
  topicId: string;
  nodes: NetworkNodeDatum[];
  edges: NetworkEdgeDatum[];
  influencers: {
    reference: string;
    displayName: string;
    influence: number;
    community: number;
  }[];
  communities: { id: number; size: number; label: string }[];
  spreadPath: string[];
}

export interface ReportSection {
  heading: string;
  body: string;
}

export interface Report {
  topicId: string;
  title: string;
  executiveSummary: string;
  keyObservations: string[];
  sections: ReportSection[];
  recommendedActions: string[];
  modelVersion: string;
  generatedAt: string;
}

export interface OverviewPayload {
  topicId: string;
  range: DateRange;
  totalPosts: number;
  postsDelta: number;
  activeUsers: number;
  usersDelta: number;
  avgSentiment: number;
  sentimentLabel: string;
  trendVelocity: TrendVelocity;
  velocityScore: number;
  engagementTotal: number;
  latestPosts: Post[];
}

export interface ApiErrorShape {
  code: string;
  message: string;
}
