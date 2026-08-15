from pydantic import BaseModel, Field

class PredictionRequest(BaseModel):
    distance_km: float = Field(..., gt=0, description="Distância da corrida em quilômetros", examples=[12.5])
    hour_of_day: int = Field(..., ge=0, le=23, description="Hora do dia (0 a 23)", examples=[18])
    day_of_week: int = Field(..., ge=0, le=6, description="Dia da semana (0=Segunda, 6=Domingo)", examples=[4])

class PredictionResponse(BaseModel):
    eta_minutes: int = Field(..., description="Estimativa de tempo de chegada em minutos", examples=[25])
    surge_multiplier: float = Field(..., description="Multiplicador de preço dinâmico", examples=[1.35])
    base_fare: float = Field(default=2.50, description="Tarifa base da plataforma em BRL")
    distance_fare: float = Field(..., description="Tarifa referente à distância percorrida")
    estimated_price: float = Field(..., description="Valor total estimado final em BRL")
