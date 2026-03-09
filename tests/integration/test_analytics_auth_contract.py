from fastapi.testclient import TestClient

import analytics_service.rate_limiter as rate_limiter
from analytics_service.core.config import settings as analytics_settings
from analytics_service.main import app as analytics_app


def _different_key(valid_key: str) -> str:
    candidate = "invalid-api-key"
    if candidate == valid_key:
        candidate = f"{valid_key}-wrong"
    return candidate


def test_analytics_summary_requires_api_key_header():
    rate_limiter._rate_limit_store.clear()
    with TestClient(analytics_app) as client:
        response = client.get("/analytics/summary")

    assert response.status_code == 401
    assert response.json() == {"detail": "Invalid or missing API key"}


def test_analytics_summary_rejects_incorrect_api_key():
    rate_limiter._rate_limit_store.clear()
    wrong_key = _different_key(analytics_settings.ORDERS_API_KEY)

    with TestClient(analytics_app) as client:
        response = client.get(
            "/analytics/summary",
            headers={"X-API-Key": wrong_key},
        )

    assert response.status_code == 401
    assert response.json() == {"detail": "Invalid or missing API key"}
