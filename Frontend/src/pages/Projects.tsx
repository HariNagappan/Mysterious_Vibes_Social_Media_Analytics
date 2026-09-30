import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Plus, Radar } from "lucide-react";

import { APP_NAME } from "@/app/constants";
import { EmptyState, ErrorState, Loader } from "@/components/common/States";
import { PlatformBadge } from "@/components/dashboard/Widgets";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Dialog } from "@/components/ui/dialog";
import { StatusBadge } from "@/components/ui/badge";
import { useAuthStore } from "@/features/auth/store";
import { CreateProjectForm } from "@/features/projects/components/CreateProjectForm";
import { useProjects } from "@/features/projects/hooks";
import { timeAgo } from "@/utils/format";

export function ProjectsPage() {
  const projects = useProjects();
  const [dialogOpen, setDialogOpen] = useState(false);
  const user = useAuthStore((state) => state.user);
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-base">
      <header className="border-b border-line bg-panel/40">
        <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
          <Link to="/" className="focus-ring flex items-center gap-2.5 rounded-md">
            <span className="flex h-7 w-7 items-center justify-center rounded-md bg-accent/10">
              <Radar className="h-4 w-4 text-accent" />
            </span>
            <span className="text-sm font-medium text-ink">{APP_NAME}</span>
          </Link>
          <div className="flex items-center gap-3">
            <span className="hidden text-xs text-muted sm:block">
              {user?.name ?? "Analyst"}
            </span>
            <Button size="sm" onClick={() => setDialogOpen(true)}>
              <Plus className="h-3.5 w-3.5" />
              New project
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-4 py-8">
        <div className="mb-6 flex items-end justify-between">
          <div>
            <p className="label-caps mb-1.5">Monitoring workspace</p>
            <h1 className="text-xl text-ink">Projects</h1>
          </div>
          <p className="hidden text-xs text-muted sm:block">
            Each project tracks one or more monitored conversations.
          </p>
        </div>

        {projects.isPending ? (
          <Loader label="Loading projects..." />
        ) : projects.isError ? (
          <ErrorState
            description="Projects could not be loaded."
            onRetry={() => void projects.refetch()}
          />
        ) : (projects.data ?? []).length === 0 ? (
          <EmptyState
            title="No projects yet"
            description="Create a monitoring project to start collecting conversations."
            action={
              <Button size="sm" onClick={() => setDialogOpen(true)}>
                Create project
              </Button>
            }
          />
        ) : (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {(projects.data ?? []).map((project) => (
              <button
                key={project.id}
                onClick={() => navigate(`/projects/${project.id}`)}
                className="focus-ring group text-left"
              >
                <Card className="h-full px-5 py-5 transition-colors group-hover:border-lineStrong">
                  <div className="flex items-start justify-between gap-3">
                    <p className="text-sm font-medium text-ink">
                      {project.name}
                    </p>
                    <StatusBadge status={project.status} />
                  </div>
                  <p className="mt-2 line-clamp-2 text-xs leading-relaxed text-muted">
                    {project.description}
                  </p>
                  <div className="mt-4 flex items-center gap-1.5">
                    {project.platforms.map((platform) => (
                      <PlatformBadge key={platform} platform={platform} />
                    ))}
                  </div>
                  <div className="mt-4 flex flex-wrap gap-1.5">
                    {project.keywords.slice(0, 4).map((keyword) => (
                      <span
                        key={keyword}
                        className="rounded-[4px] border border-line bg-white/5 px-1.5 py-0.5 font-mono text-[10px] text-faint"
                      >
                        #{keyword}
                      </span>
                    ))}
                  </div>
                  <div className="mt-4 flex items-center justify-between border-t border-line pt-3 font-mono text-[10px] text-faint">
                    <span>
                      {project.topicIds.length} monitored topic
                      {project.topicIds.length === 1 ? "" : "s"}
                    </span>
                    <span>updated {timeAgo(project.lastUpdated)}</span>
                  </div>
                </Card>
              </button>
            ))}
          </div>
        )}
      </main>

      <Dialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        title="New monitoring project"
      >
        <CreateProjectForm
          onCreated={(projectId) => {
            setDialogOpen(false);
            navigate(`/projects/${projectId}`);
          }}
        />
      </Dialog>
    </div>
  );
}
