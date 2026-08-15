import os
import joblib
import pandas as pd
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException
from app.schemas import PredictionRequest, PredictionResponse
from train import train_and_save_model

MODEL_PATH = os.getenv("MODEL_PATH", "model.joblib")

# Multiplicadores e tarifas padrão
BASE_FARE = 2.50
RATE_PER_KM = 1.80

model = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    global model
    if not os.path.exists(MODEL_PATH):
        print(f"⚠️ Modelo {MODEL_PATH} não encontrado. Iniciando treinamento automático...")
        model = train_and_save_model(MODEL_PATH)
    else:
        print(f"📦 Carregando modelo treinado de {MODEL_PATH}...")
        model = joblib.load(MODEL_PATH)
    print("✅ Modelo ML pronto para requisições!")
    yield
    print("🧹 Encerrando microsserviço de ML...")

app = FastAPI(
    title="Routeforge ML Service",
    description="Microsserviço de Predição de ETA e Preço Dinâmico (Surge Pricing)",
    version="1.0.0",
    lifespan=lifespan
)

@app.get("/health")
def health_check():
    return {"status": "ok", "service": "routeforge-ml"}

@app.post("/predict", response_model=PredictionResponse)
def predict_pricing_and_eta(request: PredictionRequest):
    global model
    if model is None:
        raise HTTPException(status_code=500, detail="Modelo ML não inicializado.")
    
    # Prepara dataframe de entrada para o scikit-learn
    input_data = pd.DataFrame([{
        'distance_km': request.distance_km,
        'hour_of_day': request.hour_of_day,
        'day_of_week': request.day_of_week
    }])
    
    # Predição (retorna array [[eta_minutes, surge_multiplier]])
    prediction = model.predict(input_data)[0]
    
    predicted_eta = int(round(max(2, prediction[0])))
    predicted_surge = float(round(max(1.0, prediction[1]), 2))
    
    distance_fare = round(request.distance_km * RATE_PER_KM, 2)
    subtotal = BASE_FARE + distance_fare
    estimated_price = round(subtotal * predicted_surge, 2)
    
    return PredictionResponse(
        eta_minutes=predicted_eta,
        surge_multiplier=predicted_surge,
        base_fare=BASE_FARE,
        distance_fare=distance_fare,
        estimated_price=estimated_price
    )
