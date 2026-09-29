#!/usr/bin/env python3
"""PulseGraph demo-data seeder.

Generates 1000+ realistic social posts across three scenarios:

  1. Flood misinformation (Assam flood rumour + relief coordination)
  2. City Marathon 2026 (public event discussion)
  3. Nimbus X1 Launch (product launch reaction)

The data is deterministic (fixed RNG seed) so every demo environment gets
identical content: sentiment evolves over a 14-day window, authors interact
(replies + mentions) so the graph builder finds real structure, and language,
region and interest mixes exercise the demographics estimator.

Usage:
    python seed_db.py --emit-sql > demo_data.sql     # inspect / pipe to psql
    python seed_db.py --dsn postgres://...            # insert via psycopg2

Demo credentials (see README): all three users share the password
"PulseGraph@2026".
"""

from __future__ import annotations

import argparse
import random
import sys
from datetime import datetime, timedelta, timezone

try:
    import psycopg2  # type: ignore
except ImportError:  # pragma: no cover - only required for --dsn mode
    psycopg2 = None


DEMO_PASSWORD_HASH = "$2a$10$2xrT7RbijyYHv7cVu8NaeOhi/lCC/IT/8ahRmBrOVAl2DhR1xoZS."

USERS = [
    (1, "Asha Verma", "admin@pulsegraph.dev", "admin"),
    (2, "Rohan Iyer", "analyst@pulsegraph.dev", "analyst"),
    (3, "Meera Kapoor", "viewer@pulsegraph.dev", "viewer"),
]

SCENARIOS = [
    {
        "id": 1,
        "name": "Flood misinformation - Assam",
        "keywords": ["flood", "assam", "relief", "evacuation", "misinformation"],
        "description": "Tracks rumour spread and relief coordination during severe flooding in Assam.",
        "arc": "misinfo",
        "languages": [("en", 0.45), ("hi", 0.35), ("bn", 0.20)],
        "regions": [("Assam", 0.55), ("West Bengal", 0.20), ("Delhi", 0.15), ("Other", 0.10)],
        "posts": 380,
    },
    {
        "id": 2,
        "name": "City Marathon 2026",
        "keywords": ["marathon", "run", "traffic", "route", "event"],
        "description": "Monitors public discussion around the city marathon: route changes, registrations, race day.",
        "arc": "hype",
        "languages": [("en", 0.55), ("hi", 0.30), ("ta", 0.15)],
        "regions": [("Maharashtra", 0.45), ("Karnataka", 0.25), ("Delhi", 0.20), ("Other", 0.10)],
        "posts": 370,
    },
    {
        "id": 3,
        "name": "Nimbus X1 Launch",
        "keywords": ["nimbus", "launch", "phone", "battery", "price"],
        "description": "Product launch reaction: hype, first reviews and criticism for the Nimbus X1.",
        "arc": "launch",
        "languages": [("en", 0.60), ("hi", 0.25), ("te", 0.15)],
        "regions": [("Karnataka", 0.30), ("Telangana", 0.25), ("Delhi", 0.25), ("Other", 0.20)],
        "posts": 370,
    },
]

