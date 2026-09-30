import { useMemo } from "react";
import {
  Background,
  BackgroundVariant,
  Controls,
  Handle,
  MiniMap,
  Position,
  ReactFlow,
  type Edge,
  type Node,
  type NodeProps,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import { withAlpha } from "@/components/charts/ChartKit";
import type { NetworkPayload } from "@/types";

export const COMMUNITY_COLORS = [
  "#2DD4BF",
  "#60A5FA",
  "#FBBF24",
  "#A78BFA",
  "#F87171",
];

type PulseNode = Node<
  {
    label: string;
    community: number;
    influence: number;
    highlight: boolean;
    selected: boolean;
  },
  "pulse"
>;

function PulseNodeComponent({ data }: NodeProps<PulseNode>) {
  const color = COMMUNITY_COLORS[data.community % COMMUNITY_COLORS.length];
  const size = 44 + Math.round(data.influence * 46);
  return (
    <div className="flex flex-col items-center gap-1">
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      <div
        className="flex items-center justify-center rounded-full border font-mono text-[10px] transition-colors"
        style={{
          width: size,
          height: size,
          borderColor: withAlpha(color, 0.7),
          backgroundColor: withAlpha(
            color,
            data.selected ? 0.42 : data.highlight ? 0.3 : 0.12,
          ),
          color,
          boxShadow: data.selected
            ? `0 0 0 2px ${withAlpha(color, 0.5)}`
            : undefined,
        }}
      >
        {Math.round(data.influence * 100)}
      </div>
      <span className="max-w-[110px] truncate text-center text-[10px] text-muted">
        {data.label}
      </span>
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  );
}

const nodeTypes = { pulse: PulseNodeComponent };

export function NetworkGraph({
  network,
  search,
  selectedId,
  onSelect,
}: {
  network: NetworkPayload;
  search: string;
  selectedId?: string;
  onSelect: (nodeId?: string) => void;
}) {
  const { flowNodes, flowEdges } = useMemo(() => {
    const query = search.trim().toLowerCase();
    const radius = 330;
    const flowNodes: PulseNode[] = network.nodes.map((node, index) => {
      const angle = (index / Math.max(1, network.nodes.length)) * Math.PI * 2;
      const ring = node.community % 2 === 0 ? 1 : 0.62;
      return {
        id: node.id,
        type: "pulse",
        position: {
          x: 520 + Math.cos(angle) * radius * ring,
          y: 330 + Math.sin(angle) * radius * ring,
        },
        data: {
          label: node.label,
          community: node.community,
          influence: node.influence,
          highlight:
            query.length > 0 &&
            (node.label.toLowerCase().includes(query) ||
              node.handle.toLowerCase().includes(query)),
          selected: selectedId === node.id,
        },
      };
    });

    const flowEdges: Edge[] = network.edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      animated: edge.weight >= 7,
      style: {
        stroke: "rgba(139,152,165,0.35)",
        strokeWidth: Math.min(3, 0.6 + edge.weight / 5),
      },
    }));

    return { flowNodes, flowEdges };
  }, [network, search, selectedId]);

  return (
    <ReactFlow
      nodes={flowNodes}
      edges={flowEdges}
      nodeTypes={nodeTypes}
      fitView
      minZoom={0.3}
      maxZoom={2}
      proOptions={{ hideAttribution: true }}
      onNodeClick={(_, node) => onSelect(node.id)}
      onPaneClick={() => onSelect(undefined)}
    >
      <Background
        variant={BackgroundVariant.Dots}
        gap={26}
        size={1}
        color="rgba(255,255,255,0.06)"
      />
      <Controls showInteractive={false} />
      <MiniMap
        pannable
        zoomable
        maskColor="rgba(10,14,19,0.72)"
        nodeColor={(node) => {
          const community = (node.data as { community?: number })?.community ?? 0;
          return COMMUNITY_COLORS[community % COMMUNITY_COLORS.length];
        }}
      />
    </ReactFlow>
  );
}
