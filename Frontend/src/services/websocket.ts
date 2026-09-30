/**
 * Live-data abstraction.
 *
 * When VITE_WS_URL is configured a real WebSocket is used; otherwise a
 * deterministic simulator emits live events (new posts) at a steady cadence
 * so the product feels alive without any backend.
 */

import { mock } from "@/services/mock/engine";
import type { Post } from "@/types";

export type LiveEvent =
  | { type: "post"; payload: Post }
  | { type: "heartbeat"; payload: { at: string } };

export interface LiveSocket {
  subscribe: (handler: (event: LiveEvent) => void) => () => void;
  close: () => void;
}

function createSimulatedSocket(topicId: string): LiveSocket {
  const handlers = new Set<(event: LiveEvent) => void>();
  const timer = window.setInterval(() => {
    if (handlers.size === 0) return;
    const post = mock.simulateLivePost(topicId);
    for (const handler of handlers) {
      handler({ type: "post", payload: post });
    }
  }, 3800);

  return {
    subscribe(handler) {
      handlers.add(handler);
      return () => {
        handlers.delete(handler);
      };
    },
    close() {
      window.clearInterval(timer);
      handlers.clear();
    },
  };
}

function createRealSocket(url: string, topicId: string): LiveSocket {
  const handlers = new Set<(event: LiveEvent) => void>();
  let socket: WebSocket | null = null;
  let closed = false;

  const connect = () => {
    socket = new WebSocket(url);
    socket.onmessage = (message) => {
      try {
        const event = JSON.parse(String(message.data)) as LiveEvent & { topicId?: string };
        if (event.topicId && event.topicId !== topicId) return;
        for (const handler of handlers) {
          handler(event);
        }
      } catch {
        /* ignore malformed frames */
      }
    };
    socket.onclose = () => {
      if (!closed) window.setTimeout(connect, 4000);
    };
  };
  connect();

  return {
    subscribe(handler) {
      handlers.add(handler);
      return () => {
        handlers.delete(handler);
      };
    },
    close() {
      closed = true;
      socket?.close();
      handlers.clear();
    },
  };
}

export function createLiveSocket(topicId: string): LiveSocket {
  const url = (import.meta.env.VITE_WS_URL ?? "").trim();
  if (url) return createRealSocket(url, topicId);
  return createSimulatedSocket(topicId);
}