# (handle, display name, platform, tier)
AUTHORS = [
    ("floodwatch_assam", "FloodWatch Assam", "twitter", "broadcaster"),
    ("newsbharat", "News Bharat", "twitter", "broadcaster"),
    ("civic_watchdog", "Civic Watchdog", "twitter", "amplifier"),
    ("monsoon_alerts", "Monsoon Alerts", "twitter", "broadcaster"),
    ("citizen_ravi", "Ravi K", "twitter", "participant"),
    ("metro_maya", "Maya R", "twitter", "amplifier"),
    ("techwithravi", "Ravi Tech", "twitter", "amplifier"),
    ("trafficpulse_city", "TrafficPulse City", "twitter", "amplifier"),
    ("assam_relief_net", "Assam Relief Net", "telegram", "broadcaster"),
    ("monsoon_watch", "Monsoon Watch", "telegram", "amplifier"),
    ("city_events_hub", "City Events Hub", "telegram", "amplifier"),
    ("metro_updates", "Metro Updates", "telegram", "amplifier"),
    ("product_deals_in", "Deals India", "telegram", "amplifier"),
    ("techbharat_official", "TechBharat", "telegram", "broadcaster"),
    ("runner_ananya", "Ananya S", "twitter", "participant"),
    ("marathon_mumbai", "Mumbai Marathon Fan", "twitter", "amplifier"),
    ("pace_setter", "Pace Setter", "twitter", "participant"),
    ("gadget_guru", "Gadget Guru", "twitter", "broadcaster"),
    ("nimbus_fan_club", "Nimbus Fan Club", "twitter", "amplifier"),
    ("honest_reviewer", "Honest Reviewer", "twitter", "amplifier"),
    ("delhi_dad", "Delhi Dad", "twitter", "participant"),
    ("kolkata_kriti", "Kriti B", "twitter", "participant"),
    ("chennai_charts", "Chennai Charts", "twitter", "amplifier"),
    ("bengaluru_buzz", "Bengaluru Buzz", "twitter", "amplifier"),
    ("hyderabad_techie", "Hyd Techie", "twitter", "participant"),
    ("pune_runner", "Pune Runner", "twitter", "participant"),
    ("smartphone_sachin", "Sachin Reviews", "twitter", "amplifier"),
    ("daily_digest_in", "Daily Digest", "telegram", "broadcaster"),
    ("relief_volunteers", "Relief Volunteers", "telegram", "amplifier"),
    ("race_day_crew", "Race Day Crew", "telegram", "amplifier"),
    ("nimbus_updates", "Nimbus Updates", "telegram", "broadcaster"),
    ("city_general", "City General", "telegram", "participant"),
    ("mausam_mitra", "Mausam Mitra", "telegram", "amplifier"),
    ("run_with_riya", "Riya Runs", "twitter", "participant"),
    ("gadget_throws", "Gadget Throws", "twitter", "participant"),
    ("kirana_ketan", "Ketan Stores", "twitter", "participant"),
]

TEMPLATES = {
    "misinfo": {
        "very_neg": [
            "{kw} update: this is much worse than officials admit. Do not trust the official numbers.",
            "People are still trapped because of {kw}. Where is the administration?",
            "My cousin near the river says the {kw} situation is critical and rescue has not reached. We are scared.",
        ],
        "neg": [
            "Getting mixed reports about {kw}. Some say rescue is on the way, others say nothing has moved.",
            "The {kw} situation is tense. We need verified information, not forwarded rumours.",
            "Relief material for {kw} is delayed again. Families are waiting in the shelter.",
        ],
        "neu": [
            "Traffic being diverted near the bridge because of {kw}. Plan your commute accordingly.",
            "Helpline numbers for {kw} relief are pinned in the group. Please verify before forwarding.",
            "Weather office says {kw} conditions will ease over the next two days.",
        ],
        "pos": [
            "Rescue teams are finally reaching the worst-hit areas. Thank you to everyone helping with {kw} relief.",
            "Community kitchens for {kw} relief are running. Volunteers are doing great work.",
            "Supplies are moving again on the highway. {kw} relief effort is picking up pace.",
        ],
        "very_pos": [
            "Unbelievable response to the {kw} relief drive today. Proud of this community.",
            "The {kw} coordination between volunteers and officials has been excellent this week.",
        ],
    },
    "hype": {
        "very_neg": [
            "The {kw} route change is a disaster for local businesses. Zero notice, zero consultation.",
            "Three days of {kw} closures and nobody asked residents. Extremely poor planning.",
        ],
        "neg": [
            "Not sure the {kw} closures are worth it for one event. Traffic pain for everyone.",
            "Organisers of the {kw} should fix the water points before race day. Basic stuff.",
        ],
        "neu": [
            "Route map for the {kw} is out. Check which roads close on Sunday morning.",
            "Registration for the {kw} closes this week. Entry list details are in the link.",
            "The {kw} route passes the old town this year. Expect diversions from 5 AM.",
        ],
        "pos": [
            "Registered for the {kw}! Training plan is going well, see you at the start line.",
            "The {kw} expo was great today. Picked up my bib and the new shoes.",
            "Volunteering at the {kw} this year. The energy in the city is already building.",
        ],
        "very_pos": [
            "The {kw} energy is unreal. Whole neighbourhood is out cheering. Best morning of the year.",
            "Personal best at the {kw} today! The crowd support carried everyone home.",
        ],
    },
    "launch": {
        "very_neg": [
            "The {kw} pricing is a joke. Same specs as last year with a bigger number.",
            "Battery claims for the {kw} look inflated again. Wait for real reviews before buying.",
            "Disappointed by the {kw}. No {kw2} upgrade and they removed the charger from the box.",
        ],
        "neg": [
            "The {kw} looks fine but not worth the upgrade from last year's model.",
            "Feels like the {kw} is a small step. The {kw2} story is marketing more than engineering.",
        ],
        "neu": [
            "The {kw} goes on sale tomorrow. Spec sheet and price comparison are in the thread.",
            "Nimbus announced the {kw} today at their event. Full specs and {kw2} details here.",
        ],
        "pos": [
            "Got hands-on with the {kw}. Build quality is genuinely good and the {kw2} upgrade is noticeable.",
            "Pre-ordered the {kw}. The {kw2} improvements actually matter this time.",
            "The {kw} camera samples look great for this price range.",
        ],
        "very_pos": [
            "The {kw} sold out in 40 minutes. That {kw2} upgrade is worth every rupee.",
            "Two days with the {kw} and I am impressed. Battery and {kw2} are excellent.",
        ],
    },
}

