import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { postsAPI } from "@/services/endpoints";
import { createLiveSocket } from "@/services/websocket";
import type { Post } from "@/types";

/** Live activity feed: seeded from history, then fed by the live socket. */
export function useLiveFeed(topicId?: string, capacity = 36) {
  const seed = useQuery({
    queryKey: ["posts", topicId, "live-seed"],
    queryFn: () => postsAPI.list(topicId as string, { limit: capacity }),
    enabled: Boolean(topicId),
  });

  const [live, setLive] = useState<Post[]>([]);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    if (!topicId) return;
    setLive([]);
    const socket = createLiveSocket(topicId);
    setConnected(true);
    const unsubscribe = socket.subscribe((event) => {
      if (event.type === "post") {
        const post = event.payload as Post;
        setLive((previous) => [post, ...previous].slice(0, capacity));
      }
    });
    return () => {
      unsubscribe();
      socket.close();
      setConnected(false);
    };
  }, [topicId, capacity]);

  const posts = useMemo(() => {
    const seen = new Set<string>();
    const merged: Post[] = [];
    for (const post of [...live, ...(seed.data?.items ?? [])]) {
      if (seen.has(post.id)) continue;
      seen.add(post.id);
      merged.push(post);
    }
    return merged.slice(0, capacity);
  }, [live, seed.data, capacity]);

  return {
    posts,
    connected,
    isLoading: seed.isPending && posts.length === 0,
  };
}
