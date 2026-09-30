import { Outlet, useNavigate, useParams } from "react-router-dom";
import { Download, LogOut, Menu, Search } from "lucide-react";

import { DATE_RANGES } from "@/app/constants";
import { Button } from "@/components/ui/button";
import { Input, Select } from "@/components/ui/form";
import { useAuthStore } from "@/features/auth/store";
import { useProject } from "@/features/projects/hooks";
import { useUiStore } from "@/stores/ui";
import type { DateRange } from "@/types";

import { Sidebar } from "./Sidebar";

export function DashboardLayout() {
  const { projectId } = useParams();
  const { data: project } = useProject(projectId);
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);
  const search = useUiStore((state) => state.search);
  const setSearch = useUiStore((state) => state.setSearch);
  const range = useUiStore((state) => state.range);
  const setRange = useUiStore((state) => state.setRange);
  const toggleSidebar = useUiStore((state) => state.toggleSidebar);

  return (
    <div className="flex h-screen overflow-hidden bg-base">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="no-print flex h-14 shrink-0 items-center gap-3 border-b border-line bg-panel/40 px-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={toggleSidebar}
            aria-label="Toggle sidebar"
          >
            <Menu className="h-4 w-4" />
          </Button>
          <div className="min-w-0">
            <p className="truncate text-sm font-medium text-ink">
              {project?.name ?? "Loading project..."}
            </p>
            <p className="font-mono text-[10px] uppercase text-faint">
              {project
                ? `${project.platforms.join(" + ")} · ${project.frequency}`
                : ""}
            </p>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <div className="relative hidden md:block">
              <Search className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-faint" />
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="Search posts, authors..."
                className="h-8 w-56 pl-8 text-xs"
              />
            </div>
            <Select
              value={range}
              onChange={(event) =>
                setRange(event.target.value as DateRange)
              }
              className="h-8 w-[138px] text-xs"
            >
              {DATE_RANGES.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </Select>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => window.print()}
            >
              <Download className="h-3.5 w-3.5" />
              Export
            </Button>
            <div className="hidden items-center gap-1 border-l border-line pl-3 lg:flex">
              <span className="max-w-[140px] truncate text-xs text-muted">
                {user?.name ?? "Analyst"}
              </span>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Log out"
                onClick={() => {
                  logout();
                  navigate("/login", { replace: true });
                }}
              >
                <LogOut className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </header>
        <main className="min-h-0 flex-1 overflow-y-auto px-4 py-5 lg:px-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
