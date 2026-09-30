import { useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { PLATFORM_META } from "@/app/constants";
import { Button } from "@/components/ui/button";
import {
  FieldError,
  Input,
  Label,
  Select,
  Textarea,
} from "@/components/ui/form";
import { useCreateProject } from "@/features/projects/hooks";
import type { Platform } from "@/types";
import { cn } from "@/utils/cn";

const schema = z.object({
  name: z.string().min(3, "Name must be at least 3 characters."),
  description: z
    .string()
    .max(300, "Keep the description under 300 characters."),
  keywords: z.string().min(3, "Add at least one keyword."),
  frequency: z.enum(["realtime", "hourly", "daily"]),
});

type FormValues = z.infer<typeof schema>;

const PLATFORM_OPTIONS: Platform[] = [
  "twitter",
  "telegram",
  "instagram",
  "facebook",
  "reddit",
  "youtube",
];

export function CreateProjectForm({
  onCreated,
}: {
  onCreated: (projectId: string) => void;
}) {
  const createProject = useCreateProject();
  const [platforms, setPlatforms] = useState<Platform[]>([
    "twitter",
    "telegram",
  ]);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      description: "",
      keywords: "",
      frequency: "hourly",
    },
  });

  const togglePlatform = (platform: Platform) => {
    setPlatforms((current) =>
      current.includes(platform)
        ? current.filter((item) => item !== platform)
        : [...current, platform],
    );
  };

  const onSubmit = handleSubmit((values) => {
    const keywords = values.keywords
      .split(/[,;\n]+/)
      .map((keyword) => keyword.trim().toLowerCase())
      .filter(Boolean);
    if (keywords.length === 0 || platforms.length === 0) return;
    createProject.mutate(
      {
        name: values.name,
        description: values.description,
        keywords,
        frequency: values.frequency,
        platforms,
      },
      { onSuccess: (project) => onCreated(project.id) },
    );
  });

  return (
    <form onSubmit={onSubmit} className="space-y-4" noValidate>
      <div>
        <Label htmlFor="project-name">Project name</Label>
        <Input
          id="project-name"
          placeholder="Flood Rumour Analysis"
          {...register("name")}
        />
        <FieldError message={errors.name?.message} />
      </div>
      <div>
        <Label htmlFor="project-description">Description</Label>
        <Textarea
          id="project-description"
          rows={3}
          placeholder="What should this project monitor, and why?"
          {...register("description")}
        />
        <FieldError message={errors.description?.message} />
      </div>
      <div>
        <Label>Platforms</Label>
        <div className="grid grid-cols-3 gap-2">
          {PLATFORM_OPTIONS.map((platform) => {
            const active = platforms.includes(platform);
            return (
              <button
                key={platform}
                type="button"
                onClick={() => togglePlatform(platform)}
                className={cn(
                  "focus-ring rounded-md border px-2 py-2 text-xs transition-colors",
                  active
                    ? "border-accent/50 bg-accent/10 text-accent"
                    : "border-line bg-panel2 text-muted hover:border-lineStrong",
                )}
              >
                {PLATFORM_META[platform].label}
              </button>
            );
          })}
        </div>
        {platforms.length === 0 ? (
          <p className="mt-1 text-xs text-neg">
            Select at least one platform.
          </p>
        ) : null}
      </div>
      <div>
        <Label htmlFor="project-keywords">Keywords (comma separated)</Label>
        <Input
          id="project-keywords"
          placeholder="flood, rain, warning, relief"
          {...register("keywords")}
        />
        <FieldError message={errors.keywords?.message} />
      </div>
      <div>
        <Label htmlFor="project-frequency">Monitoring frequency</Label>
        <Select id="project-frequency" {...register("frequency")}>
          <option value="realtime">Realtime</option>
          <option value="hourly">Hourly</option>
          <option value="daily">Daily</option>
        </Select>
      </div>
      {createProject.isError ? (
        <p className="border border-neg/30 bg-neg/10 px-3 py-2 text-xs text-neg">
          {createProject.error instanceof Error
            ? createProject.error.message
            : "Could not create the project."}
        </p>
      ) : null}
      <div className="flex justify-end gap-2 pt-1">
        <Button
          type="submit"
          loading={createProject.isPending}
          disabled={platforms.length === 0}
        >
          Create project
        </Button>
      </div>
    </form>
  );
}
