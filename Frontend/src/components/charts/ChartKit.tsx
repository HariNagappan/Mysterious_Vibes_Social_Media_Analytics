import { useId } from "react";
import type { ReactElement, ReactNode } from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { CHART, EMOTION_META } from "@/app/constants";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import type { Emotion } from "@/types";
import { cn } from "@/utils/cn";
import { formatDateTime, formatDay } from "@/utils/format";

export const tooltipStyle = {
  backgroundColor: CHART.tooltipBg,
  border: `1px solid ${CHART.tooltipBorder}`,
  borderRadius: 8,
  fontSize: 12,
  color: "#E6EDF3",
};

export function withAlpha(hex: string, alpha: number): string {
  const value = hex.replace("#", "");
  const r = parseInt(value.slice(0, 2), 16);
  const g = parseInt(value.slice(2, 4), 16);
  const b = parseInt(value.slice(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

export function ChartCard({
  title,
  kicker,
  action,
  height = 250,
  className,
  children,
}: {
  title: string;
  kicker?: string;
  action?: ReactNode;
  height?: number;
  className?: string;
  children: ReactElement;
}) {
  return (
    <Card className={className}>
      <CardHeader>
        <div>
          {kicker ? <p className="label-caps mb-1">{kicker}</p> : null}
          <CardTitle>{title}</CardTitle>
        </div>
        {action}
      </CardHeader>
      <div className="px-2 pb-3 pt-4" style={{ height }}>
        <ResponsiveContainer width="100%" height="100%">
          {children}
        </ResponsiveContainer>
      </div>
    </Card>
  );
}

export function SentimentAreaChart({
  data,
}: {
  data: { bucket: string; avgScore: number }[];
}) {
  const gradientId = useId();
  return (
    <AreaChart data={data} margin={{ top: 4, right: 10, bottom: 0, left: -14 }}>
      <defs>
        <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={CHART.accent} stopOpacity={0.32} />
          <stop offset="100%" stopColor={CHART.accent} stopOpacity={0} />
        </linearGradient>
      </defs>
      <CartesianGrid stroke={CHART.grid} vertical={false} />
      <XAxis
        dataKey="bucket"
        tickFormatter={(value) => formatDay(String(value))}
        tick={{ fill: CHART.axis, fontSize: 10 }}
        tickLine={false}
        axisLine={false}
        minTickGap={36}
      />
      <YAxis
        domain={[-1, 1]}
        tick={{ fill: CHART.axis, fontSize: 10 }}
        tickLine={false}
        axisLine={false}
        width={40}
      />
      <Tooltip
        contentStyle={tooltipStyle}
        labelFormatter={(value) => formatDateTime(String(value))}
        formatter={(value) => [Number(value).toFixed(2), "Sentiment"]}
      />
      <Area
        type="monotone"
        dataKey="avgScore"
        stroke={CHART.accent}
        strokeWidth={2}
        fill={`url(#${gradientId})`}
      />
    </AreaChart>
  );
}

export function GrowthLineChart({
  data,
  dataKey = "posts",
  label = "Posts",
  color = CHART.accent,
}: {
  data: Record<string, string | number>[];
  dataKey?: string;
  label?: string;
  color?: string;
}) {
  return (
    <LineChart data={data} margin={{ top: 4, right: 10, bottom: 0, left: -18 }}>
      <CartesianGrid stroke={CHART.grid} vertical={false} />
      <XAxis
        dataKey="bucket"
        tickFormatter={(value) => formatDay(String(value))}
        tick={{ fill: CHART.axis, fontSize: 10 }}
        tickLine={false}
        axisLine={false}
        minTickGap={36}
      />
      <YAxis
        tick={{ fill: CHART.axis, fontSize: 10 }}
        tickLine={false}
        axisLine={false}
        width={44}
      />
      <Tooltip
        contentStyle={tooltipStyle}
        labelFormatter={(value) => formatDateTime(String(value))}
        formatter={(value) => [value as number, label]}
      />
      <Line
        type="monotone"
        dataKey={dataKey}
        stroke={color}
        strokeWidth={2}
        dot={false}
      />
    </LineChart>
  );
}

export function KeywordTrendChart({
  series,
}: {
  series: { keyword: string; points: { bucket: string; value: number }[] }[];
}) {
  if (series.length === 0) {
    return (
      <div className="flex h-full items-center justify-center text-xs text-faint">
        No keyword series available.
      </div>
    );
  }
  const length = Math.max(...series.map((item) => item.points.length));
  const rows: Record<string, string | number>[] = [];
  for (let index = 0; index < length; index++) {
    const row: Record<string, string | number> = {
      bucket: series[0].points[index]?.bucket ?? String(index),
    };
    for (const item of series) {
      row[item.keyword] = item.points[index]?.value ?? 0;
    }
    rows.push(row);
  }
  return (
    <LineChart data={rows} margin={{ top: 4, right: 10, bottom: 0, left: -18 }}>
      <CartesianGrid stroke={CHART.grid} vertical={false} />
      <XAxis
        dataKey="bucket"
        tickFormatter={(value) => formatDay(String(value))}
        tick={{ fill: CHART.axis, fontSize: 10 }}
        tickLine={false}
        axisLine={false}
        minTickGap={36}
      />
      <YAxis
        tick={{ fill: CHART.axis, fontSize: 10 }}
        tickLine={false}
        axisLine={false}
        width={44}
      />
      <Tooltip
        contentStyle={tooltipStyle}
        labelFormatter={(value) => formatDateTime(String(value))}
      />
      <Legend wrapperStyle={{ fontSize: 11, color: "#8B98A5" }} />
      {series.map((item, index) => (
        <Line
          key={item.keyword}
          type="monotone"
          dataKey={item.keyword}
          stroke={CHART.series[index % CHART.series.length]}
          strokeWidth={2}
          dot={false}
        />
      ))}
    </LineChart>
  );
}

export function EmotionPieChart({
  slices,
}: {
  slices: { emotion: Emotion; share: number }[];
}) {
  return (
    <PieChart>
      <Tooltip
        contentStyle={tooltipStyle}
        formatter={(value) => `${Math.round(Number(value) * 100)}%`}
      />
      <Pie
        data={slices}
        dataKey="share"
        nameKey="emotion"
        innerRadius="52%"
        outerRadius="82%"
        paddingAngle={2}
        stroke="none"
      >
        {slices.map((slice) => (
          <Cell key={slice.emotion} fill={EMOTION_META[slice.emotion].color} />
        ))}
      </Pie>
      <Legend
        formatter={(value) => EMOTION_META[value as Emotion]?.label ?? String(value)}
        wrapperStyle={{ fontSize: 11, color: "#8B98A5" }}
      />
    </PieChart>
  );
}

export function DistributionBarChart({
  data,
  color = CHART.accent,
  valueFormatter = (value: number) => `${Math.round(value * 100)}%`,
}: {
  data: { name: string; value: number }[];
  color?: string;
  valueFormatter?: (value: number) => string;
}) {
  return (
    <BarChart data={data} layout="vertical" margin={{ left: 8, right: 18 }}>
      <CartesianGrid stroke={CHART.grid} horizontal={false} />
      <XAxis
        type="number"
        tick={{ fill: CHART.axis, fontSize: 10 }}
        tickLine={false}
        axisLine={false}
        tickFormatter={(value) => valueFormatter(Number(value))}
      />
      <YAxis
        type="category"
        dataKey="name"
        width={118}
        tick={{ fill: CHART.axis, fontSize: 11 }}
        tickLine={false}
        axisLine={false}
      />
      <Tooltip
        contentStyle={tooltipStyle}
        cursor={{ fill: "rgba(255,255,255,0.04)" }}
        formatter={(value) => valueFormatter(Number(value))}
      />
      <Bar dataKey="value" fill={color} radius={[0, 3, 3, 0]} barSize={14} />
    </BarChart>
  );
}

export function HeatMap({
  columns,
  rows,
}: {
  columns: string[];
  rows: { label: string; color: string; values: number[] }[];
}) {
  const max = Math.max(1, ...rows.flatMap((row) => row.values));
  return (
    <div className="overflow-x-auto">
      <div className="min-w-[520px]">
        <div className="mb-2 flex gap-1 pl-[110px]">
          {columns.map((column, index) => (
            <div
              key={`${column}-${index}`}
              className="flex-1 text-center font-mono text-[9px] uppercase tracking-wider text-faint"
            >
              {column}
            </div>
          ))}
        </div>
        {rows.map((row) => (
          <div key={row.label} className="mb-1 flex items-center gap-1">
            <div className="w-[106px] shrink-0 pr-2 text-right text-[11px] text-muted">
              {row.label}
            </div>
            {row.values.map((value, index) => (
              <div
                key={index}
                title={`${row.label}: ${value}`}
                className="h-6 flex-1 rounded-[3px] border border-line"
                style={{
                  backgroundColor:
                    value === 0
                      ? "rgba(255,255,255,0.02)"
                      : withAlpha(row.color, 0.12 + (value / max) * 0.75),
                }}
              />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

export function Sparkline({
  values,
  stroke = CHART.accent,
  className,
}: {
  values: number[];
  stroke?: string;
  className?: string;
}) {
  if (values.length < 2) {
    return <div className={cn("h-7 w-full", className)} />;
  }
  const max = Math.max(...values, 1);
  const points = values
    .map((value, index) => {
      const x = (index / (values.length - 1)) * 100;
      const y = 26 - (value / max) * 22;
      return `${x},${y}`;
    })
    .join(" ");
  return (
    <svg
      viewBox="0 0 100 28"
      preserveAspectRatio="none"
      className={cn("h-7 w-full", className)}
      aria-hidden
    >
      <polyline
        points={points}
        fill="none"
        stroke={stroke}
        strokeWidth={1.6}
        vectorEffect="non-scaling-stroke"
      />
    </svg>
  );
}
