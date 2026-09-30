/**
 * Deterministic mock data engine.
 *
 * Generates a coherent PulseGraph universe — 3 projects, 5 monitored topics,
 * 1000+ posts, and every derived analytic (sentiment, trends, audience,
 * network, AI report) — so the whole product is demonstrable without a
 * backend. Everything is seeded, so numbers stay consistent across calls.
 */

import { ApiError } from "@/services/api";
import type {
  AudiencePayload,
  AuthResponse,
  CreateProjectInput,
  DateRange,
  Emotion,
  EmotionSlice,
  GrowthBucket,
  KeywordMetric,
  NetworkEdgeDatum,
  NetworkNodeDatum,
  NetworkPayload,
  OverviewPayload,
  Platform,
  Post,
  Project,
  Report,
  Scenario,
  SentimentPayload,
  SentimentPoint,
  TimelineEvent,
  TimelinePayload,
  Topic,
  TrendCluster,
  TrendPayload,
  TrendTopic,
  TrendVelocity,
  User,
} from "@/types";

// ── time constants + a fixed "now" anchor ───────────────────────────────────

const MIN = 60 * 1000;
const HOUR = 60 * MIN;
const DAY = 24 * HOUR;
const NOW = Date.now();

const iso = (ms: number) => new Date(ms).toISOString();

// ── deterministic PRNG helpers ──────────────────────────────────────────────

type Rng = () => number;

