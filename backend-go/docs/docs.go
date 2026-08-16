package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "Retorna o status de saúde do microsserviço Go",
                "produces": ["application/json"],
                "tags": ["Health"],
                "summary": "Healthcheck da API",
                "responses": {
                    "200": { "description": "OK" }
                }
            }
        },
        "/api/v1/users": {
            "post": {
                "description": "Cadastra um passageiro ou motorista no sistema",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Users"],
                "summary": "Cadastrar Usuário",
                "parameters": [
                    {
                        "description": "Dados do usuário",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "name": { "type": "string", "example": "Lucas Passageiro" },
                                "email": { "type": "string", "example": "lucas@example.com" },
                                "user_type": { "type": "string", "example": "passenger" }
                            }
                        }
                    }
                ],
                "responses": {
                    "201": { "description": "Created" },
                    "409": { "description": "Conflict" }
                }
            }
        },
        "/api/v1/rides/estimate": {
            "post": {
                "description": "Calcula a distância (Haversine), ETA e precificação dinâmica (ML ou Fallback em contingência)",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Rides"],
                "summary": "Calcular Estimativa de Corrida",
                "parameters": [
                    {
                        "description": "Coordenadas de Origem e Destino",
                        "name": "estimate_request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "origin_latitude": { "type": "number", "example": -23.550520 },
                                "origin_longitude": { "type": "number", "example": -46.633308 },
                                "destination_latitude": { "type": "number", "example": -23.561684 },
                                "destination_longitude": { "type": "number", "example": -46.655981 }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": { "description": "OK" }
                }
            }
        },
        "/api/v1/rides": {
            "post": {
                "description": "Solicita uma corrida e registra o preço estimado",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Rides"],
                "summary": "Criar Corrida",
                "parameters": [
                    {
                        "description": "Dados da corrida",
                        "name": "ride_request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "passenger_id": { "type": "string", "example": "841af28b-04a1-47bb-97c1-09ef4666a236" },
                                "origin_latitude": { "type": "number", "example": -23.550520 },
                                "origin_longitude": { "type": "number", "example": -46.633308 },
                                "destination_latitude": { "type": "number", "example": -23.561684 },
                                "destination_longitude": { "type": "number", "example": -46.655981 }
                            }
                        }
                    }
                ],
                "responses": {
                    "201": { "description": "Created" }
                }
            }
        },
        "/api/v1/rides/{id}/accept": {
            "post": {
                "description": "Atribui um motorista à corrida solicitada",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Rides"],
                "summary": "Aceitar Corrida",
                "parameters": [
                    { "type": "string", "description": "Ride UUID", "name": "id", "in": "path", "required": true },
                    {
                        "description": "ID do Motorista",
                        "name": "accept_request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "driver_id": { "type": "string", "example": "fd7477a3-c755-40ba-9496-2e0a9e5493d6" }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": { "description": "OK" }
                }
            }
        },
        "/api/v1/rides/{id}/complete": {
            "post": {
                "description": "Finaliza a corrida e consolida o valor final",
                "produces": ["application/json"],
                "tags": ["Rides"],
                "summary": "Finalizar Corrida",
                "parameters": [
                    { "type": "string", "description": "Ride UUID", "name": "id", "in": "path", "required": true }
                ],
                "responses": {
                    "200": { "description": "OK" }
                }
            }
        },
        "/api/v1/rides/{id}": {
            "get": {
                "description": "Retorna os detalhes da corrida e o histórico de auditoria de preços",
                "produces": ["application/json"],
                "tags": ["Rides"],
                "summary": "Consultar Detalhes da Corrida",
                "parameters": [
                    { "type": "string", "description": "Ride UUID", "name": "id", "in": "path", "required": true }
                ],
                "responses": {
                    "200": { "description": "OK" }
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Routeforge Mobility API",
	Description:      "API REST em Go para simulação de corridas de aplicativo com precificação dinâmica via Machine Learning e Fallback Pattern.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
