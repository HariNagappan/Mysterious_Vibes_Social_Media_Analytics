import { useParams } from "react-router-dom";

import { SentimentPanel } from "@/components/analytics/panels";
import { SectionHeading } from "@/components/dashboard/Widgets";
import { useUiStore } from "@/stores/ui";

export function SentimentPage() {
  const { projectId = "" } = useParams();
  const range = useUiStore((state) => state.range);

  return (
    <div>
      <SectionHeading
        kicker="Emotion & stance intelligence"
        title="Sentiment"
      />
      <SentimentPanel topicId={projectId} range={range} />
    </div>
  );
}