function mulberry32(seed: number): Rng {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function hashString(input: string): number {
  let hash = 2166136261;
  for (let i = 0; i < input.length; i++) {
    hash ^= input.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return hash >>> 0;
}

function pick<T>(rng: Rng, items: readonly T[]): T {
  return items[Math.floor(rng() * items.length)];
}

function between(rng: Rng, min: number, max: number): number {
  return min + rng() * (max - min);
}

function weighted(rng: Rng, pairs: readonly (readonly [string, number])[]): string {
  const total = pairs.reduce((sum, [, weight]) => sum + weight, 0);
  let roll = rng() * total;
  for (const [value, weight] of pairs) {
    roll -= weight;
    if (roll <= 0) return value;
  }
  return pairs[pairs.length - 1][0];
}

function lerp(start: number, end: number, t: number): number {
  return start + (end - start) * t;
}

function r2(value: number, digits = 2): number {
  const factor = 10 ** digits;
  return Math.round(value * factor) / factor;
}

const clampScore = (value: number) => Math.max(-1, Math.min(1, value));

// ── static catalogue ────────────────────────────────────────────────────────

interface AuthorDef {
  handle: string;
  name: string;
  platform: Platform;
  tier: "broadcaster" | "amplifier" | "participant";
}

const AUTHORS: AuthorDef[] = [
  { handle: "floodwatch_assam", name: "FloodWatch Assam", platform: "twitter", tier: "broadcaster" },
  { handle: "newsbharat", name: "News Bharat", platform: "twitter", tier: "broadcaster" },
  { handle: "civic_watchdog", name: "Civic Watchdog", platform: "twitter", tier: "amplifier" },
  { handle: "monsoon_alerts", name: "Monsoon Alerts", platform: "twitter", tier: "broadcaster" },
  { handle: "citizen_ravi", name: "Ravi K", platform: "twitter", tier: "participant" },
  { handle: "metro_maya", name: "Maya R", platform: "twitter", tier: "amplifier" },
  { handle: "trafficpulse_city", name: "TrafficPulse City", platform: "twitter", tier: "amplifier" },
  { handle: "runner_ananya", name: "Ananya S", platform: "twitter", tier: "participant" },
  { handle: "marathon_mumbai", name: "Mumbai Marathon Fan", platform: "twitter", tier: "amplifier" },
  { handle: "pace_setter", name: "Pace Setter", platform: "twitter", tier: "participant" },
  { handle: "gadget_guru", name: "Gadget Guru", platform: "twitter", tier: "broadcaster" },
  { handle: "honest_reviewer", name: "Honest Reviewer", platform: "twitter", tier: "amplifier" },
  { handle: "delhi_dad", name: "Delhi Dad", platform: "twitter", tier: "participant" },
  { handle: "kolkata_kriti", name: "Kriti B", platform: "twitter", tier: "participant" },
  { handle: "bengaluru_buzz", name: "Bengaluru Buzz", platform: "twitter", tier: "amplifier" },
  { handle: "hyderabad_techie", name: "Hyd Techie", platform: "twitter", tier: "participant" },
  { handle: "pune_runner", name: "Pune Runner", platform: "twitter", tier: "participant" },
  { handle: "smartphone_sachin", name: "Sachin Reviews", platform: "twitter", tier: "amplifier" },
  { handle: "run_with_riya", name: "Riya Runs", platform: "twitter", tier: "participant" },
  { handle: "gadget_throws", name: "Gadget Throws", platform: "twitter", tier: "participant" },
  { handle: "assam_relief_net", name: "Assam Relief Net", platform: "telegram", tier: "broadcaster" },
  { handle: "monsoon_watch", name: "Monsoon Watch", platform: "telegram", tier: "amplifier" },
  { handle: "city_events_hub", name: "City Events Hub", platform: "telegram", tier: "amplifier" },
  { handle: "metro_updates", name: "Metro Updates", platform: "telegram", tier: "amplifier" },
  { handle: "product_deals_in", name: "Deals India", platform: "telegram", tier: "amplifier" },
  { handle: "daily_digest_in", name: "Daily Digest", platform: "telegram", tier: "broadcaster" },
  { handle: "relief_volunteers", name: "Relief Volunteers", platform: "telegram", tier: "amplifier" },
  { handle: "race_day_crew", name: "Race Day Crew", platform: "telegram", tier: "amplifier" },
  { handle: "nimbus_updates", name: "Nimbus Updates", platform: "telegram", tier: "broadcaster" },
  { handle: "city_general", name: "City General", platform: "telegram", tier: "participant" },
  { handle: "mausam_mitra", name: "Mausam Mitra", platform: "telegram", tier: "amplifier" },
];

type Bucket = "very_neg" | "neg" | "neu" | "pos" | "very_pos";

interface ScenarioDef {
  templates: Record<Bucket, string[]>;
  emotions: Record<Bucket, Emotion[]>;
  communityLabels: string[];
  clusters: { label: string; words: string[] }[];
}

const SCENARIOS: Record<Scenario, ScenarioDef> = {
  misinformation: {
    templates: {
      very_neg: [
        "{kw} update: this is much worse than officials admit. Do not trust the official numbers.",
        "Families are still stranded because of {kw} and rescue has not reached. Where is the administration?",
        "Forwarding an urgent alert about {kw} — share before it gets deleted.",
      ],
      neg: [
        "Getting mixed reports about {kw}. Some say relief is moving, others say nothing has changed.",
        "The {kw} situation is tense. We need verified information, not forwarded rumours.",
        "Relief material for {kw} is delayed again. Shelters are still waiting.",
      ],
      neu: [
        "Traffic diverted near the bridge because of {kw}. Plan your commute accordingly.",
        "Helpline numbers for {kw} relief are pinned in the group. Please verify before forwarding.",
        "Weather office: {kw} conditions should ease over the next two days.",
      ],
      pos: [
        "Rescue teams are finally reaching the worst-hit areas. Gratitude to everyone helping with {kw} relief.",
        "Community kitchens for {kw} relief are running. Volunteers are doing extraordinary work.",
        "Supplies are moving again on the highway — the {kw} effort is picking up pace.",
      ],
      very_pos: [
        "Incredible response to the {kw} relief drive today. Proud of this community.",
        "The coordination between volunteers and officials on {kw} has been outstanding this week.",
        "Update: water levels receding. The {kw} situation is finally stabilising.",
      ],
    },
    emotions: {
      very_neg: ["fear", "fear", "fear", "anger", "anger", "sadness"],
      neg: ["fear", "fear", "anger", "sadness", "neutral"],
      neu: ["neutral", "neutral", "neutral", "surprise"],
      pos: ["joy", "joy", "surprise"],
      very_pos: ["joy", "joy", "joy", "surprise"],
    },
    communityLabels: [
      "Relief response network",
      "Local news cluster",
      "Rumour propagation",
      "Official advisories",
    ],
    clusters: [
      { label: "Relief operations", words: ["relief", "rescue", "shelter", "supplies", "volunteers", "kit"] },
      { label: "Rumour mill", words: ["rumour", "alert", "forward", "deleted", "viral", "verify"] },
      { label: "Official response", words: ["officials", "administration", "helpline", "weather", "highway", "government"] },
    ],
  },
  event: {
    templates: {
      very_neg: [
        "The {kw} route change is a disaster for local businesses. Zero notice, zero consultation.",
        "Three days of {kw} closures and residents were never asked. Poor planning.",
        "The {kw} registration process crashed again. Absolute chaos for participants.",
      ],
      neg: [
        "Not convinced the {kw} closures are worth it. Traffic pain for everyone.",
        "Organisers of the {kw} should fix the water points before race day. Basic stuff.",
        "Still waiting for clarity on the {kw} access passes.",
      ],
      neu: [
        "Route map for the {kw} is out — check which roads close this weekend.",
        "Registration for the {kw} closes this week. Details in the thread.",
        "{kw} briefing at 4 PM: expected timelines and diversions.",
      ],
      pos: [
        "Registered for the {kw}! Training going well, see you at the start line.",
        "The {kw} expo was great today. Picked up the bib and new shoes.",
        "Volunteering at the {kw} this year. Energy in the city is building.",
      ],
      very_pos: [
        "The {kw} energy today is unreal. Whole neighbourhood is out cheering.",
        "Personal best at the {kw} today! The crowd carried everyone home.",
      ],
    },
    emotions: {
      very_neg: ["anger", "anger", "disgust", "sadness"],
      neg: ["anger", "anger", "sadness"],
      neu: ["neutral", "neutral", "neutral"],
      pos: ["joy", "joy", "surprise"],
      very_pos: ["joy", "joy", "joy", "surprise", "surprise"],
    },
    communityLabels: [
      "Organisers & logistics",
      "Participant community",
      "Local residents",
      "City media",
    ],
    clusters: [
      { label: "Logistics & route", words: ["route", "traffic", "diversion", "closures", "roads", "briefing"] },
      { label: "Participation buzz", words: ["registration", "training", "expo", "bib", "volunteering", "start"] },
      { label: "City impact", words: ["businesses", "residents", "energy", "crowd", "water"] },
    ],
  },
  launch: {
    templates: {
      very_neg: [
        "The {kw} pricing is a joke. Same specs as last year with a bigger number.",
        "Battery claims for the {kw} look inflated again. Wait for real reviews.",
        "Disappointed with the {kw} — no {kw2} upgrade and no charger in the box.",
      ],
      neg: [
        "The {kw} looks fine but not worth the upgrade from last year's model.",
        "The {kw2} improvement on the {kw} feels more marketing than engineering.",
        "Preorder bonuses for the {kw} are underwhelming this year.",
      ],
      neu: [
        "The {kw} goes on sale tomorrow. Spec sheet and price comparison in the thread.",
        "Nimbus announced the {kw} today — full specs and {kw2} details here.",
        "First impressions thread for the {kw} will be updated once reviews drop.",
      ],
      pos: [
        "Got hands-on with the {kw}. Build quality is genuinely good and the {kw2} upgrade is noticeable.",
        "Preordered the {kw}. The {kw2} improvements actually matter this time.",
        "Camera samples from the {kw} look great for this price range.",
      ],
      very_pos: [
        "The {kw} sold out in 40 minutes. That {kw2} upgrade is worth every rupee.",
        "Two days with the {kw} and it is excellent. Battery and {kw2} are superb.",
      ],
    },
    emotions: {
      very_neg: ["disgust", "anger", "sadness", "anger"],
      neg: ["disgust", "sadness", "anger"],
      neu: ["neutral", "neutral", "surprise"],
      pos: ["joy", "surprise", "joy"],
      very_pos: ["joy", "joy", "surprise"],
    },
    communityLabels: [
      "Enthusiast community",
      "Reviewer cluster",
      "Retail & deals",
      "Critics",
    ],
    clusters: [
      { label: "Product & specs", words: ["battery", "camera", "specs", "upgrade", "build", "charger"] },
      { label: "Commerce", words: ["preorder", "sale", "stock", "sold", "price", "rupee"] },
      { label: "Reviews", words: ["reviews", "hands", "impressions", "thread", "disappointed", "excellent"] },
    ],
  },
};

interface InterestRule {
  label: string;
  words: string[];
}

const INTEREST_RULES: InterestRule[] = [
  { label: "Disaster response", words: ["flood", "rescue", "relief", "evacuation", "shelter", "monsoon"] },
  { label: "Public safety", words: ["alert", "warning", "safety", "emergency", "advisory"] },
  { label: "Transport & mobility", words: ["traffic", "route", "metro", "diversion", "roads", "commute"] },
  { label: "Technology", words: ["battery", "camera", "chip", "specs", "upgrade", "device"] },
  { label: "Events & culture", words: ["marathon", "race", "expo", "event", "volunteering"] },
  { label: "Business & retail", words: ["price", "sale", "preorder", "businesses", "stock"] },
];

// ── projects + topics ───────────────────────────────────────────────────────

interface TopicDef {
  topic: Topic;
  languages: [string, number][];
  regions: [string, number][];
  postCount: number;
}

const PROJECT_DEFS: Project[] = [
  {
    id: "flood-watch",
    name: "Flood Rumour Analysis",
    description:
      "Monitors rumour spread and relief coordination during the Assam flood season.",
    platforms: ["twitter", "telegram"],
    keywords: ["flood", "relief", "warning", "evacuation"],
    frequency: "realtime",
    status: "active",
    topicIds: ["flood-rumour", "monsoon-drive"],
    lastUpdated: iso(NOW - 4 * MIN),
    createdAt: iso(NOW - 26 * DAY),
  },
  {
    id: "city-pulse",
    name: "City Pulse - Public Events",
    description:
      "Tracks city-scale events: marathon logistics, metro openings and commuter sentiment.",
    platforms: ["twitter", "telegram"],
    keywords: ["marathon", "metro", "traffic", "route"],
    frequency: "hourly",
    status: "active",
    topicIds: ["city-marathon", "metro-line-5"],
    lastUpdated: iso(NOW - 11 * MIN),
    createdAt: iso(NOW - 18 * DAY),
  },
  {
    id: "product-radar",
    name: "Product Radar - Nimbus X1",
    description:
      "Launch-day listening for the Nimbus X1: hype, reviews, criticism and pricing talk.",
    platforms: ["twitter", "telegram"],
    keywords: ["nimbus", "battery", "price", "camera"],
    frequency: "hourly",
    status: "paused",
    topicIds: ["nimbus-x1"],
    lastUpdated: iso(NOW - 3 * HOUR),
    createdAt: iso(NOW - 9 * DAY),
  },
];

const BASE_TOPIC_DEFS: TopicDef[] = [
  {
    topic: {
      id: "flood-rumour",
      projectId: "flood-watch",
      name: "Assam Flood - Rumour Watch",
      keywords: ["flood", "relief", "evacuation", "dam", "rumour"],
      description: "Tracks rumour spread versus verified relief information.",
      scenario: "misinformation",
    },
    languages: [["en", 0.45], ["hi", 0.35], ["bn", 0.2]],
    regions: [["Assam", 5], ["West Bengal", 2], ["Delhi", 1.5], ["Other", 1]],
    postCount: 240,
  },
  {
    topic: {
      id: "monsoon-drive",
      projectId: "flood-watch",
      name: "Monsoon Preparedness Drive",
      keywords: ["monsoon", "shelter", "advisory", "drill", "alert"],
      description: "Public communication around monsoon readiness and drills.",
      scenario: "misinformation",
    },
    languages: [["en", 0.5], ["hi", 0.35], ["bn", 0.15]],
    regions: [["Assam", 4], ["Maharashtra", 2], ["Delhi", 2], ["Other", 1]],
    postCount: 180,
  },
  {
    topic: {
      id: "city-marathon",
      projectId: "city-pulse",
      name: "City Marathon 2026",
      keywords: ["marathon", "route", "registration", "traffic", "race"],
      description: "Route changes, registrations and race-day energy.",
      scenario: "event",
    },
    languages: [["en", 0.55], ["hi", 0.3], ["ta", 0.15]],
    regions: [["Maharashtra", 4.5], ["Karnataka", 2.5], ["Delhi", 2], ["Other", 1]],
    postCount: 220,
  },
  {
    topic: {
      id: "metro-line-5",
      projectId: "city-pulse",
      name: "Metro Line 5 Opening",
      keywords: ["metro", "line5", "opening", "commute", "fares"],
      description: "Commuter reaction and logistics around the new metro line.",
      scenario: "event",
    },
    languages: [["en", 0.5], ["hi", 0.32], ["ta", 0.18]],
    regions: [["Delhi", 4], ["Karnataka", 2], ["Telangana", 2], ["Other", 1]],
    postCount: 180,
  },
  {
    topic: {
      id: "nimbus-x1",
      projectId: "product-radar",
      name: "Nimbus X1 Launch",
      keywords: ["nimbus", "battery", "camera", "price", "preorder"],
      description: "Launch spike, reviews and pricing debate for the Nimbus X1.",
      scenario: "launch",
    },
    languages: [["en", 0.6], ["hi", 0.25], ["te", 0.15]],
    regions: [["Karnataka", 3], ["Telangana", 2.5], ["Delhi", 2.5], ["Other", 2]],
    postCount: 220,
  },
];

const projects: Project[] = PROJECT_DEFS.map((project) => ({ ...project }));
const topicDefs: TopicDef[] = BASE_TOPIC_DEFS.map((def) => ({
  ...def,
  topic: { ...def.topic },
}));

function topicDefById(topicId: string): TopicDef | undefined {
  return topicDefs.find((def) => def.topic.id === topicId);
}

function projectForTopic(topicId: string): Project | undefined {
  const def = topicDefById(topicId);
  if (!def) return undefined;
  return projects.find((project) => project.id === def.topic.projectId);
}

// ── post generation ─────────────────────────────────────────────────────────

function bucketFor(score: number): Bucket {
  if (score <= -0.55) return "very_neg";
  if (score <= -0.18) return "neg";
  if (score < 0.18) return "neu";
  if (score < 0.55) return "pos";
  return "very_pos";
}

function arcScore(scenario: Scenario, progress: number): number {
  if (scenario === "misinformation") {
    return progress < 0.65
      ? lerp(-0.05, -0.68, progress / 0.65)
      : lerp(-0.68, -0.05, (progress - 0.65) / 0.35);
  }
  if (scenario === "event") {
    return progress < 0.5
      ? lerp(-0.12, 0.62, progress / 0.5)
      : lerp(0.62, 0.3, (progress - 0.5) / 0.5);
  }
  if (progress < 0.35) return lerp(0.05, 0.7, progress / 0.35);
  if (progress < 0.7) return lerp(0.7, -0.3, (progress - 0.35) / 0.35);
  return lerp(-0.3, 0.15, (progress - 0.7) / 0.3);
}

function fillTemplate(template: string, keywords: string[], rng: Rng): string {
  const kw = pick(rng, keywords);
  const kw2 = pick(rng, keywords);
  return template.split("{kw2}").join(kw2).split("{kw}").join(kw);
}

function tierEngagement(rng: Rng, tier: AuthorDef["tier"]): number {
  if (tier === "broadcaster") return Math.round(between(rng, 400, 4200));
  if (tier === "amplifier") return Math.round(between(rng, 60, 900));
  return Math.round(between(rng, 1, 140));
}

interface TopicData {
  def: TopicDef;
  posts: Post[];
  sentimentCache: Map<DateRange, SentimentPayload>;
  trendsCache: Map<DateRange, TrendPayload>;
  timelineCache: Map<DateRange, TimelinePayload>;
  overviewCache: Map<DateRange, OverviewPayload>;
  audienceCache?: AudiencePayload;
  networkCache?: NetworkPayload;
  reportCache?: Report;
}

const topicData = new Map<string, TopicData>();

function buildTopicData(def: TopicDef): TopicData {
  const rng = mulberry32(hashString(def.topic.id) ^ 0x9e3779b9);
  const project = projects.find((item) => item.id === def.topic.projectId);
  const platforms = project?.platforms ?? ["twitter"];
  const authors = AUTHORS.filter((author) => platforms.includes(author.platform));
  const scenario = SCENARIOS[def.topic.scenario];

  const start = NOW - 14 * DAY;
  const posts: Post[] = [];

  for (let index = 0; index < def.postCount; index++) {
    const progress = index / Math.max(1, def.postCount - 1);
    const offset = progress * 14 * DAY + between(rng, -0.3, 0.3) * DAY;
    const timestamp = Math.min(NOW - MIN, Math.max(start, start + offset));

    const author = pick(rng, authors);
    const score = clampScore(arcScore(def.topic.scenario, progress) + between(rng, -0.25, 0.25));
    const bucket = bucketFor(score);
    const content = fillTemplate(pick(rng, scenario.templates[bucket]), def.topic.keywords, rng);
    const engagementBase = tierEngagement(rng, author.tier);
    const engagement = index % 47 === 23 ? engagementBase * 8 : engagementBase;

    posts.push({
      id: `${def.topic.id}-p${index + 1}`,
      topicId: def.topic.id,
      platform: author.platform,
      author: author.name,
      authorHandle: author.handle,
      content,
      language: weighted(rng, def.languages),
      region: weighted(rng, def.regions),
      timestamp: iso(timestamp),
      engagement,
      emotion: pick(rng, scenario.emotions[bucket]),
      sentimentScore: r2(score, 3),
      confidence: r2(between(rng, 0.55, 0.92), 3),
    });
  }

  posts.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

  return {
    def,
    posts,
    sentimentCache: new Map(),
    trendsCache: new Map(),
    timelineCache: new Map(),
    overviewCache: new Map(),
  };
}

function ensureTopicData(topicId: string): TopicData {
  const existing = topicData.get(topicId);
  if (existing) return existing;
  const def = topicDefById(topicId);
  if (!def) throw new ApiError("topic_not_found", `Topic ${topicId} was not found.`);
  const data = buildTopicData(def);
  topicData.set(topicId, data);
  return data;
}

// ── shared aggregation helpers ──────────────────────────────────────────────

function rangeMs(range: DateRange): number {
  if (range === "24h") return DAY;
  if (range === "7d") return 7 * DAY;
  return 14 * DAY;
}

function rangeBucketMs(range: DateRange): number {
  if (range === "24h") return HOUR;
  if (range === "7d") return 4 * HOUR;
  return 12 * HOUR;
}

function postsInRange(data: TopicData, range: DateRange): Post[] {
  const cutoff = NOW - rangeMs(range);
  return data.posts.filter((post) => new Date(post.timestamp).getTime() >= cutoff);
}

const STOPWORDS = new Set([
  "the", "and", "for", "with", "that", "this", "from", "are", "was", "has",
  "have", "you", "your", "our", "their", "not", "but", "who", "what", "when",
  "where", "will", "would", "could", "should", "about", "they", "them",
  "then", "than", "into", "over", "under", "here", "there", "been", "being",
  "just", "also", "after", "before", "between", "during", "can", "may",
  "still", "now", "new", "hai", "hain", "aur", "ke", "ki", "ka", "se", "ko",
  "kya", "nahi", "mein", "par", "yeh", "wo", "kar", "koi", "bhi", "tha",
  "raha", "rahe", "hoga", "liye", "bahut", "kuch", "sab", "tak",
]);

function tokenize(text: string): string[] {
  const matches = text.toLowerCase().match(/[a-z\u0900-\u097F]{3,}/g);
  return (matches ?? []).filter((word) => !STOPWORDS.has(word));
}

interface KeywordAccumulator {
  count: number;
  first: number;
  second: number;
  recent: number;
}

function accumulateKeywords(posts: Post[], range: DateRange): Map<string, KeywordAccumulator> {
  const halfPoint = NOW - rangeMs(range) / 2;
  const recentPoint = NOW - rangeMs(range) * 0.25;
  const map = new Map<string, KeywordAccumulator>();

  for (const post of posts) {
    const time = new Date(post.timestamp).getTime();
    for (const word of tokenize(post.content)) {
      const acc = map.get(word) ?? { count: 0, first: 0, second: 0, recent: 0 };
      acc.count += 1;
      if (time < halfPoint) acc.first += 1;
      else acc.second += 1;
      if (time >= recentPoint) acc.recent += 1;
      map.set(word, acc);
    }
  }
  return map;
}

function keywordMetrics(posts: Post[], range: DateRange, limit: number): KeywordMetric[] {
  const accumulators = accumulateKeywords(posts, range);
  const totalRecent = Math.max(1, posts.filter((post) => new Date(post.timestamp).getTime() >= NOW - rangeMs(range) * 0.25).length);
  const metrics: KeywordMetric[] = [];

  for (const [keyword, acc] of accumulators) {
    if (acc.count < 3) continue;
    const growth = (acc.second - acc.first) / Math.max(1, acc.first);
    const volumeNorm = Math.min(1, Math.log1p(acc.count) / Math.log1p(120));
    const growthNorm = Math.min(1, Math.max(0, growth) / 2);
    const recencyNorm = Math.min(1, acc.recent / totalRecent * 3);
    const score = r2(100 * (0.5 * volumeNorm + 0.3 * growthNorm + 0.2 * recencyNorm));
    metrics.push({ keyword, count: acc.count, growth: r2(growth, 3), velocity: r2(growth, 3), score });
  }

  metrics.sort((a, b) => b.score - a.score || a.keyword.localeCompare(b.keyword));
  return metrics.slice(0, limit);
}

function keywordSeries(posts: Post[], keyword: string, range: DateRange, points = 12): number[] {
  const size = rangeMs(range) / points;
  const start = NOW - rangeMs(range);
  const buckets = new Array<number>(points).fill(0);
  for (const post of posts) {
    if (!post.content.toLowerCase().includes(keyword)) continue;
    const index = Math.floor((new Date(post.timestamp).getTime() - start) / size);
    if (index >= 0 && index < points) buckets[index] += 1;
  }
  return buckets;
}

function sentimentTimeline(posts: Post[], range: DateRange): SentimentPoint[] {
  const size = rangeBucketMs(range);
  const start = NOW - rangeMs(range);
  const buckets = new Map<number, { sum: number; count: number }>();

  for (const post of posts) {
    const index = Math.floor((new Date(post.timestamp).getTime() - start) / size);
    const bucket = buckets.get(index) ?? { sum: 0, count: 0 };
    bucket.sum += post.sentimentScore;
    bucket.count += 1;
    buckets.set(index, bucket);
  }

  const total = Math.ceil(rangeMs(range) / size);
  const points: SentimentPoint[] = [];
  for (let index = 0; index <= total; index++) {
    const bucket = buckets.get(index);
    if (!bucket) continue;
    points.push({
      bucket: iso(start + index * size),
      avgScore: r2(bucket.sum / bucket.count, 3),
      posts: bucket.count,
    });
  }
  return points;
}

function growthSeries(posts: Post[], range: DateRange): GrowthBucket[] {
  const size = rangeBucketMs(range);
  const start = NOW - rangeMs(range);
  const buckets = new Map<number, number>();
  for (const post of posts) {
    const index = Math.floor((new Date(post.timestamp).getTime() - start) / size);
    buckets.set(index, (buckets.get(index) ?? 0) + 1);
  }
  const total = Math.ceil(rangeMs(range) / size);
  const out: GrowthBucket[] = [];
  for (let index = 0; index <= total; index++) {
    const count = buckets.get(index);
    if (!count) continue;
    out.push({ bucket: iso(start + index * size), posts: count });
  }
  return out;
}

function describeSentiment(score: number): string {
  if (score >= 0.35) return "strongly positive";
  if (score >= 0.1) return "leaning positive";
  if (score > -0.1) return "mixed";
  if (score > -0.35) return "leaning negative";
  return "strongly negative";
}

// ── analytic builders ───────────────────────────────────────────────────────

function buildSentiment(data: TopicData, range: DateRange): SentimentPayload {
  const cached = data.sentimentCache.get(range);
  if (cached) return cached;

  const posts = postsInRange(data, range);
  const counts = new Map<Emotion, number>();
  let scoreSum = 0;
  for (const post of posts) {
    counts.set(post.emotion, (counts.get(post.emotion) ?? 0) + 1);
    scoreSum += post.sentimentScore;
  }
  const analyzed = posts.length;
  const distribution: EmotionSlice[] = [...counts.entries()]
    .map(([emotion, count]) => ({ emotion, count, share: r2(count / Math.max(1, analyzed), 3) }))
    .sort((a, b) => b.count - a.count);

  const heatmapEmotions = distribution.slice(0, 5).map((item) => item.emotion);
  const columnCount = 10;
  const columnSize = rangeMs(range) / columnCount;
  const rangeStart = NOW - rangeMs(range);
  const heatmapColumns = Array.from({ length: columnCount }, (_, index) =>
    iso(rangeStart + ((index + 1) * rangeMs(range)) / columnCount),
  );
  const heatmapValues = heatmapEmotions.map(() => new Array<number>(columnCount).fill(0));
  for (const post of posts) {
    const column = Math.min(
      columnCount - 1,
      Math.max(0, Math.floor((new Date(post.timestamp).getTime() - rangeStart) / columnSize)),
    );
    const row = heatmapEmotions.indexOf(post.emotion);
    if (row >= 0) heatmapValues[row][column] += 1;
  }

  const payload: SentimentPayload = {
    topicId: data.def.topic.id,
    range,
    overallScore: r2(scoreSum / Math.max(1, analyzed), 3),
    analyzed,
    distribution,
    timeline: sentimentTimeline(posts, range),
    keywords: keywordMetrics(posts, range, 12),
    heatmap: { emotions: heatmapEmotions, columns: heatmapColumns, values: heatmapValues },
  };
  data.sentimentCache.set(range, payload);
  return payload;
}

function velocityLabel(growth: number): TrendVelocity {
  if (growth >= 0.9) return "high";
  if (growth >= 0.25) return "medium";
  return "low";
}

function buildTrends(data: TopicData, range: DateRange): TrendPayload {
  const cached = data.trendsCache.get(range);
  if (cached) return cached;

  const posts = postsInRange(data, range);
  const halfPoint = NOW - rangeMs(range) / 2;
  const first = posts.filter((post) => new Date(post.timestamp).getTime() < halfPoint).length;
  const second = posts.length - first;
  const overallGrowth = r2((second - first) / Math.max(1, first), 3);

  const metrics = keywordMetrics(posts, range, 10);
  const rising: TrendTopic[] = metrics.slice(0, 8).map((metric) => ({
    keyword: metric.keyword,
    count: metric.count,
    growth: metric.growth,
    velocity: velocityLabel(metric.growth),
    score: metric.score,
    sparkline: keywordSeries(posts, metric.keyword, range),
  }));

  const series = rising.slice(0, 5).map((topic) => ({
    keyword: topic.keyword,
    points: keywordSeries(posts, topic.keyword, range).map((value, index) => ({
      bucket: iso(NOW - rangeMs(range) + ((index + 1) * rangeMs(range)) / 12),
      value,
    })),
  }));

  const scenario = SCENARIOS[data.def.topic.scenario];
  const clusterTotals = scenario.clusters.map((cluster) => ({ ...cluster, total: 0 }));
  let otherTotal = 0;
  let grandTotal = 0;

  for (const metric of metrics) {
    grandTotal += metric.count;
    const cluster = clusterTotals.find((item) => item.words.includes(metric.keyword));
    if (cluster) cluster.total += metric.count;
    else otherTotal += metric.count;
  }

  const clusters: TrendCluster[] = clusterTotals
    .filter((cluster) => cluster.total > 0)
    .map((cluster) => ({
      label: cluster.label,
      keywords: metrics
        .filter((metric) => cluster.words.includes(metric.keyword))
        .slice(0, 4)
        .map((metric) => metric.keyword),
      share: r2(cluster.total / Math.max(1, grandTotal), 3),
    }));
  if (otherTotal > 0) {
    clusters.push({
      label: "Other signals",
      keywords: metrics.filter((metric) => !clusterTotals.some((c) => c.words.includes(metric.keyword))).slice(0, 4).map((metric) => metric.keyword),
      share: r2(otherTotal / Math.max(1, grandTotal), 3),
    });
  }

  const payload: TrendPayload = {
    topicId: data.def.topic.id,
    range,
    overallGrowth,
    rising,
    series,
    clusters: clusters.filter((cluster) => cluster.keywords.length > 0),
  };
  data.trendsCache.set(range, payload);
  return payload;
}

function buildTimeline(data: TopicData, range: DateRange): TimelinePayload {
  const cached = data.timelineCache.get(range);
  if (cached) return cached;

  const posts = postsInRange(data, range);
  const ascending = [...posts].sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime(),
  );
  const growth = growthSeries(posts, range);

  const events: TimelineEvent[] = [];
  const firstPost = ascending[0];
  if (firstPost) {
    events.push({
      id: `${data.def.topic.id}-first`,
      timestamp: firstPost.timestamp,
      kind: "first-post",
      title: "First post detected",
      detail: `Monitoring began with a post by ${firstPost.author} on ${firstPost.platform === "twitter" ? "X" : "Telegram"}.`,
      platform: firstPost.platform,
    });
  }

  let peakIndex = -1;
  let peak = 0;
  growth.forEach((bucket, index) => {
    if (bucket.posts > peak) {
      peak = bucket.posts;
      peakIndex = index;
    }
  });
  if (peakIndex >= 0 && growth.length > 1) {
    const previous = growth[Math.max(0, peakIndex - 1)];
    const delta = previous && previous.posts > 0 ? (peak - previous.posts) / previous.posts : 0;
    events.push({
      id: `${data.def.topic.id}-growth`,
      timestamp: growth[peakIndex].bucket,
      kind: "growth",
      title: "Discussion growth spiked",
      detail: `Volume hit ${peak} posts in one window${delta > 0 ? ` (+${Math.round(delta * 100)}% vs previous window)` : ""}.`,
    });
  }

  const spikes = [...posts]
    .sort((a, b) => b.engagement - a.engagement)
    .slice(0, 3);
  for (const spike of spikes) {
    events.push({
      id: `${data.def.topic.id}-spike-${spike.id}`,
      timestamp: spike.timestamp,
      kind: "spike",
      title: "Viral activity spike",
      detail: `${spike.author} drove ${spike.engagement.toLocaleString()} engagements with a ${spike.emotion} post.`,
      platform: spike.platform,
      engagement: spike.engagement,
    });
  }

  const network = buildNetwork(data);
  events.push({
    id: `${data.def.topic.id}-community`,
    timestamp: iso(NOW - 6 * HOUR),
    kind: "community",
    title: "New communities joined",
    detail: `${network.communities.length} communities are active; ${network.communities[0]?.label ?? "core cluster"} is expanding fastest.`,
  });

  events.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

  const payload: TimelinePayload = {
    topicId: data.def.topic.id,
    range,
    growth,
    events: events.slice(0, 12),
    totals: {
      posts: posts.length,
      engagements: posts.reduce((sum, post) => sum + post.engagement, 0),
      authors: new Set(posts.map((post) => post.authorHandle)).size,
    },
  };
  data.timelineCache.set(range, payload);
  return payload;
}

function buildNetwork(data: TopicData): NetworkPayload {
  if (data.networkCache) return data.networkCache;

  const rng = mulberry32(hashString(`${data.def.topic.id}:network`) ^ 0x51ed270b);
  const project = projectForTopic(data.def.topic.id);
  const platforms = project?.platforms ?? ["twitter"];
  const authors = AUTHORS.filter((author) => platforms.includes(author.platform));
  const scenario = SCENARIOS[data.def.topic.scenario];
  const communityCount = 4;

  const nodes: NetworkNodeDatum[] = authors.map((author, index) => ({
    id: author.handle,
    label: author.name,
    handle: author.handle,
    community: index % communityCount,
    influence: 0,
    platform: author.platform,
    posts: 0,
  }));
  const nodeByHandle = new Map(nodes.map((node) => [node.id, node]));

  const forTopic = data.posts;
  for (const post of forTopic) {
    const node = nodeByHandle.get(post.authorHandle);
    if (node) node.posts += 1;
  }

  const edges: NetworkEdgeDatum[] = [];
  const used = new Set<string>();
  const edgeTarget = 72 + Math.floor(rng() * 40);
  for (let index = 0; index < edgeTarget; index++) {
    const source = pick(rng, nodes);
    const target = pick(rng, nodes);
    if (source.id === target.id) continue;
    const key = `${source.id}->${target.id}`;
    if (used.has(key)) continue;
    used.add(key);
    edges.push({
      id: `e${edges.length + 1}`,
      source: source.id,
      target: target.id,
      weight: 1 + Math.floor(rng() * 9),
      type: rng() < 0.45 ? "reply" : "mention",
    });
  }

  const incoming = new Map<string, number>();
  for (const edge of edges) {
    incoming.set(edge.target, (incoming.get(edge.target) ?? 0) + edge.weight);
  }
  const maxIncoming = Math.max(1, ...incoming.values());
  for (const node of nodes) {
    const base = (incoming.get(node.id) ?? 0) / maxIncoming;
    const activity = Math.min(1, node.posts / 8);
    node.influence = r2(0.75 * base + 0.25 * activity, 3);
  }

  const influencers = [...nodes]
    .sort((a, b) => b.influence - a.influence)
    .slice(0, 8)
    .map((node) => ({
      reference: node.handle,
      displayName: node.label,
      influence: node.influence,
      community: node.community,
    }));

  const communities = scenario.communityLabels.map((label, id) => ({
    id,
    label,
    size: nodes.filter((node) => node.community === id).length,
  }));

  const top = influencers[0];
  const amplifiers = nodes.filter((node) => node.community === top?.community && node.id !== top?.reference).length;
  const spreadPath = [
    top ? `Seed spread from ${top.displayName}` : "Seed account unidentified",
    `${amplifiers} amplifiers carried it within ${scenario.communityLabels[top?.community ?? 0]}`,
    `Crossed into ${scenario.communityLabels[(top?.community ?? 0) + 1] ?? scenario.communityLabels[0]}`,
    `Reached ${new Set(edges.map((edge) => edge.target)).size} accounts across ${communities.length} communities`,
  ];

  data.networkCache = {
    topicId: data.def.topic.id,
    nodes,
    edges,
    influencers,
    communities,
    spreadPath,
  };
  return data.networkCache;
}

function buildAudience(data: TopicData): AudiencePayload {
  if (data.audienceCache) return data.audienceCache;

  const posts = data.posts;
  const total = Math.max(1, posts.length);
  const languages = new Map<string, number>();
  const regions = new Map<string, number>();
  const interests = new Map<string, number>();
  let broadcasters = 0;
  let amplifiers = 0;
  let participants = 0;
  let critics = 0;
  let advocates = 0;

  for (const post of posts) {
    languages.set(post.language, (languages.get(post.language) ?? 0) + 1);
    regions.set(post.region, (regions.get(post.region) ?? 0) + 1);
    const lower = post.content.toLowerCase();
    for (const rule of INTEREST_RULES) {
      if (rule.words.some((word) => lower.includes(word))) {
        interests.set(rule.label, (interests.get(rule.label) ?? 0) + 1);
      }
    }
    if (post.engagement >= 500) broadcasters += 1;
    else if (post.engagement >= 100) amplifiers += 1;
    else participants += 1;
    if (post.sentimentScore <= -0.2) critics += 1;
    else if (post.sentimentScore >= 0.2) advocates += 1;
  }

  const toShares = (map: Map<string, number>): Record<string, number> => {
    const out: Record<string, number> = {};
    for (const [key, value] of [...map.entries()].sort((a, b) => b[1] - a[1])) {
      out[key] = r2(value / total, 3);
    }
    return out;
  };

  const segments = [
    {
      label: "Broadcasters",
      share: r2(broadcasters / total, 3),
      description: "High-reach accounts (500+ engagements) shaping the dominant narrative.",
    },
    {
      label: "Amplifiers",
      share: r2(amplifiers / total, 3),
      description: "Mid-reach accounts carrying content between communities.",
    },
    {
      label: "Participants",
      share: r2(participants / total, 3),
      description: "Grass-roots accounts replying, quoting and reacting.",
    },
  ];
  if (critics / total >= 0.1) {
    segments.push({
      label: "Critical voices",
      share: r2(critics / total, 3),
      description: "Accounts leaning negative - early warning segment for escalation.",
    });
  }
  if (advocates / total >= 0.1) {
    segments.push({
      label: "Advocates",
      share: r2(advocates / total, 3),
      description: "Accounts leaning positive - counter-narrative and support mobilisation.",
    });
  }

  data.audienceCache = {
    topicId: data.def.topic.id,
    sampleSize: total,
    languages: toShares(languages),
    regions: toShares(regions),
    interests: toShares(interests),
    segments,
  };
  return data.audienceCache;
}

function buildOverview(data: TopicData, range: DateRange): OverviewPayload {
  const cached = data.overviewCache.get(range);
  if (cached) return cached;

  const posts = postsInRange(data, range);
  const span = rangeMs(range);
  const previous = data.posts.filter((post) => {
    const time = new Date(post.timestamp).getTime();
    return time < NOW - span && time >= NOW - 2 * span;
  });

  const totalPosts = posts.length;
  const previousPosts = previous.length;
  const authors = new Set(posts.map((post) => post.authorHandle)).size;
  const previousAuthors = new Set(previous.map((post) => post.authorHandle)).size;
  const sentiment = buildSentiment(data, range);
  const trends = buildTrends(data, range);
  const topTrend = trends.rising[0];

  const payload: OverviewPayload = {
    topicId: data.def.topic.id,
    range,
    totalPosts,
    postsDelta: r2((totalPosts - previousPosts) / Math.max(1, previousPosts), 3),
    activeUsers: authors,
    usersDelta: r2((authors - previousAuthors) / Math.max(1, previousAuthors), 3),
    avgSentiment: sentiment.overallScore,
    sentimentLabel: describeSentiment(sentiment.overallScore),
    trendVelocity: topTrend?.velocity ?? "low",
    velocityScore: topTrend?.score ?? 0,
    engagementTotal: posts.reduce((sum, post) => sum + post.engagement, 0),
    latestPosts: posts.slice(0, 25),
  };
  data.overviewCache.set(range, payload);
  return payload;
}

function buildReport(data: TopicData): Report {
  if (data.reportCache) return data.reportCache;

  const sentiment = buildSentiment(data, "14d");
  const trends = buildTrends(data, "14d");
  const audience = buildAudience(data);
  const network = buildNetwork(data);
  const overview = buildOverview(data, "14d");

  const topEmotion = sentiment.distribution[0];
  const topTrend = trends.rising[0];
  const topInfluencer = network.influencers[0];
  const dominantLanguage = Object.entries(audience.languages)[0];
  const dominantRegion = Object.entries(audience.regions)[0];

  const title = `PulseGraph intelligence report - ${data.def.topic.name}`;
  const executiveSummary =
    `${overview.totalPosts} posts from ${overview.activeUsers} distinct accounts were analysed ` +
    `across the last 14 days, carrying ${overview.engagementTotal.toLocaleString()} engagements. ` +
    `Sentiment is ${overview.sentimentLabel} (${sentiment.overallScore}); ${topEmotion?.emotion ?? "neutral"} ` +
    `dominates the emotional mix at ${Math.round((topEmotion?.share ?? 0) * 100)}%. ` +
    `Discussion volume grew ${Math.round(trends.overallGrowth * 100)}% versus the previous window` +
    `${topTrend ? `, with "${topTrend.keyword}" rising fastest (+${Math.round(topTrend.growth * 100)}%)` : ""}.`;

  const keyObservations = [
    `Top emotion: ${topEmotion?.emotion ?? "neutral"} at ${Math.round((topEmotion?.share ?? 0) * 100)}% share of scored posts.`,
    `${topTrend ? `Fastest-rising keyword "${topTrend.keyword}" (${topTrend.count} mentions, ${topTrend.velocity} velocity).` : "No dominant rising keyword detected."}`,
    `Influence concentrates around ${topInfluencer?.displayName ?? "unknown"} (score ${topInfluencer?.influence ?? 0}).`,
    `${network.communities.length} communities detected; ${network.communities[0]?.label ?? "core cluster"} holds ${network.communities[0]?.size ?? 0} accounts.`,
    `Largest language community: ${dominantLanguage ? `${dominantLanguage[0]} (${Math.round(dominantLanguage[1] * 100)}%)` : "unknown"}; strongest region: ${dominantRegion?.[0] ?? "unknown"}.`,
  ];

  const sections = [
    {
      heading: "Sentiment Analysis",
      body:
        `Across ${sentiment.analyzed} scored posts the average sentiment sits at ${sentiment.overallScore} ` +
        `(${describeSentiment(sentiment.overallScore)}). The distribution is led by ${topEmotion?.emotion ?? "neutral"}, ` +
        `followed by ${sentiment.distribution[1]?.emotion ?? "-"} and ${sentiment.distribution[2]?.emotion ?? "-"}. ` +
        `The sentiment timeline shows ${sentiment.timeline.length} active windows, so shifts are tracked at window granularity.`,
    },
    {
      heading: "Trend Explanation",
      body:
        `Overall conversation volume moved ${Math.round(trends.overallGrowth * 100)}% versus the previous equivalent window. ` +
        `${topTrend ? `"${topTrend.keyword}" leads the rising set with ${topTrend.count} mentions and a growth ratio of ${topTrend.growth}.` : ""} ` +
        `Thematic clustering groups the discussion into ${trends.clusters.length} narrative clusters, led by "${trends.clusters[0]?.label ?? "general"}".`,
    },
    {
      heading: "Network Insights",
      body:
        `${network.nodes.length} accounts form the interaction graph with ${network.edges.length} weighted edges. ` +
        `Top influence scores belong to ${network.influencers.slice(0, 3).map((item) => item.displayName).join(", ")}. ` +
        network.spreadPath.join(". ") + ".",
    },
    {
      heading: "Audience Insights",
      body:
        `The audience sample covers ${audience.sampleSize} posts with ${Object.keys(audience.languages).length} languages and ` +
        `${Object.keys(audience.regions).length} regions represented. ${audience.segments[0]?.label ?? "Broadcasters"} account for ` +
        `${Math.round((audience.segments[0]?.share ?? 0) * 100)}% of activity. Top interest cluster: ${Object.keys(audience.interests)[0] ?? "general"}.`,
    },
  ];

  const scenario = data.def.topic.scenario;
  const recommendedActions =
    scenario === "misinformation"
      ? [
          "Prioritise counter-messaging where rumour propagation clusters overlap relief networks.",
          "Publish verified update cadence to reduce the fear share in the next window.",
          "Monitor critical voices segment for escalation triggers before they cross communities.",
        ]
      : scenario === "event"
        ? [
            "Track route/logistics keywords hourly on event day and pre-empt traffic complaints.",
            "Surface participant-positive content to sustain momentum; flag access-pass confusion early.",
            "Brief local media cluster accounts with the official timeline to reduce rumour friction.",
          ]
        : [
            "Watch the pricing narrative: critical share rises when price leads the trend set.",
            "Amplify hands-on review content; it correlates with positive sentiment bursts.",
            "Track community 'Critics' across launch week for early supply or quality complaints.",
          ];

  data.reportCache = {
    topicId: data.def.topic.id,
    title,
    executiveSummary,
    keyObservations,
    sections,
    recommendedActions,
    modelVersion: "mock-analyst-v1",
    generatedAt: iso(NOW),
  };
  return data.reportCache;
}

// ── public mock API ─────────────────────────────────────────────────────────

let liveCounter = 0;

function titleCase(input: string): string {
  return input
    .split(" ")
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export const mock = {
  login(input: { email: string; password: string }): AuthResponse {
    if (input.password.length < 8) {
      throw new ApiError("invalid_credentials", "Email or password is incorrect.");
    }
    if (input.email.trim().toLowerCase() === "blocked@pulsegraph.dev") {
      throw new ApiError("invalid_credentials", "Email or password is incorrect.");
    }
    const user: User = {
      id: 2,
      name: titleCase(input.email.split("@")[0].replace(/[._-]+/g, " ")) || "Demo Analyst",
      email: input.email,
      role: "analyst",
    };
    return {
      token: `mock.${btoa(input.email)}.${NOW}`,
      expiresAt: iso(NOW + DAY),
      user,
    };
  },

  register(input: { name: string; email: string; password: string }): AuthResponse {
    if (input.email.trim().toLowerCase() === "analyst@pulsegraph.dev") {
      throw new ApiError("email_taken", "An account with this email already exists.");
    }
    const user: User = {
      id: 100 + projects.length,
      name: input.name,
      email: input.email,
      role: "analyst",
    };
    return {
      token: `mock.${btoa(input.email)}.${NOW}`,
      expiresAt: iso(NOW + DAY),
      user,
    };
  },

  listProjects(): Project[] {
    return projects.map((project) => ({ ...project }));
  },

  getProject(id: string): Project {
    const project = projects.find((item) => item.id === id);
    if (!project) throw new ApiError("project_not_found", `Project ${id} was not found.`);
    return { ...project };
  },

  createProject(input: CreateProjectInput): Project {
    const id = `custom-${Date.now().toString(36)}`;
    const topicId = `${id}-topic`;
    const def: TopicDef = {
      topic: {
        id: topicId,
        projectId: id,
        name: input.name,
        keywords: input.keywords,
        description: input.description || "Custom monitoring project.",
        scenario: "event",
      },
      languages: [["en", 0.6], ["hi", 0.4]],
      regions: [["Other", 1]],
      postCount: 120,
    };
    topicDefs.push(def);
    const project: Project = {
      id,
      name: input.name,
      description: input.description,
      platforms: input.platforms,
      keywords: input.keywords,
      frequency: input.frequency,
      status: "active",
      topicIds: [topicId],
      lastUpdated: iso(Date.now()),
      createdAt: iso(Date.now()),
    };
    projects.unshift(project);
    return { ...project };
  },

  listTopics(): Topic[] {
    return topicDefs.map((def) => ({ ...def.topic }));
  },

  getTopic(id: string): Topic {
    const def = topicDefById(id);
    if (!def) throw new ApiError("topic_not_found", `Topic ${id} was not found.`);
    return { ...def.topic };
  },

  getPosts(
    topicId: string,
    options?: { limit?: number; offset?: number; search?: string; range?: DateRange },
  ): { items: Post[]; total: number } {
    const data = ensureTopicData(topicId);
    let posts = options?.range ? postsInRange(data, options.range) : data.posts;
    const search = options?.search?.trim().toLowerCase();
    if (search) {
      posts = posts.filter(
        (post) =>
          post.content.toLowerCase().includes(search) ||
          post.author.toLowerCase().includes(search) ||
          post.authorHandle.toLowerCase().includes(search),
      );
    }
    const offset = options?.offset ?? 0;
    const limit = options?.limit ?? 40;
    return { items: posts.slice(offset, offset + limit), total: posts.length };
  },

  getSentiment(topicId: string, range: DateRange): SentimentPayload {
    return buildSentiment(ensureTopicData(topicId), range);
  },

  getTrends(topicId: string, range: DateRange): TrendPayload {
    return buildTrends(ensureTopicData(topicId), range);
  },

  getTimeline(topicId: string, range: DateRange): TimelinePayload {
    return buildTimeline(ensureTopicData(topicId), range);
  },

  getAudience(topicId: string): AudiencePayload {
    return buildAudience(ensureTopicData(topicId));
  },

  getNetwork(topicId: string): NetworkPayload {
    return buildNetwork(ensureTopicData(topicId));
  },

  getReport(topicId: string): Report {
    return buildReport(ensureTopicData(topicId));
  },

  getOverview(topicId: string, range: DateRange): OverviewPayload {
    return buildOverview(ensureTopicData(topicId), range);
  },

  simulateLivePost(topicId: string): Post {
    const data = ensureTopicData(topicId);
    const rng = mulberry32((Date.now() ^ (liveCounter * 2654435761)) >>> 0);
    liveCounter += 1;
    const project = projectForTopic(topicId);
    const platforms = project?.platforms ?? ["twitter"];
    const authors = AUTHORS.filter((author) => platforms.includes(author.platform));
    const author = pick(rng, authors);
    const scenario = SCENARIOS[data.def.topic.scenario];
    const bucket = bucketFor(arcScore(data.def.topic.scenario, 0.95) + between(rng, -0.2, 0.2));
    const score = clampScore(arcScore(data.def.topic.scenario, 0.95) + between(rng, -0.2, 0.2));

    return {
      id: `live-${Date.now()}-${liveCounter}`,
      topicId,
      platform: author.platform,
      author: author.name,
      authorHandle: author.handle,
      content: fillTemplate(pick(rng, scenario.templates[bucket]), data.def.topic.keywords, rng),
      language: weighted(rng, data.def.languages),
      region: weighted(rng, data.def.regions),
      timestamp: new Date().toISOString(),
      engagement: Math.round(between(rng, 0, 45)),
      emotion: pick(rng, scenario.emotions[bucket]),
      sentimentScore: r2(score, 3),
      confidence: r2(between(rng, 0.55, 0.92), 3),
    };
  },
};

export type MockApi = typeof mock;