EMOTIONS = {
    "misinfo": {
        "very_neg": [("fear", 5), ("anger", 3), ("sadness", 2)],
        "neg": [("fear", 4), ("anger", 3), ("sadness", 2)],
        "neu": [("neutral", 6), ("surprise", 2)],
        "pos": [("joy", 4), ("surprise", 2)],
        "very_pos": [("joy", 6), ("surprise", 1)],
    },
    "hype": {
        "very_neg": [("anger", 5), ("disgust", 3), ("sadness", 1)],
        "neg": [("anger", 3), ("sadness", 3), ("disgust", 2)],
        "neu": [("neutral", 6), ("surprise", 2)],
        "pos": [("joy", 5), ("surprise", 2)],
        "very_pos": [("joy", 6), ("surprise", 2)],
    },
    "launch": {
        "very_neg": [("disgust", 5), ("anger", 3), ("sadness", 1)],
        "neg": [("disgust", 3), ("sadness", 2), ("anger", 2)],
        "neu": [("neutral", 5), ("surprise", 3)],
        "pos": [("joy", 4), ("surprise", 3)],
        "very_pos": [("joy", 5), ("surprise", 3)],
    },
}


def esc(value: str) -> str:
    """Escape a string as a single-quoted SQL literal."""
    return "'" + str(value).replace("'", "''") + "'"


def _weighted(rng: random.Random, pairs):
    total = sum(weight for _, weight in pairs)
    pick = rng.uniform(0, total)
    upto = 0.0
    for value, weight in pairs:
        upto += weight
        if pick <= upto:
            return value
    return pairs[-1][0]


def _lerp(start: float, end: float, t: float) -> float:
    return start + (end - start) * t


def _arc_score(arc: str, progress: float) -> float:
    if arc == "misinfo":
        if progress < 0.65:
            return _lerp(-0.05, -0.65, progress / 0.65)
        return _lerp(-0.65, -0.05, (progress - 0.65) / 0.35)
    if arc == "hype":
        if progress < 0.5:
            return _lerp(-0.15, 0.65, progress / 0.5)
        return _lerp(0.65, 0.35, (progress - 0.5) / 0.5)
    # launch: hype spike, mixed reviews, settle
    if progress < 0.35:
        return _lerp(0.05, 0.7, progress / 0.35)
    if progress < 0.7:
        return _lerp(0.7, -0.25, (progress - 0.35) / 0.35)
    return _lerp(-0.25, 0.15, (progress - 0.7) / 0.3)


def _bucket(score: float) -> str:
    if score <= -0.55:
        return "very_neg"
    if score <= -0.18:
        return "neg"
    if score < 0.18:
        return "neu"
    if score < 0.55:
        return "pos"
    return "very_pos"


def _ref_for(handle: str, platform: str) -> str:
    return f"x:user:{handle}" if platform == "twitter" else f"tg:user:@{handle}"


