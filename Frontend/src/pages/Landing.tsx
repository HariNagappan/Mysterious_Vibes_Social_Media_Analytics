import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import {
  Activity,
  ArrowRight,
  BarChart3,
  BookOpen,
  Check,
  Cpu,
  Database,
  Megaphone,
  Network,
  Radar,
  ShieldAlert,
  Sparkles,
  TrendingUp,
  Users,
} from "lucide-react";

import { APP_NAME } from "@/app/constants";

const FLOW = [
  {
    icon: Database,
    title: "Social media sources",
    detail:
      "X (Twitter) and Telegram today, with a connector plugin model ready for Instagram, Facebook, Reddit and YouTube.",
  },
  {
    icon: Sparkles,
    title: "AI processing engine",
    detail:
      "Language detection, sentiment scoring, embeddings, topic detection and trend calculation run as pipeline stages.",
  },
  {
    icon: BarChart3,
    title: "Intelligence dashboard",
    detail:
      "Timelines, emotions, audiences, influence networks and grounded AI reports assembled for analysts.",
  },
];

const FEATURES = [
  {
    icon: Activity,
    title: "Sentiment Intelligence",
    detail:
      "Emotion distribution, support/opposition patterns and sentiment evolution tracked per conversation.",
  },
  {
    icon: Users,
    title: "Audience Intelligence",
    detail:
      "Aggregated language, regional and interest composition - community segments, never individual profiling.",
  },
  {
    icon: TrendingUp,
    title: "Trend Detection",
    detail:
      "Rising keywords scored by growth and velocity, grouped into narrative clusters as they accelerate.",
  },
  {
    icon: Network,
    title: "Network Analysis",
    detail:
      "Influence ranking, community structure and information spread paths across reply and mention graphs.",
  },
];

const PIPELINE = [
  {
    icon: Database,
    index: "01",
    label: "Data collection",
    detail: "Connectors normalise posts, replies and mentions into a canonical stream.",
  },
  {
    icon: Cpu,
    index: "02",
    label: "AI processing",
    detail: "Sentiment, embeddings and topic detection enrich every post.",
  },
  {
    icon: BarChart3,
    index: "03",
    label: "Analytics",
    detail: "Trends, demographics and graph metrics aggregate continuously.",
  },
  {
    icon: Radar,
    index: "04",
    label: "Visualization",
    detail: "Dashboards, timelines and reports turn metrics into decisions.",
  },
];

const USE_CASES = [
  {
    icon: ShieldAlert,
    title: "Crisis monitoring",
    detail:
      "Track rumour spread against verified information during floods, fires and outages.",
  },
  {
    icon: BookOpen,
    title: "Research",
    detail:
      "Study public opinion shifts with reproducible, aggregated datasets.",
  },
  {
    icon: Megaphone,
    title: "Public communication",
    detail:
      "Measure how official messaging lands and where counter-narratives form.",
  },
  {
    icon: Users,
    title: "Community analysis",
    detail:
      "Understand which communities drive conversations and how narratives travel.",
  },
];

