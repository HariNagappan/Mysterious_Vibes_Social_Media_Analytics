import { useParams } from "react-router-dom";

import { TimelineView } from "@/components/timeline/TimelineView";

export function TimelinePage() {
  const { projectId = "" } = useParams();
  return <TimelineView topicId={projectId} />;
}
