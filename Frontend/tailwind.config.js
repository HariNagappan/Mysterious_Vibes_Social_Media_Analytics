/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        base: "#0A0E13",
        panel: "#0F141C",
        panel2: "#131A24",
        line: "rgba(255,255,255,0.08)",
        lineStrong: "rgba(255,255,255,0.14)",
        ink: "#E6EDF3",
        muted: "#8B98A5",
        faint: "#5B6673",
        accent: {
          DEFAULT: "#2DD4BF",
          soft: "rgba(45,212,191,0.12)",
          deep: "#14B8A6",
        },
        pos: "#34D399",
        neg: "#F87171",
        warn: "#FBBF24",
        info: "#60A5FA",
      },
      fontFamily: {
        sans: [
          "ui-sans-serif",
          "system-ui",
          "-apple-system",
          "Segoe UI",
          "Roboto",
          "Helvetica Neue",
          "Arial",
          "sans-serif",
        ],
        mono: [
          "ui-monospace",
          "SFMono-Regular",
          "Cascadia Mono",
          "Consolas",
          "Menlo",
          "monospace",
        ],
      },
      keyframes: {
        "fade-up": {
          "0%": { opacity: "0", transform: "translateY(8px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        "pulse-dot": {
          "0%,100%": { opacity: "1" },
          "50%": { opacity: "0.3" },
        },
      },
      animation: {
        "fade-up": "fade-up 0.5s ease-out both",
        "pulse-dot": "pulse-dot 1.6s ease-in-out infinite",
      },
    },
  },
  plugins: [],
};