export function LandingPage() {
  return (
    <div className="min-h-screen bg-base">
      <header className="sticky top-0 z-40 border-b border-line bg-base/85 backdrop-blur">
        <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
          <div className="flex items-center gap-2.5">
            <span className="flex h-7 w-7 items-center justify-center rounded-md bg-accent/10">
              <Radar className="h-4 w-4 text-accent" />
            </span>
            <span className="text-sm font-medium text-ink">{APP_NAME}</span>
          </div>
          <nav className="hidden items-center gap-6 text-xs text-muted md:flex">
            <a href="#problem" className="transition-colors hover:text-ink">
              Problem
            </a>
            <a href="#features" className="transition-colors hover:text-ink">
              Features
            </a>
            <a href="#architecture" className="transition-colors hover:text-ink">
              Architecture
            </a>
            <a href="#use-cases" className="transition-colors hover:text-ink">
              Use cases
            </a>
          </nav>
          <div className="flex items-center gap-2">
            <Link
              to="/login"
              className="focus-ring rounded-md px-3 py-1.5 text-xs text-muted transition-colors hover:text-ink"
            >
              Sign in
            </Link>
            <Link
              to="/register"
              className="focus-ring rounded-md bg-accent px-3 py-1.5 text-xs font-medium text-[#04211c] transition-colors hover:bg-accent-deep"
            >
              Get started
            </Link>
          </div>
        </div>
      </header>

      <section className="mx-auto max-w-6xl px-4 pb-16 pt-20">
        <motion.div
          initial={{ opacity: 0, y: 14 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <p className="label-caps mb-4">AI-powered social intelligence</p>
          <h1 className="max-w-3xl text-4xl leading-tight text-ink md:text-5xl md:leading-[1.1]">
            Understand the world&apos;s conversations{" "}
            <span className="text-accent">with AI</span>
          </h1>
          <p className="mt-5 max-w-2xl text-sm leading-relaxed text-muted md:text-base">
            PulseGraph transforms fragmented social media activity into
            actionable intelligence - sentiment, audiences, trends and
            influence networks, in one analytical workspace.
          </p>
          <div className="mt-8 flex flex-wrap items-center gap-3">
            <Link
              to="/register"
              className="focus-ring inline-flex h-10 items-center gap-2 rounded-md bg-accent px-5 text-sm font-medium text-[#04211c] transition-colors hover:bg-accent-deep"
            >
              Open the dashboard
              <ArrowRight className="h-4 w-4" />
            </Link>
            <a
              href="#architecture"
              className="focus-ring inline-flex h-10 items-center gap-2 rounded-md border border-lineStrong px-5 text-sm text-ink transition-colors hover:bg-white/5"
            >
              See the architecture
            </a>
          </div>
          <div className="mt-10 flex flex-wrap gap-x-10 gap-y-3 border-t border-line pt-6 font-mono text-[11px] uppercase tracking-[0.12em] text-faint">
            <span>2 live connectors - 4 planned</span>
            <span>4 intelligence dimensions</span>
            <span>1,000+ demo posts</span>
            <span>Grounded AI reports</span>
          </div>
        </motion.div>

        <div className="mt-16 grid grid-cols-1 gap-4 md:grid-cols-3">
          {FLOW.map((step, index) => (
            <motion.div
              key={step.title}
              initial={{ opacity: 0, y: 16 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: index * 0.12 }}
              className="relative border border-line bg-panel/60 px-5 py-5"
            >
              <div className="flex items-center justify-between">
                <span className="flex h-8 w-8 items-center justify-center rounded-md bg-accent/10">
                  <step.icon className="h-4 w-4 text-accent" />
                </span>
                <span className="font-mono text-[10px] text-faint">
                  0{index + 1}
                </span>
              </div>
              <p className="mt-4 text-sm font-medium text-ink">{step.title}</p>
              <p className="mt-1.5 text-xs leading-relaxed text-muted">
                {step.detail}
              </p>
              {index < FLOW.length - 1 ? (
                <motion.span
                  aria-hidden
                  className="absolute -right-3 top-1/2 hidden h-px w-6 bg-accent/40 md:block"
                  initial={{ scaleX: 0 }}
                  whileInView={{ scaleX: 1 }}
                  viewport={{ once: true }}
                  transition={{ delay: 0.3 + index * 0.15 }}
                />
              ) : null}
            </motion.div>
          ))}
        </div>
      </section>

      <section id="problem" className="border-t border-line bg-panel/30 py-16">
        <div className="mx-auto max-w-6xl px-4">
          <p className="label-caps mb-3">The problem</p>
          <h2 className="max-w-2xl text-2xl leading-snug text-ink">
            Social conversations are fragmented, fast and emotional.
          </h2>
          <div className="mt-8 grid grid-cols-1 gap-6 md:grid-cols-3">
            <div className="border-l-2 border-line pl-4">
              <p className="text-sm leading-relaxed text-muted">
                Signals are scattered across platforms, languages and
                formats - no single analyst can watch them all.
              </p>
            </div>
            <div className="border-l-2 border-line pl-4">
              <p className="text-sm leading-relaxed text-muted">
                Narratives accelerate before teams can read them, and rumour
                outruns verified information.
              </p>
            </div>
            <div className="border-l-2 border-line pl-4">
              <p className="text-sm leading-relaxed text-muted">
                Raw volume hides who influences what, which audiences matter,
                and how information actually moves.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="py-16">
        <div className="mx-auto grid max-w-6xl grid-cols-1 gap-10 px-4 lg:grid-cols-2">
          <div>
            <p className="label-caps mb-3">The solution</p>
            <h2 className="text-2xl leading-snug text-ink">
              One platform for the full intelligence loop.
            </h2>
            <p className="mt-4 max-w-xl text-sm leading-relaxed text-muted">
              PulseGraph ingests public conversations, runs them through an AI
              pipeline and assembles the result into analyst-ready evidence:
            </p>
            <ul className="mt-6 space-y-2.5 text-sm text-muted">
              <li className="flex gap-2">
                <Check className="mt-0.5 h-4 w-4 shrink-0 text-accent" />
                Sentiment and emotion intelligence
              </li>
              <li className="flex gap-2">
                <Check className="mt-0.5 h-4 w-4 shrink-0 text-accent" />
                Audience and language composition (aggregated)
              </li>
              <li className="flex gap-2">
                <Check className="mt-0.5 h-4 w-4 shrink-0 text-accent" />
                Trend and narrative tracking with velocity
              </li>
              <li className="flex gap-2">
                <Check className="mt-0.5 h-4 w-4 shrink-0 text-accent" />
                Influence networks and propagation paths
              </li>
            </ul>
          </div>
          <div className="border border-line bg-panel/60 p-5">
            <p className="label-caps mb-4">Processing pipeline</p>
            <ol className="space-y-4">
              {PIPELINE.map((step) => (
                <li key={step.index} className="flex items-start gap-3">
                  <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-md border border-line bg-panel2">
                    <step.icon className="h-3.5 w-3.5 text-accent" />
                  </span>
                  <div>
                    <p className="text-xs font-medium text-ink">
                      {step.label}
                    </p>
                    <p className="mt-0.5 text-[11px] leading-relaxed text-muted">
                      {step.detail}
                    </p>
                  </div>
                  <span className="ml-auto font-mono text-[10px] text-faint">
                    {step.index}
                  </span>
                </li>
              ))}
            </ol>
          </div>
        </div>
      </section>

      <section id="features" className="border-t border-line bg-panel/30 py-16">
        <div className="mx-auto max-w-6xl px-4">
          <p className="label-caps mb-3">Capabilities</p>
          <h2 className="text-2xl leading-snug text-ink">
            Four intelligence dimensions
          </h2>
          <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {FEATURES.map((feature) => (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, y: 12 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.35 }}
                className="border border-line bg-base/60 px-4 py-5"
              >
                <span className="flex h-8 w-8 items-center justify-center rounded-md bg-accent/10">
                  <feature.icon className="h-4 w-4 text-accent" />
                </span>
                <p className="mt-4 text-sm font-medium text-ink">
                  {feature.title}
                </p>
                <p className="mt-1.5 text-xs leading-relaxed text-muted">
                  {feature.detail}
                </p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      <section id="architecture" className="py-16">
        <div className="mx-auto max-w-6xl px-4">
          <p className="label-caps mb-3">Architecture preview</p>
          <h2 className="text-2xl leading-snug text-ink">
            From collection to visualization
          </h2>
          <div className="mt-8 grid grid-cols-1 gap-px overflow-hidden rounded-md border border-line bg-line md:grid-cols-4">
            {PIPELINE.map((step) => (
              <div key={step.label} className="bg-base px-5 py-6">
                <div className="flex items-center justify-between">
                  <step.icon className="h-4 w-4 text-accent" />
                  <span className="font-mono text-[10px] text-faint">
                    {step.index}
                  </span>
                </div>
                <p className="mt-4 text-sm text-ink">{step.label}</p>
                <p className="mt-1 text-[11px] leading-relaxed text-muted">
                  {step.detail}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section id="use-cases" className="border-t border-line bg-panel/30 py-16">
        <div className="mx-auto max-w-6xl px-4">
          <p className="label-caps mb-3">Use cases</p>
          <h2 className="text-2xl leading-snug text-ink">
            Built for teams that need to know first
          </h2>
          <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {USE_CASES.map((useCase) => (
              <div
                key={useCase.title}
                className="border border-line bg-base/60 px-4 py-5"
              >
                <useCase.icon className="h-4 w-4 text-accent" />
                <p className="mt-4 text-sm font-medium text-ink">
                  {useCase.title}
                </p>
                <p className="mt-1.5 text-xs leading-relaxed text-muted">
                  {useCase.detail}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="border-t border-line py-14">
        <div className="mx-auto flex max-w-6xl flex-col items-start justify-between gap-6 px-4 md:flex-row md:items-center">
          <div>
            <h2 className="text-xl text-ink">
              See it with a live demo workspace.
            </h2>
            <p className="mt-1.5 text-sm text-muted">
              Sign in with the demo analyst account and explore a fully
              populated environment.
            </p>
          </div>
          <Link
            to="/login"
            className="focus-ring inline-flex h-10 items-center gap-2 rounded-md bg-accent px-5 text-sm font-medium text-[#04211c] transition-colors hover:bg-accent-deep"
          >
            Enter demo workspace
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </section>

      <footer className="border-t border-line py-8">
        <div className="mx-auto flex max-w-6xl flex-col gap-2 px-4 font-mono text-[10px] uppercase tracking-[0.12em] text-faint md:flex-row md:items-center md:justify-between">
          <span>{APP_NAME} - AI social intelligence</span>
          <span>Smart India Hackathon 2026 - Problem Statement 26152</span>
        </div>
      </footer>
    </div>
  );
}
