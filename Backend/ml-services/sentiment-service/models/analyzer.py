"""Lexicon + rule based sentiment analyzer.

The analyzer is deliberately transparent and explainable: familiar word
lists, negation windows and intensity modifiers. It handles English and
romanised Hindi (Hinglish) and works fully offline, which keeps demos and CI
deterministic. A transformer backend can replace it later behind the same
`LexiconAnalyzer.predict` interface.
"""

from __future__ import annotations

import re
from dataclasses import dataclass
from typing import Dict

MODEL_VERSION = "lexicon-v1"

_TOKEN_RE = re.compile(r"[A-Za-z\u0900-\u097F]+")

NEGATORS = {
    "not", "no", "never", "none", "cannot", "cant", "can't", "dont", "don't",
    "isnt", "isn't", "wasnt", "wasn't", "wont", "won't", "without",
    "nahi", "nahin", "na", "mat",
}

INTENSIFIERS = {
    "very", "extremely", "really", "so", "totally", "absolutely", "insanely",
    "bahut", "बहुत", "बेहद", "hadd",
}

# word -> emotion category
POSITIVE: Dict[str, str] = {
    "good": "joy", "great": "joy", "awesome": "joy", "amazing": "joy",
    "fantastic": "joy", "wonderful": "joy", "love": "joy", "happy": "joy",
    "glad": "joy", "relief": "joy", "relieved": "joy", "success": "joy",
    "win": "joy", "wins": "joy", "winning": "joy", "brilliant": "joy",
    "excellent": "joy", "proud": "joy", "hope": "joy", "hopeful": "joy",
    "support": "joy", "helpful": "joy", "thanks": "joy", "thank": "joy",
    "best": "joy", "better": "joy", "safe": "joy", "rescued": "joy",
    "recovered": "joy", "calm": "joy", "brave": "joy", "courage": "joy",
    "generous": "joy", "kind": "joy", "peace": "joy", "smooth": "joy",
    "impressive": "joy", "stunning": "surprise", "unbelievable": "surprise",
    "surprising": "surprise",
    # Hinglish
    "accha": "joy", "achha": "joy", "badhiya": "joy", "shandar": "joy",
    "mubarak": "joy", "khushi": "joy", "jeet": "joy", "safal": "joy",
    "sahi": "joy", "madad": "joy", "zabardast": "joy",
}

NEGATIVE: Dict[str, str] = {
    # anger
    "angry": "anger", "outrage": "anger", "outrageous": "anger",
    "corrupt": "anger", "corruption": "anger", "scam": "anger",
    "shame": "anger", "shameless": "anger", "furious": "anger", "hate": "anger",
    "disgrace": "anger", "betrayal": "anger", "liar": "anger", "lies": "anger",
    "propaganda": "anger", "fraud": "anger", "cheating": "anger",
    "cheat": "anger", "stupid": "anger", "pathetic": "anger",
    "incompetent": "anger", "jhooth": "anger", "jhuth": "anger",
    "dhoka": "anger", "galat": "anger", "bakwas": "anger",
    # fear
    "fear": "fear", "feared": "fear", "panic": "fear", "scared": "fear",
    "danger": "fear", "dangerous": "fear", "threat": "fear",
    "warning": "fear", "worried": "fear", "worry": "fear", "anxiety": "fear",
    "crisis": "fear", "emergency": "fear", "terrified": "fear",
    "alarming": "fear", "urgent": "fear", "khatra": "fear", "dar": "fear",
    "darr": "fear", "pareshani": "fear", "mushkil": "fear", "tabahi": "fear",
    # sadness
    "sad": "sadness", "loss": "sadness", "lost": "sadness", "died": "sadness",
    "dead": "sadness", "death": "sadness", "victims": "sadness",
    "victim": "sadness", "tragic": "sadness", "tragedy": "sadness",
    "grief": "sadness", "missing": "sadness", "broken": "sadness",
    "sorrow": "sadness", "mourning": "sadness", "devastated": "sadness",
    "dukhi": "sadness", "nuksan": "sadness",
    # disgust
    "disgusting": "disgust", "gross": "disgust", "vile": "disgust",
    "nasty": "disgust", "repulsive": "disgust", "filth": "disgust",
    "filthy": "disgust", "shameful": "disgust", "revolting": "disgust",
    "kharab": "disgust", "bura": "disgust", "bekar": "disgust",
    # surprise (negative-leaning)
    "shocking": "surprise", "shocked": "surprise", "stunned": "surprise",
    "astonishing": "surprise",
}


@dataclass
class SentimentResult:
    """Structured verdict for one piece of text."""

    emotion: str
    sentiment_score: float
    confidence: float
    model_version: str = MODEL_VERSION

    def as_dict(self) -> dict:
        return {
            "emotion": self.emotion,
            "sentiment_score": self.sentiment_score,
            "confidence": self.confidence,
            "model_version": self.model_version,
        }


class LexiconAnalyzer:
    """Deterministic, explainable sentiment scorer."""

    def predict(self, text: str) -> dict:
        """Score text and return the API response payload."""
        return self._analyze(text).as_dict()

    def _analyze(self, text: str) -> SentimentResult:
        tokens = [token.lower() for token in _TOKEN_RE.findall(text)]
        if not tokens:
            return SentimentResult("neutral", 0.0, 0.5)

        positives = 0.0
        negatives = 0.0
        category_hits: Dict[str, float] = {}
        hits = 0

        for index, token in enumerate(tokens):
            weight = 1.5 if index > 0 and tokens[index - 1] in INTENSIFIERS else 1.0
            negated = any(
                tokens[j] in NEGATORS for j in range(max(0, index - 2), index)
            )

            if token in POSITIVE:
                hits += 1
                category = POSITIVE[token]
                category_hits[category] = category_hits.get(category, 0) + 1
                if negated:
                    negatives += weight
                else:
                    positives += weight
            elif token in NEGATIVE:
                hits += 1
                category = NEGATIVE[token]
                category_hits[category] = category_hits.get(category, 0) + 1
                if negated:
                    positives += weight * 0.5
                else:
                    negatives += weight

        raw = (positives - negatives) / 4.0
        score = round(max(-1.0, min(1.0, raw)), 4)

        if category_hits:
            emotion = max(
                category_hits.items(), key=lambda item: (item[1], item[0])
            )[0]
        elif score > 0.15:
            emotion = "joy"
        elif score < -0.15:
            emotion = "anger"
        else:
            emotion = "neutral"

        confidence = 0.5 if hits == 0 else max(0.55, min(0.92, 0.55 + 0.08 * hits))
        return SentimentResult(emotion, score, round(confidence, 4))
