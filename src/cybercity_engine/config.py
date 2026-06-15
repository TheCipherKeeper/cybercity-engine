"""Runtime configuration for the engine."""

from __future__ import annotations

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict

__all__ = ["EngineConfig"]


class EngineConfig(BaseSettings):
    """Engine settings loaded from environment variables and/or .env."""

    model_config = SettingsConfigDict(
        env_prefix="CYBERCITY_",
        extra="ignore",
    )

    app_name: str = "cybercity-engine"
    debug: bool = Field(default=False)

    # Engine runtime
    tick_ms: int = Field(default=1000, ge=0)
    engine_zip_url: str = Field(default="http://localhost:9000/cybercity/engine.zip")
    engine_zip_path: str | None = Field(default=None)

    # Redpanda / Kafka
    kafka_bootstrap_servers: str = Field(default="localhost:9092")
    kafka_group_id: str = Field(default="cybercity-engine")
    kafka_topics: list[str] = Field(default=["city.commands", "city.events"])

    # PostgreSQL
    database_url: str = Field(
        default="postgresql://engine:engine@localhost:5432/cybercity"
    )
    snapshot_interval_ticks: int = Field(default=10, ge=1)

    # Web/API
    host: str = Field(default="0.0.0.0")
    port: int = Field(default=8000, ge=1, le=65535)

    # MinIO / S3
    s3_endpoint: str | None = Field(default=None)
    s3_bucket: str = Field(default="cybercity")
    s3_access_key: str | None = Field(default=None)
    s3_secret_key: str | None = Field(default=None)
