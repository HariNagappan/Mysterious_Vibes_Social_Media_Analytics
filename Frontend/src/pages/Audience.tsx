import { useParams } from "react-router-dom";

import { AudiencePanel } from "@/components/analytics/panels";
import { SectionHeading } from "@/components/dashboard/Widgets";

export function AudiencePage() {
  const { projectId = "" } = useParams();

  return (
    <div>
      <SectionHeading
        kicker="Aggregated audience analytics"
        title="Audience"
      />
      <AudiencePanel topicId={projectId} />
    </div>
  );
}
