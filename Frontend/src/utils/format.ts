/** Compact number formatting, e.g. 12_345 -> "12.3K". */
export function compactNumber(value: number): string {
  return new Intl.NumberFormat("en", {
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(value);
}

/** 0.421 -> "42%" */
export function percent(value: number, digits = 0): string {
  return `${(value * 100).toFixed(digits)}%`;
}

/** 0.421 -> "+42%"; -0.5 -> "-50%" (expects a ratio). */
export function signedPercent(value: number, digits = 0): string {
  const sign = value > 0 ? "+" : "";
  return `${sign}${(value * 100).toFixed(digits)}%`;
}

/** Display a growth ratio like 2.4 as "+240%". */
export function growthLabel(ratio: number): string {
  return signedPercent(ratio, ratio >= 1 ? 0 : 1);
}

export function formatDateTime(input: string | number | Date): string {
  const date = new Date(input);
  return date.toLocaleString("en-GB", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
}

export function formatDay(input: string | number | Date): string {
  const date = new Date(input);
  return date.toLocaleDateString("en-GB", { day: "2-digit", month: "short" });
}

export function formatTime(input: string | number | Date): string {
  const date = new Date(input);
  return date.toLocaleTimeString("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
}

export function timeAgo(input: string | number | Date): string {
  const date = new Date(input).getTime();
  const seconds = Math.max(1, Math.floor((Date.now() - date) / 1000));
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

export function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

export function round(value: number, digits = 2): number {
  const factor = 10 ** digits;
  return Math.round(value * factor) / factor;
}
