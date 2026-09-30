import { useState } from "react";
import { useParams } from "react-router-dom";

import { NetworkInsightsSide } from "@/components/analytics/panels";
import { ErrorState, Loader } from "@/components/common/States";
import { SectionHeading } from "@/components/dashboard/Widgets";
import { NetworkGraph } from "@/components/graphs/NetworkGraph";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { useNetwork } from "@/hooks/useIntelligence";
import { useUiStore } from "@/stores/ui";

export function NetworkPage() {
  const { projectId = "" } = useParams();
  const search = useUiStore((state) => state.search);
  const [selectedId, setSelectedId] = useState<string | undefined>(undefined);
  const { data, isPending, isError, refetch } = useNetwork(projectId);

  if (isPending) return <Loader label="Mapping the influence network..." />;
  if (isError || !data) {
    return (
      <ErrorState
        description="The interaction graph could not be loaded."
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <div>
      <SectionHeading
        kicker="Influence & propagation"
        title="Network analysis"
        action={
          <Badge>
            {data.nodes.length} nodes · {data.edges.length} edges
          </Badge>
        }
      />
      <div className="grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,2.2fr)_minmax(0,1fr)]">
        <Card className="overflow-hidden">
          <div className="h-[600px]">
            <NetworkGraph
              network={data}
              search={search}
              selectedId={selectedId}
              onSelect={setSelectedId}
            />
          </div>
          <div className="border-t border-line px-4 py-2.5">
            <p className="text-[11px] text-muted">
              Search in the top bar highlights matching accounts. Click a node
              to focus it; drag to rearrange; scroll to zoom.
            </p>
          </div>
        </Card>
        <NetworkInsightsSide
          network={data}
          selectedId={selectedId}
          onSelect={setSelectedId}
        />
      </div>
    </div>
  );
}
