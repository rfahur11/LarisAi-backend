from fastapi.testclient import TestClient
from main import app
import database

client = TestClient(app)

def test_forecasting_endpoint():
    database.client = None # Reset for event loop
    with TestClient(app) as client:
        response = client.get("/api/v1/ai/forecasting/stockouts")
        assert response.status_code == 200
        data = response.json()
        assert "predictions" in data

def test_clustering_endpoint():
    database.client = None # Reset for event loop
    with TestClient(app) as client:
        response = client.get("/api/v1/ai/clustering/customers")
        assert response.status_code == 200
        data = response.json()
        assert "clusters" in data

def test_promo_endpoint():
    payload = {
        "cluster_label": "Loyal",
        "message": "Diskon 50% untuk Anda!",
        "channel": "whatsapp"
    }
    with TestClient(app) as client:
        response = client.post("/api/v1/ai/promo/send", json=payload)
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
