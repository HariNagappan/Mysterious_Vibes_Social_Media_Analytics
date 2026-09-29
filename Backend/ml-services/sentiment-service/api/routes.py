"""HTTP routes of the sentiment service."""

from fastapi import APIRouter
from pydantic import BaseModel, Field

from models.analyzer import LexiconAnalyzer

router = APIRouter()

_analyzer = LexiconAnalyzer()


class PredictRequest(BaseModel):
    """Input contract for POST /predict."""

    text: str = Field(..., min_length=1, max_length=5000, description="Text to score")


class PredictResponse(BaseModel):
    """Output contract for POST /predict."""

    emotion: str
    sentiment_score: float = Field(..., ge=-1, le=1)
    confidence: float = Field(..., ge=0, le=1)
    model_version: str


@router.post("/predict", response_model=PredictResponse)
def predict(payload: PredictRequest) -> PredictResponse:
    """Score one piece of text.

    Returns the dominant emotion, a sentiment score in [-1, 1] (negative →
    positive) and a confidence estimate.
    """
    return PredictResponse(**_analyzer.predict(payload.text))
