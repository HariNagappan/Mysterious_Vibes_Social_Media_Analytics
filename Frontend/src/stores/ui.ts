import { create } from "zustand";

import type { DateRange } from "@/types";

interface UiState {
  sidebarCollapsed: boolean;
  range: DateRange;
  search: string;
  toggleSidebar: () => void;
  setRange: (range: DateRange) => void;
  setSearch: (search: string) => void;
}

/** Small global UI store: sidebar state + dashboard filters. */
export const useUiStore = create<UiState>((set) => ({
  sidebarCollapsed: false,
  range: "14d",
  search: "",
  toggleSidebar: () =>
    set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
  setRange: (range) => set({ range }),
  setSearch: (search) => set({ search }),
}));
