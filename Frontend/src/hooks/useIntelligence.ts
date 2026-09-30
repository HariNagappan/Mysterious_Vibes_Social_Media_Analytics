import { useQuery } from "@tanstack/react-query";

import {
  demographicAPI,
  networkAPI,
  overviewAPI,
  reportAPI,
  sentimentAPI,
  timelineAPI,
  trendAPI,
} from "@/services/endpoints";
import type { DateRange } from "@/types";

const DEFAULT_RANGE: DateRange = "14d";

export function useOverview(topicId?: string, range: DateRange = DEFAULT_RANGE) {
  return useQuery({
    queryKey: ["overview", topicId, range],
    queryFn: () => overviewAPI.forTopic(topicId as string, range),
    enabled: Boolean(topicId),
  });
}

export function useTimeline(topicId?: string, range: DateRange = DEFAULT_RANGE) {
  return useQuery({
    queryKey: ["timeline", topicId, range],
    queryFn: () => timelineAPI.forTopic(topicId as string, range),
    enabled: Boolean(topicId),
  });
}

export function useSentiment(topicId?: string, range: DateRange = DEFAULT_RANGE) {
  return useQuery({
    queryKey: ["sentiment", topicId, range],
    queryFn: () => sentimentAPI.forTopic(topicId as string, range),
    enabled: Boolean(topicId),
  });
}

export function useAudience(topicId?: string) {
  return useQuery({
    queryKey: ["audience", topicId],
    queryFn: () => demographicAPI.forTopic(topicId as string),
    enabled: Boolean(topicId),
  });
}

export function useTrends(topicId?: string, range: DateRange = DEFAULT_RANGE) {
  return useQuery({
    queryKey: ["trends", topicId, range],
    queryFn: () => trendAPI.forTopic(topicId as string, range),
    enabled: Boolean(topicId),
  });
}

export function useNetwork(topicId?: string) {
  return useQuery({
    queryKey: ["network", topicId],
    queryFn: () => networkAPI.forTopic(topicId as string),
    enabled: Boolean(topicId),
  });
}

export function useReport(topicId?: string) {
  return useQuery({
    queryKey: ["report", topicId],
    queryFn: () => reportAPI.forTopic(topicId as string),
    enabled: Boolean(topicId),
  });
}
