import os
import joblib
import grpc
import pandas as pd
from concurrent import futures

from app.proto import prediction_pb2
from app.proto import prediction_pb2_grpc
from train import train_and_save_model, WEATHER_MAP

MODEL_PATH = os.getenv("MODEL_PATH", "model.joblib")
BASE_FARE = 2.50
RATE_PER_KM = 1.80

model = None

def load_or_train_model():
    global model
    if not os.path.exists(MODEL_PATH):
        print(f"Modelo {MODEL_PATH} não encontrado. Iniciando treinamento automático...")
        model = train_and_save_model(MODEL_PATH)
    else:
        print(f"Carregando modelo treinado de {MODEL_PATH}...")
        model = joblib.load(MODEL_PATH)
    print("Modelo ML pronto para requisições gRPC/Protobuf!")

class PredictionServiceServicer(prediction_pb2_grpc.PredictionServiceServicer):

    def PredictPricingAndETA(self, request, context):
        global model
        if model is None:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details("Modelo ML não inicializado.")
            return prediction_pb2.PredictionResponse()

        traffic = request.traffic_level if request.traffic_level > 0 else 1.0
        weather_str = request.weather_condition.upper() if request.weather_condition else "CLEAR"
        weather_enc = WEATHER_MAP.get(weather_str, 1.0)

        # DataFrame de entrada com os recursos de trânsito e clima em tempo real
        input_data = pd.DataFrame([{
            'distance_km': request.distance_km,
            'hour_of_day': request.hour_of_day,
            'day_of_week': request.day_of_week,
            'traffic_level': traffic,
            'weather_encoded': weather_enc
        }])

        prediction = model.predict(input_data)[0]

        predicted_eta = int(round(max(2, prediction[0])))
        predicted_surge = float(round(max(1.0, prediction[1]), 2))

        distance_fare = round(request.distance_km * RATE_PER_KM, 2)
        subtotal = BASE_FARE + distance_fare
        estimated_price = round(subtotal * predicted_surge, 2)

        return prediction_pb2.PredictionResponse(
            eta_minutes=predicted_eta,
            surge_multiplier=predicted_surge,
            base_fare=BASE_FARE,
            distance_fare=distance_fare,
            estimated_price=estimated_price
        )

def serve():
    load_or_train_model()
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    prediction_pb2_grpc.add_PredictionServiceServicer_to_server(PredictionServiceServicer(), server)
    
    grpc_port = os.getenv("GRPC_PORT", "50051")
    server.add_insecure_port(f"[::]:{grpc_port}")
    server.start()
    print(f"Servidor gRPC do Routeforge ML iniciado na porta {grpc_port} (HTTP/2 Protobuf)")
    server.wait_for_termination()

if __name__ == "__main__":
    serve()