def _generate_posts(rng: random.Random, scenario: dict, anchor: datetime, window_days: int):
    posts = []
    external_ids = []
    start = anchor - timedelta(days=window_days)
    count = scenario["posts"]

    for index in range(count):
        progress = index / max(1, count - 1)
        offset_days = window_days * progress + rng.uniform(-0.35, 0.35)
        offset_days = max(0.0, min(float(window_days), offset_days))
        timestamp = start + timedelta(days=offset_days)
        hour = int(min(23, max(0, rng.gauss(14, 4))))
        timestamp = timestamp.replace(
            hour=hour, minute=rng.randrange(0, 60), second=rng.randrange(0, 60)
        )

        handle, display, platform, tier = rng.choice(AUTHORS)

        score = max(-1.0, min(1.0, _arc_score(scenario["arc"], progress) + rng.uniform(-0.25, 0.25)))
        bucket = _bucket(score)

        keyword = rng.choice(scenario["keywords"])
        keyword2 = rng.choice(scenario["keywords"])
        template = rng.choice(TEMPLATES[scenario["arc"]][bucket])
        content = template.replace("{kw2}", keyword2).replace("{kw}", keyword)

        language = _weighted(rng, scenario["languages"])
        region = _weighted(rng, scenario["regions"])

        if tier == "broadcaster":
            engagement = int(rng.lognormvariate(6.4, 0.7))
        elif tier == "amplifier":
            engagement = int(rng.lognormvariate(4.6, 0.8))
        else:
            engagement = int(rng.lognormvariate(2.6, 1.0))
        if index % 47 == 23:
            engagement *= 9  # periodic viral spikes become "important events"
        engagement = max(0, engagement)

        reply_to = ""
        if index > 0 and rng.random() < 0.22:
            reply_to = rng.choice(external_ids)

        mentions = []
        if rng.random() < 0.35:
            for _ in range(1 + rng.randrange(0, 2)):
                other_handle, _, other_platform, _ = rng.choice(AUTHORS)
                if other_handle != handle:
                    mentions.append(_ref_for(other_handle, other_platform))
        mentions = list(dict.fromkeys(mentions))[:3]

        emotion = _weighted(rng, EMOTIONS[scenario["arc"]][bucket])
        confidence = round(rng.uniform(0.55, 0.92), 3)

        external_id = f"{platform}-seed-{scenario['id']}-{index + 1}"
        posts.append(
            {
                "platform": platform,
                "external_id": external_id,
                "author_ref": _ref_for(handle, platform),
                "author_name": display,
                "content": content,
                "language": language,
                "region": region,
                "timestamp": timestamp,
                "engagement": engagement,
                "reply_to": reply_to,
                "mentions": mentions,
                "emotion": emotion,
                "score": round(score, 3),
                "confidence": confidence,
            }
        )
        external_ids.append(external_id)

    return posts


def _post_row_sql(post_id: int, topic_id: int, post: dict) -> str:
    mentions = "ARRAY[]::text[]"
    if post["mentions"]:
        mentions = "ARRAY[" + ", ".join(esc(item) for item in post["mentions"]) + "]"
    return (
        f"({post_id}, {topic_id}, {esc(post['platform'])}, {esc(post['external_id'])}, "
        f"{esc(post['author_ref'])}, {esc(post['author_name'])}, {esc(post['content'])}, "
        f"{esc(post['language'])}, {esc(post['region'])}, {esc(post['timestamp'].isoformat())}, "
        f"{post['engagement']}, {esc(post['reply_to'])}, {mentions})"
    )


def _chunks(items, size):
    for start in range(0, len(items), size):
        yield items[start : start + size]


def _parse_anchor(raw):
    if not raw:
        return datetime.now(timezone.utc)
    parsed = datetime.fromisoformat(raw.replace("Z", "+00:00"))
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed


def build_statements(seed: int, anchor_raw):
    rng = random.Random(seed)
    anchor = _parse_anchor(anchor_raw)
    window_days = 14
    statements = [
        "-- PulseGraph demo data (deterministic; generated by scripts/seed_db.py)",
        f"-- seed={seed} anchor={anchor.isoformat()}",
        "BEGIN;",
    ]

    usernames = ", ".join(
        f"({uid}, {esc(name)}, {esc(email)}, {esc(DEMO_PASSWORD_HASH)}, {esc(role)})"
        for uid, name, email, role in USERS
    )
    statements.append(
        "INSERT INTO users (id, name, email, password_hash, role) VALUES "
        + usernames
        + " ON CONFLICT DO NOTHING;"
    )

    statements.append(
        "INSERT INTO social_sources (id, platform, source_name, configuration, status) VALUES "
        "(1, 'twitter', 'x-recent-search-demo', '{\"mode\": \"demo\"}'::jsonb, 'active'), "
        "(2, 'telegram', 'telegram-channels-demo', '{\"mode\": \"demo\"}'::jsonb, 'active') "
        "ON CONFLICT DO NOTHING;"
    )

    topic_values = ", ".join(
        f"({scenario['id']}, {esc(scenario['name'])}, "
        f"ARRAY[{', '.join(esc(word) for word in scenario['keywords'])}], "
        f"{esc(scenario['description'])})"
        for scenario in SCENARIOS
    )
    statements.append(
        "INSERT INTO topics (id, name, keywords, description) VALUES "
        + topic_values
        + " ON CONFLICT DO NOTHING;"
    )

    post_id = 1000
    sentiment_id = 2000
    total_posts = 0
    post_rows = []
    sentiment_rows = []

    for scenario in SCENARIOS:
        generated = _generate_posts(rng, scenario, anchor, window_days)
        for post in generated:
            post_id += 1
            sentiment_id += 1
            post_rows.append(_post_row_sql(post_id, scenario["id"], post))
            sentiment_rows.append(
                f"({sentiment_id}, {post_id}, {esc(post['emotion'])}, {post['score']}, "
                f"{post['confidence']}, {esc('seed-lexicon-v1')})"
            )
            total_posts += 1

    for chunk in _chunks(post_rows, 50):
        statements.append(
            "INSERT INTO social_posts (id, topic_id, platform, external_id, author_reference, "
            "author_display_name, content, language, region_hint, \"timestamp\", engagement_count, "
            "reply_to_external_id, mention_refs) VALUES\n"
            + ",\n".join(chunk)
            + "\nON CONFLICT DO NOTHING;"
        )

    for chunk in _chunks(sentiment_rows, 100):
        statements.append(
            "INSERT INTO sentiment_results (id, post_id, emotion, sentiment_score, confidence, "
            "model_version) VALUES\n"
            + ",\n".join(chunk)
            + "\nON CONFLICT DO NOTHING;"
        )

    for table in ("users", "social_sources", "topics", "social_posts", "sentiment_results"):
        statements.append(
            f"SELECT setval(pg_get_serial_sequence('{table}', 'id'), "
            f"GREATEST((SELECT COALESCE(MAX(id), 1) FROM {table}), 1));"
        )

    statements.append("COMMIT;")
    print(
        f"seed_db: generated {total_posts} posts with sentiment across "
        f"{len(SCENARIOS)} topics (seed={seed})",
        file=sys.stderr,
    )
    return statements


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate/insert PulseGraph demo data.")
    parser.add_argument("--dsn", help="PostgreSQL DSN for direct insert mode")
    parser.add_argument("--emit-sql", action="store_true", help="print SQL to stdout")
    parser.add_argument("--seed", type=int, default=42, help="RNG seed (default 42)")
    parser.add_argument(
        "--anchor",
        default=None,
        help="optional fixed anchor timestamp (ISO 8601) for fully deterministic reruns",
    )
    args = parser.parse_args()
    if not args.dsn and not args.emit_sql:
        parser.error("choose --dsn <url> or --emit-sql")

    statements = build_statements(args.seed, args.anchor)

    if args.emit_sql:
        for statement in statements:
            print(statement)
        return 0

    if psycopg2 is None:
        print(
            "psycopg2 is required for direct insert mode: "
            "pip install -r scripts/requirements.txt",
            file=sys.stderr,
        )
        return 2

    connection = psycopg2.connect(args.dsn)
    try:
        with connection, connection.cursor() as cursor:
            for statement in statements:
                cursor.execute(statement)
    finally:
        connection.close()
    print("seed_db: demo data inserted", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
