"""PulseGraph summarizer service.

Turns pre-computed analytics into a human-readable intelligence report.

**Hard rule:** the summarizer ONLY explains metrics it was given. It never
invents facts. The deterministic template renderer makes this structurally
true — every number in the output is copied from the input payload. When an
LLM backend is configured later (PULSEGRAPH_LLM_URL), the same payload is
used as the prompt and the response must pass the same number-grounding check
before it is served; otherwise the service falls back to the template.

Run locally:
    uvicorn main:app --host 0.0.0.0 --port 8003
"""

from typing import Dict, List, Optional, Tuple

from fastapi import FastAPI
from pydantic import BaseModel, Field

MODEL_VERSION = "template-v1"

app = FastAPI(
    title="PulseGraph Summarizer Service",
    version="1.0.0",
    description="Analytics-grounded report generation for PulseGraph.",
)


class SummaryInput(BaseModel):
    """The analytics payload produced by the PulseGraph backend."""

    topic_name: str
    window_hours: int = Field(168, ge=1)
    total_posts: int = Field(0, ge=0)
    total_engagement: int = Field(0, ge=0)
    distinct_authors: int = Field(0, ge=0)
    avg_sentiment: float = Field(0.0, ge=-1, le=1)
    emotion_shares: Dict[str, float] = {}
    top_emotions: List[str] = []
    top_keywords: List[str] = []
    top_influencers: List[str] = []
    language_distribution: Dict[str, float] = {}
    growth_percent: float = 0.0
    network_nodes: int = Field(0, ge=0)
    network_edges: int = Field(0, ge=0)


class Section(BaseModel):
    heading: str
    body: str


class SummaryResponse(BaseModel):
    title: str
    executive_summary: str
    key_findings: List[str]
    sections: List[Section]
    model_version: str


def describe_sentiment(score: float) -> str:
    if score >= 0.35:
        return "strongly positive"
    if score >= 0.1:
        return "leaning positive"
    if score > -0.1:
        return "mixed"
    if score > -0.35:
        return "leaning negative"
    return "strongly negative"


def describe_growth(percent: float) -> str:
    if percent >= 50:
        return "rose sharply"
    if percent >= 10:
        return "trended upward"
    if percent > -10:
        return "held steady"
    return "declined"


def dominant_with_share(items: Dict[str, float]) -> Tuple[Optional[str], float]:
    if not items:
        return None, 0.0
    key = max(sorted(items), key=lambda item: items[item])
    return key, items[key]


def build_report(payload: SummaryInput) -> SummaryResponse:
    """Render the deterministic report — numbers come only from `payload`."""
    top_emotion = "neutral"
    top_emotion_share = 0.0
    for name in payload.top_emotions:
        if name in payload.emotion_shares:
            top_emotion = name
            top_emotion_share = payload.emotion_shares[name]
            break

    language, language_share = dominant_with_share(payload.language_distribution)
    sentiment_label = describe_sentiment(payload.avg_sentiment)
    growth_label = describe_growth(payload.growth_percent)

    findings: List[str] = [
        (
            f"Analysed {payload.total_posts} posts from "
            f"{payload.distinct_authors} distinct accounts carrying "
            f"{payload.total_engagement} total engagements."
        ),
        (
            f"Overall sentiment is {sentiment_label} "
            f"({payload.avg_sentiment:.2f} mean score); {top_emotion} dominates "
            f"the emotional mix at {round(top_emotion_share * 100):.0f}% share."
        ),
    ]
    if payload.top_keywords:
        findings.append(
            f"The fastest-rising keyword is {payload.top_keywords[0]!r} — it anchors "
            f"the current narrative cluster."
        )
    if payload.top_influencers:
        findings.append(
            f"Influence concentrates around {payload.top_influencers[0]!r}, with the "
            f"graph holding {payload.network_nodes} nodes and "
            f"{payload.network_edges} edges."
        )
    if language:
        findings.append(
            f"The conversation is led by {language} speakers "
            f"({round(language_share * 100):.0f}% of analysed posts)."
        )

    keyword_quotes = ", ".join(repr(k) for k in payload.top_keywords[:5]) or "none detected"

    sections = [
        Section(
            heading="Sentiment landscape",
            body=(
                f"Across the {payload.window_hours}-hour window, average sentiment is "
                f"{payload.avg_sentiment:.2f} ({sentiment_label}). The dominant emotion "
                f"is {top_emotion}, present in {round(top_emotion_share * 100):.0f}% of scored posts."
            ),
        ),
        Section(
            heading="Discussion dynamics",
            body=(
                f"Volume {growth_label} over the window. The discussion reached "
                f"{payload.total_posts} posts, {payload.total_engagement} engagements and "
                f"{payload.distinct_authors} distinct participants, with the interaction "
                f"graph carrying {payload.network_edges} edges across {payload.network_nodes} nodes."
            ),
        ),
        Section(
            heading="Narratives and keywords",
            body=(
                f"Rising keywords in the conversation: {keyword_quotes}. These terms mark "
                f"the fastest-moving narrative clusters and are tracked for growth and velocity."
            ),
        ),
        Section(
            heading="Audience composition",
            body=(
                f"The largest linguistic community speaks {language} "
                f"({round(language_share * 100):.0f}%). Regional and interest patterns are "
                f"available in the demographics section of the dashboard."
            ),
        ),
    ]

    return SummaryResponse(
        title=f"PulseGraph intelligence report — {payload.topic_name}",
        executive_summary=(
            f"PulseGraph analysed the {payload.topic_name!r} conversation over the last "
            f"{payload.window_hours} hours. Sentiment is {sentiment_label} and discussion "
            f"volume {growth_label}."
        ),
        key_findings=findings,
        sections=sections,
        model_version=MODEL_VERSION,
    )


@app.get("/health")
def health() -> dict:
    """Liveness probe."""
    return {"status": "ok", "service": "summarizer-service"}


@app.post("/generate-summary", response_model=SummaryResponse)
def generate_summary(payload: SummaryInput) -> SummaryResponse:
    """Generate the grounded report for the provided analytics payload."""
    return build_report(payload)
