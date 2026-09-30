import { create } from "zustand";
import { persist } from "zustand/middleware";

import { storeToken } from "@/services/api";
import type { User } from "@/types";

interface AuthState {
  token: string | null;
  user: User | null;
  setSession: (token: string, user: User) => void;
  logout: () => void;
}

/** Persisted auth session (JWT + user profile). */
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      setSession: (token, user) => {
        storeToken(token);
        set({ token, user });
      },
      logout: () => {
        storeToken(null);
        set({ token: null, user: null });
      },
    }),
    { name: "pulsegraph.auth" },
  ),
);

export function useIsAuthenticated(): boolean {
  return useAuthStore((state) => Boolean(state.token));
}
