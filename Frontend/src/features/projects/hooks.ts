import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { projectAPI, topicAPI } from "@/services/endpoints";
import type { CreateProjectInput } from "@/types";

export function useProjects() {
  return useQuery({
    queryKey: ["projects"],
    queryFn: () => projectAPI.list(),
  });
}

export function useProject(projectId?: string) {
  return useQuery({
    queryKey: ["projects", projectId],
    queryFn: () => projectAPI.get(projectId as string),
    enabled: Boolean(projectId),
  });
}

export function useTopics() {
  return useQuery({
    queryKey: ["topics"],
    queryFn: () => topicAPI.list(),
  });
}

export function useCreateProject() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateProjectInput) => projectAPI.create(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}
