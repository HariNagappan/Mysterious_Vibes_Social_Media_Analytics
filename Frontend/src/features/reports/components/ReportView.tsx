import { useState } from "react";
import {
  CheckCircle2,
  ClipboardCopy,
  Download,
  FileText,
  Sparkles,
} from "lucide-react";

import { ErrorState, Loader } from "@/components/common/States";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { useReport } from "@/hooks/useIntelligence";
import type { Report } from "@/types";
import { formatDateTime } from "@/utils/format";

function buildPlainText(report: Report): string {
  const lines: string[] = [
    report.title,
    "",
    "Executive summary",
    report.executiveSummary,
    "",
    "Key observations",
    ...report.keyObservations.map((observation) => `- ${observation}`),
    "",
  ];
  for (const section of report.sections) {
    lines.push(section.heading, section.body, "");
  }
  lines.push(
    "Recommended actions",
    ...report.recommendedActions.map((action, index) => `${index + 1}. ${action}`),
  );
  return lines.join("\n");
}

export function ReportView({ topicId }: { topicId: string }) {
  const { data, isPending, isError, refetch } = useReport(topicId);
  const [copied, setCopied] = useState(false);

  if (isPending) {
    return <Loader label="Generating the intelligence report..." />;
  }
  if (isError || !data) {
    return (
      <ErrorState
        description="The AI report could not be generated."
        onRetry={() => void refetch()}
      />
    );
  }

  const onShare = async () => {
    try {
      await navigator.clipboard.writeText(buildPlainText(data));
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2200);
    } catch {
      setCopied(false);
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <div className="flex flex-wrap items-start justify-between gap-3 px-5 py-5">
          <div>
            <p className="label-caps mb-2">AI report · {data.modelVersion}</p>
            <h2 className="text-lg text-ink">{data.title}</h2>
            <p className="mt-1 font-mono text-[10px] text-faint">
              Generated {formatDateTime(data.generatedAt)} · grounded in
              analysed metrics only
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button variant="secondary" size="sm" onClick={() => window.print()}>
              <Download className="h-3.5 w-3.5" />
              Export PDF
            </Button>
            <Button variant="secondary" size="sm" onClick={() => void onShare()}>
              {copied ? (
                <CheckCircle2 className="h-3.5 w-3.5 text-pos" />
              ) : (
                <ClipboardCopy className="h-3.5 w-3.5" />
              )}
              {copied ? "Copied" : "Share Report"}
            </Button>
          </div>
        </div>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Executive summary</CardTitle>
          <Sparkles className="h-3.5 w-3.5 text-accent" />
        </CardHeader>
        <div className="px-5 py-4">
          <p className="text-sm leading-relaxed text-ink">
            {data.executiveSummary}
          </p>
        </div>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Key observations</CardTitle>
          <Badge>{data.keyObservations.length}</Badge>
        </CardHeader>
        <ul className="space-y-3 px-5 py-4">
          {data.keyObservations.map((observation, index) => (
            <li
              key={index}
              className="flex gap-2.5 text-xs leading-relaxed text-muted"
            >
              <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-accent" />
              {observation}
            </li>
          ))}
        </ul>
      </Card>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
        {data.sections.map((section) => (
          <Card key={section.heading}>
            <CardHeader>
              <CardTitle>{section.heading}</CardTitle>
            </CardHeader>
            <div className="px-5 py-4">
              <p className="text-xs leading-relaxed text-muted">
                {section.body}
              </p>
            </div>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recommended actions</CardTitle>
          <FileText className="h-3.5 w-3.5 text-accent" />
        </CardHeader>
        <ol className="space-y-3 px-5 py-4">
          {data.recommendedActions.map((action, index) => (
            <li
              key={index}
              className="flex gap-3 text-xs leading-relaxed text-muted"
            >
              <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full border border-line bg-panel2 font-mono text-[10px] text-accent">
                {index + 1}
              </span>
              {action}
            </li>
          ))}
        </ol>
      </Card>

      <p className="text-center text-[10px] leading-relaxed text-faint">
        This report explains existing computed metrics only - every figure
        above derives from the analysed post set. No facts are invented.
      </p>
    </div>
  );
}
