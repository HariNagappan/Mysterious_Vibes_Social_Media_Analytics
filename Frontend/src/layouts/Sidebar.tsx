import { NavLink, useParams } from "react-router-dom";
import { ArrowLeft, Radar } from "lucide-react";

import { APP_NAME, APP_TAGLINE, PROJECT_SECTIONS } from "@/app/constants";
import { useProject } from "@/features/projects/hooks";
import { useUiStore } from "@/stores/ui";
import { cn } from "@/utils/cn";

export function Sidebar() {
  const { projectId } = useParams();
  const collapsed = useUiStore((state) => state.sidebarCollapsed);
  const { data: project } = useProject(projectId);

  return (
    <aside
      className={cn(
        "no-print flex h-full shrink-0 flex-col border-r border-line bg-panel/60 transition-[width] duration-200",
        collapsed ? "w-[64px]" : "w-[232px]",
      )}
    >
      <div className="flex h-14 items-center gap-2.5 border-b border-line px-4">
        <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-accent/15">
          <Radar className="h-4 w-4 text-accent" />
        </span>
        {!collapsed ? (
          <div className="min-w-0">
            <p className="truncate text-sm font-medium text-ink">{APP_NAME}</p>
            <p className="truncate font-mono text-[9px] uppercase tracking-[0.14em] text-faint">
              {APP_TAGLINE}
            </p>
          </div>
        ) : null}
      </div>

      <div className="border-b border-line px-3 py-3">
        <NavLink
          to="/projects"
          className="focus-ring flex items-center gap-2 rounded-md px-2 py-1.5 text-xs text-muted transition-colors hover:bg-white/5 hover:text-ink"
        >
          <ArrowLeft className="h-3.5 w-3.5 shrink-0" />
          {!collapsed ? <span>All projects</span> : null}
        </NavLink>
        {!collapsed && project ? (
          <p className="mt-1 truncate px-2 text-[11px] text-faint">
            {project.name}
          </p>
        ) : null}
      </div>

      <nav className="flex-1 space-y-0.5 overflow-y-auto px-2 py-3">
        {PROJECT_SECTIONS.map((item) => (
          <NavLink
            key={item.label}
            to={item.to}
            end={item.end}
            className={({ isActive }) =>
              cn(
                "focus-ring flex items-center gap-2.5 rounded-md px-2.5 py-2 text-xs transition-colors",
                isActive
                  ? "bg-accent/10 text-accent"
                  : "text-muted hover:bg-white/5 hover:text-ink",
              )
            }
          >
            <item.icon className="h-4 w-4 shrink-0" />
            {!collapsed ? <span>{item.label}</span> : null}
          </NavLink>
        ))}
      </nav>

      <div className="border-t border-line px-3 py-3">
        {!collapsed ? (
          <p className="font-mono text-[9px] uppercase tracking-[0.12em] text-faint">
            SIH 2026 · PS 26152
          </p>
        ) : null}
      </div>
    </aside>
  );
}
