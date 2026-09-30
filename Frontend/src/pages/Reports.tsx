import { useParams } from "react-router-dom";

import { SectionHeading } from "@/components/dashboard/Widgets";
import { ReportView } from "@/features/reports/components/ReportView";

export function ReportsPage() {
  const { projectId = "" } = useParams();

  return (
    <div>
      <SectionHeading kicker="Grounded intelligence" title="AI Report" />
      <ReportView topicId={projectId} />
    </div>
  );
}
