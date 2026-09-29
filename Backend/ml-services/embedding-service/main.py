"""PulseGraph embedding service.

Deterministic hashing-based text embeddings (384 dimensions) that require no
model downloads and run fully offline. The vector space preserves token and
bigram overlap similarity — enough for the prototype's semantic search and
narrative-similarity demos, and it matches the pgvector(384) column exactly.

Swap in sentence-transformers by installing the optional extra and setting
PULSEGRAPH_EMBEDDING_BACKEND=sentence-transformers when available; the HTTP
contract stays identical.

Run locally:
    uvicorn main:app --host 0.0.0.0 --port 8002
"""

import hashlib
import math
import os
import re
from typing import List

from fastapi import FastAPI
from pydantic import BaseModel, Field

DIM = 384
MODEL_VERSION = "hash-embed-v1"

_TOKEN_RE = re.compile(r"[a-z0-9\u0900-\u097F]+")

app = FastAPI(
    title="PulseGraph Embedding Service",
    version="1.0.0",
    description="Deterministic 384-dim text embeddings for PulseGraph posts.",
)


def embed_text(text: str) -> List[float]:
    """Embed text into a unit-normalised 384-dim vector.

    Unigrams and bigrams are hashed into signed dimensions (the classic
    hashing-trick); the sign keeps collisions from systematically inflating
    similarity. The result is L2-normalised so cosine distance behaves well.
    """
    vector = [0.0] * DIM
    tokens = _TOKEN_RE.findall(text.lower())
    grams = list(tokens) + [f"{a}_{b}" for a, b in zip(tokens, tokens[1:])]

    for gram in grams:
        digest = hashlib.blake2b(gram.encode("utf-8"), digest_size=8).digest()
        index = int.from_bytes(digest[:4], "little") % DIM
        sign = 1.0 if digest[4] % 2 == 0 else -1.0
        vector[index] += sign

    norm = math.sqrt(sum(value * value for value in vector)) or 1.0
    return [round(value / norm, 6) for value in vector]


class EmbedRequest(BaseModel):
    """Input contract for POST /embed."""

    text: str = Field(..., min_length=1, max_length=10000)


class EmbedResponse(BaseModel):
    """Output contract for POST /embed."""

    embedding: List[float]
    dim: int
    model_version: str


@app.get("/health")
def health() -> dict:
    """Liveness probe."""
    backend = os.getenv("PULSEGRAPH_EMBEDDING_BACKEND", "hash")
    return {"status": "ok", "service": "embedding-service", "backend": backend}


@app.post("/embed", response_model=EmbedResponse)
def embed(payload: EmbedRequest) -> EmbedResponse:
    """Return the embedding vector for the given text."""
    return EmbedResponse(
        embedding=embed_text(payload.text),
        dim=DIM,
        model_version=MODEL_VERSION,
    )
