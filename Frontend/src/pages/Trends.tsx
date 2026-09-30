import { useParams } from "react-router-dom";

import { TrendsPanel } from "@/components/analytics/panels";
import { SectionHeading } from "@/components/dashboard/Widgets";
import { useUiStore } from "@/stores/ui";

export function TrendsPage() {
  const { projectId = "" } = useParams();
  const range = useUiStore((state) => state.range);

  return (
    <div>
      <SectionHeading kicker="Narrative acceleration" title="Trends" />
      <TrendsPanel topicId={projectId} range={range} />
    </div>
  );
}
