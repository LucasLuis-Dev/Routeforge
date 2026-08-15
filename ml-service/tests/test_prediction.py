import sys
import os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

import pytest
from fastapi.testclient import TestClient
from app.main import app

def test_health_check():
    with TestClient(app) as client:
        response = client.get("/health")
        assert response.status_code == 200
        assert response.json() == {"status": "ok", "service": "routeforge-ml"}

def test_predict_valid_request():
    with TestClient(app) as client:
        payload = {
            "distance_km": 10.0,
            "hour_of_day": 18,
            "day_of_week": 4
        }
        response = client.post("/predict", json=payload)
        assert response.status_code == 200
        
        data = response.json()
        assert "eta_minutes" in data
        assert "surge_multiplier" in data
        assert "estimated_price" in data
        
        assert data["eta_minutes"] > 0
        assert data["surge_multiplier"] >= 1.0
        assert data["estimated_price"] > 0.0
        
        # Formula check: price = (2.50 + 10.0 * 1.80) * surge_multiplier
        expected_distance_fare = round(10.0 * 1.80, 2)
        expected_total = round((2.50 + expected_distance_fare) * data["surge_multiplier"], 2)
        assert abs(data["estimated_price"] - expected_total) <= 0.01

def test_predict_invalid_distance():
    with TestClient(app) as client:
        payload = {
            "distance_km": -5.0,
            "hour_of_day": 12,
            "day_of_week": 1
        }
        response = client.post("/predict", json=payload)
        assert response.status_code == 422

def test_predict_invalid_hour():
    with TestClient(app) as client:
        payload = {
            "distance_km": 5.0,
            "hour_of_day": 25,
            "day_of_week": 1
        }
        response = client.post("/predict", json=payload)
        assert response.status_code == 422
