from analytics_service.main import app as analytics_app
from orders_service.main import app as orders_app


def _assert_api_key_security_contract(app, expected_path: str, expected_method: str) -> None:
    spec = app.openapi()

    security_schemes = spec.get("components", {}).get("securitySchemes", {})
    assert "APIKeyHeader" in security_schemes

    scheme = security_schemes["APIKeyHeader"]
    assert scheme.get("type") == "apiKey"
    assert scheme.get("name") == "X-API-Key"
    assert scheme.get("in") == "header"

    operation = spec["paths"][expected_path][expected_method]
    operation_security = operation.get("security", [])
    assert {"APIKeyHeader": []} in operation_security


def test_orders_openapi_exposes_api_key_security():
    _assert_api_key_security_contract(
        app=orders_app,
        expected_path="/orders",
        expected_method="get",
    )


def test_analytics_openapi_exposes_api_key_security():
    _assert_api_key_security_contract(
        app=analytics_app,
        expected_path="/analytics/summary",
        expected_method="get",
    )
