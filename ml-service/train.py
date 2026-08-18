import os
import joblib
import numpy as np
import pandas as pd
from sklearn.ensemble import RandomForestRegressor

WEATHER_MAP = {
    "CLEAR": 1.0,
    "RAIN": 1.3,
    "STORM": 1.6
}

def generate_synthetic_data(n_samples=3500, random_state=42):
    np.random.seed(random_state)
    
    # Features
    distance_km = np.random.uniform(0.5, 40.0, n_samples)
    hour_of_day = np.random.randint(0, 24, n_samples)
    day_of_week = np.random.randint(0, 7, n_samples) # 0=Monday, 6=Sunday
    
    # Real-time traffic variables (Waze/Google Maps API style)
    traffic_level = np.random.choice([1.0, 1.25, 1.5, 1.75, 2.0, 2.5], size=n_samples, p=[0.4, 0.2, 0.15, 0.1, 0.1, 0.05])
    
    # Weather conditions (CLEAR, RAIN, STORM)
    weather_labels = np.random.choice(["CLEAR", "RAIN", "STORM"], size=n_samples, p=[0.7, 0.2, 0.1])
    weather_encoded = np.array([WEATHER_MAP[w] for w in weather_labels])
    
    # Calculations with realistic traffic and weather adjustments
    is_weekday = day_of_week < 5
    is_morning_peak = is_weekday & (hour_of_day >= 7) & (hour_of_day <= 9)
    is_evening_peak = is_weekday & (hour_of_day >= 17) & (hour_of_day <= 19)
    is_night_weekend = (~is_weekday) & ((hour_of_day >= 21) | (hour_of_day <= 3))
    
    peak_factor = np.ones(n_samples)
    peak_factor[is_morning_peak] += np.random.uniform(0.3, 0.7, np.sum(is_morning_peak))
    peak_factor[is_evening_peak] += np.random.uniform(0.4, 0.9, np.sum(is_evening_peak))
    
    # Combined effective traffic & weather delay factor
    combined_factor = peak_factor * traffic_level * weather_encoded
    
    # Target 1: ETA in minutes (base speed 30 km/h)
    base_speed_kmh = 30.0
    raw_eta = (distance_km / base_speed_kmh * 60.0) * combined_factor
    noise_eta = np.random.normal(0, 1.0, n_samples)
    eta_minutes = np.clip(np.round(raw_eta + noise_eta), 2, 180)
    
    # Target 2: Surge Multiplier (1.0 to 3.5)
    surge_multiplier = np.ones(n_samples) * (traffic_level * 0.5 + weather_encoded * 0.5)
    surge_multiplier[is_morning_peak] += np.random.uniform(0.2, 0.5, np.sum(is_morning_peak))
    surge_multiplier[is_evening_peak] += np.random.uniform(0.3, 0.8, np.sum(is_evening_peak))
    surge_multiplier[is_night_weekend] += np.random.uniform(0.2, 0.5, np.sum(is_night_weekend))
    surge_multiplier = np.clip(np.round(surge_multiplier, 2), 1.0, 3.5)
    
    X = pd.DataFrame({
        'distance_km': distance_km,
        'hour_of_day': hour_of_day,
        'day_of_week': day_of_week,
        'traffic_level': traffic_level,
        'weather_encoded': weather_encoded
    })
    
    y = pd.DataFrame({
        'eta_minutes': eta_minutes,
        'surge_multiplier': surge_multiplier
    })
    
    return X, y

def train_and_save_model(model_path="model.joblib"):
    print("Gerando dataset sintético enriquecido com dados de trânsito e clima em tempo real...")
    X, y = generate_synthetic_data()
    
    print("Treinando modelo Random Forest Regressor de alta precisão...")
    model = RandomForestRegressor(n_estimators=120, random_state=42, n_jobs=-1)
    model.fit(X, y)
    
    print(f"Salvando modelo em {model_path}...")
    joblib.dump(model, model_path)
    print("Modelo ML treinado com sucesso!")
    return model

if __name__ == "__main__":
    train_and_save_model()
