"""PulseGraph sentiment service.

FastAPI microservice that scores social-media text: emotion, sentiment score
in [-1, 1] and a confidence estimate.

The default model is a transparent lexicon + rule engine that works fully
offline (no model downloads). Heavier backends can be plugged in later behind
the same /predict contract without touching the Go pipeline.

Run locally:
    uvicorn main:app --host 0.0.0.0 --port 8001
"""

from fastapi import FastAPI

from api.routes import router

app = FastAPI(
    title="PulseGraph Sentiment Service",
    version="1.0.0",
    description="Emotion + sentiment scoring for PulseGraph posts (English + Hinglish).",
)

app.include_router(router)


@app.get("/health")
def health() -> dict:
    """Liveness probe used by Docker/compose and Kubernetes."""
    return {"status": "ok", "service": "sentiment-service"}
