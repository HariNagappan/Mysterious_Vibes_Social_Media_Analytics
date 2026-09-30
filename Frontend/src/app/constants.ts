import {
  Activity,
  Clock,
  FileText,
  LayoutDashboard,
  Network,
  Smile,
  TrendingUp,
  Users,
  type LucideIcon,
} from "lucide-react";

import type { Emotion, Platform } from "@/types";

export const APP_NAME = "PulseGraph";
export const APP_TAGLINE = "AI Social Intelligence";

export interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
  end?: boolean;
}

/** Sidebar sections inside a project workspace. */
export const PROJECT_SECTIONS: NavItem[] = [
  { to: "", label: "Overview", icon: LayoutDashboard, end: true },
  { to: "analytics", label: "Analytics", icon: Activity },
  { to: "timeline", label: "Timeline", icon: Clock },
  { to: "sentiment", label: "Sentiment", icon: Smile },
  { to: "audience", label: "Audience", icon: Users },
  { to: "trends", label: "Trends", icon: TrendingUp },
  { to: "network", label: "Network", icon: Network },
  { to: "reports", label: "AI Report", icon: FileText },
];

export const PLATFORM_META: Record<
  Platform,
  { label: string; short: string; color: string }
> = {
  twitter: { label: "X (Twitter)", short: "X", color: "#8AB4FF" },
  telegram: { label: "Telegram", short: "TG", color: "#4FC3F7" },
  instagram: { label: "Instagram", short: "IG", color: "#F472B6" },
  facebook: { label: "Facebook", short: "FB", color: "#818CF8" },
  reddit: { label: "Reddit", short: "RD", color: "#FB923C" },
  youtube: { label: "YouTube", short: "YT", color: "#F87171" },
};

/** Emotion palette + analyst-facing labels (stance framing). */
export const EMOTION_META: Record<Emotion, { label: string; color: string }> = {
  joy: { label: "Support", color: "#34D399" },
  surprise: { label: "Excitement", color: "#FBBF24" },
  neutral: { label: "Neutral", color: "#8B98A5" },
  fear: { label: "Anxiety", color: "#F87171" },
  anger: { label: "Opposition", color: "#FB7185" },
  sadness: { label: "Concern", color: "#A78BFA" },
  disgust: { label: "Disgust", color: "#C084FC" },
};

export const CHART = {
  accent: "#2DD4BF",
  grid: "rgba(255,255,255,0.06)",
  axis: "#5B6673",
  tooltipBg: "#131A24",
  tooltipBorder: "rgba(255,255,255,0.12)",
  series: ["#2DD4BF", "#60A5FA", "#FBBF24", "#F87171", "#A78BFA", "#34D399"],
};

export const DATE_RANGES = [
  { value: "24h", label: "Last 24 hours" },
  { value: "7d", label: "Last 7 days" },
  { value: "14d", label: "Last 14 days" },
] as const;

export type DateRangeValue = (typeof DATE_RANGES)[number]["value"];

export const DEMO_CREDENTIALS = {
  email: "analyst@pulsegraph.dev",
  password: "PulseGraph@2026",
};
