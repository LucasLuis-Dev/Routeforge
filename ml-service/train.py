import os
import joblib
import numpy as np
import pandas as pd
from sklearn.ensemble import RandomForestRegressor

def generate_synthetic_data(n_samples=2500, random_state=42):
    np.random.seed(random_state)
    
    # Features
    distance_km = np.random.uniform(0.5, 40.0, n_samples)
    hour_of_day = np.random.randint(0, 24, n_samples)
    day_of_week = np.random.randint(0, 7, n_samples) # 0=Monday, 6=Sunday
    
    # Base calculations with realistic noise
    traffic_factor = np.ones(n_samples)
    
    # Peak traffic hours (7-9 AM and 17-19 PM on weekdays)
    is_weekday = day_of_week < 5
    is_morning_peak = is_weekday & (hour_of_day >= 7) & (hour_of_day <= 9)
    is_evening_peak = is_weekday & (hour_of_day >= 17) & (hour_of_day <= 19)
    is_night_weekend = (~is_weekday) & ((hour_of_day >= 21) | (hour_of_day <= 3))
    
    traffic_factor[is_morning_peak] += np.random.uniform(0.4, 0.9, np.sum(is_morning_peak))
    traffic_factor[is_evening_peak] += np.random.uniform(0.5, 1.1, np.sum(is_evening_peak))
    
    # Target 1: ETA in minutes (avg speed 28 km/h adjusted by traffic factor)
    base_speed_kmh = 28.0
    raw_eta = (distance_km / base_speed_kmh * 60.0) * traffic_factor
    noise_eta = np.random.normal(0, 1.5, n_samples)
    eta_minutes = np.clip(np.round(raw_eta + noise_eta), 2, 180)
    
    # Target 2: Surge Multiplier (1.0 to 3.0)
    surge_multiplier = np.ones(n_samples)
    surge_multiplier[is_morning_peak] += np.random.uniform(0.2, 0.7, np.sum(is_morning_peak))
    surge_multiplier[is_evening_peak] += np.random.uniform(0.3, 1.0, np.sum(is_evening_peak))
    surge_multiplier[is_night_weekend] += np.random.uniform(0.2, 0.6, np.sum(is_night_weekend))
    surge_multiplier = np.clip(np.round(surge_multiplier, 2), 1.0, 3.0)
    
    X = pd.DataFrame({
        'distance_km': distance_km,
        'hour_of_day': hour_of_day,
        'day_of_week': day_of_week
    })
    
    y = pd.DataFrame({
        'eta_minutes': eta_minutes,
        'surge_multiplier': surge_multiplier
    })
    
    return X, y

def train_and_save_model(model_path="model.joblib"):
    print("Gerando dataset sintético de corridas...")
    X, y = generate_synthetic_data()
    
    print("Treinando modelo Random Forest Regressor...")
    model = RandomForestRegressor(n_estimators=100, random_state=42)
    model.fit(X, y)
    
    print(f"Salvando modelo em {model_path}...")
    joblib.dump(model, model_path)
    print("Modelo treinado com sucesso!")
    return model

if __name__ == "__main__":
    train_and_save_model()
