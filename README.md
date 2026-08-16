# 🚖 Routeforge — Urban Mobility & Dynamic Pricing Simulation API

> Sistema de microsserviços para simulação de corridas de aplicativo em tempo real, com cálculo de rota/distância, estimativa preditiva de preço dinâmico (*Surge Pricing*) e ETA via Machine Learning, com arquitetura resiliente e *Fallback Pattern* em Go e Python.

---

## 📌 Visão Geral do Projeto

O **Routeforge** é um projeto de portfólio backend focado no domínio de mobilidade urbana e delivery (inspirado em plataformas como Uber e 99). O objetivo principal é demonstrar o desenvolvimento de uma arquitetura em Go preparada para ambientes de alta disponibilidade, integrando um microsserviço de Machine Learning e aplicando os princípios do **SOLID** e da **Clean Architecture**.

### 🌟 Destaques Arquiteturais
- **Serviço Principal em Go**: API REST com roteamento leve (`go-chi`), injeção de dependências por interface e separação em camadas (*Handler*, *Service*, *Repository*, *Client*).
- **Microsserviço de Predição (FastAPI / Python)**: Modelo `RandomForestRegressor` treinado para estimar o tempo de chegada (ETA) e a tarifa dinâmica com base em distância e horário.
- **Resiliência & Alta Disponibilidade (*Graceful Degradation*)**: Cliente HTTP em Go com `timeout` de 2s e mecanismo de **Fallback automático**. Se o microsserviço de ML falhar ou estiver indisponível, a API em Go assume uma precificação de contingência sem interromper a experiência do usuário.
- **Persistência Relacional**: Banco de dados PostgreSQL 16 com esquemas normalizados (`users`, `rides`, `price_history`) e migrações SQL versionadas.
- **Infraestrutura**: Orquestração completa via Docker e Docker Compose com *healthchecks*.

---

## 🏗️ Arquitetura do Sistema

```
+-----------------------------------------------------------------------------------+
|                                    CLIENTE                                        |
+------------------------------------------+----------------------------------------+
                                           | HTTP REST
                                           v
+-----------------------------------------------------------------------------------+
|                             SERVIÇO PRINCIPAL (Go API)                            |
|                                                                                   |
|  [Handlers/Router (Chi)]  -->  [Service Layer (SOLID Business Logic)]              |
|                                         |                  \                      |
|                                         v                   v                     |
|                             [Repository Layer]       [ML Client + Fallback]       |
+-------------------------------------+-----------------------+---------------------+
                                      |                       |
                             SQL      v                       | HTTP Sync (Timeout: 2s)
                       +--------------+---+                   v
                       | PostgreSQL DB    |       +-----------+---------------------+
                       | (Versioned Migr.)|       | MICROSSERVIÇO ML (FastAPI)       |
                       +------------------+       | (Model: Random Forest Regressor)|
                                                  +---------------------------------+
```

---

## 🛠️ Tecnologias Utilizadas

- **Go (Golang)**: `go-chi/chi/v5` (Router), `database/sql` + `lib/pq`, `stretchr/testify` (Testes e Mocks).
- **Python**: `FastAPI`, `Uvicorn`, `scikit-learn`, `pandas`, `pytest`.
- **Banco de Dados**: `PostgreSQL 16` com scripts de migração nativos.
- **DevOps**: `Docker`, `Docker Compose`.

---

## 📜 Requisitos do Sistema

Para entender em detalhes todos os Requisitos Funcionais (RF), Requisitos Não-Funcionais (RNF) e Regras de Negócio (RN), consulte o documento de engenharia de requisitos:

👉 **[Especificação Completa de Requisitos (docs/requirements.md)](docs/requirements.md)**

---

## 🗺️ Roadmap de Desenvolvimento

- [x] **Etapa 1: Estrutura do Projeto & Infraestrutura Inicial**
  - Especificação detalhada de Requisitos (`docs/requirements.md`).
  - Modelagem do banco de dados relacional e migrações SQL (`db/migrations/`).
  - Orquestração inicial via Docker Compose com PostgreSQL.
- [x] **Etapa 2: Microsserviço de Predição de Tarifa & ETA (FastAPI)**
  - Script de geração de dados sintéticos (2.500 amostras) e treino do modelo `RandomForestRegressor`.
  - Endpoints REST `/health` e `/predict` validados com Pydantic V2.
  - Testes unitários com `pytest` (100% aprovados) e `Dockerfile` multi-stage executável.
- [x] **Etapa 3: Serviço Go — Camada de Domínio & Repositórios (SOLID)**
  - Entidades de domínio (`User`, `Ride`, `PriceHistory`) e abstrações por interface.
  - Implementação concreta dos repositórios PostgreSQL (`userRepository` e `rideRepository`) com `database/sql` e pool de conexões.
- [x] **Etapa 4: Serviço Go — Regras de Negócio & Resiliência (Fallback)**
  - Cálculo de distância geodésica via Fórmula de Haversine (`pkg/geo/haversine.go`).
  - Cliente HTTP síncrono com timeout de 2s (`client/ml_client.go`).
  - Lógica de negócios `RideService` com acionamento do **Fallback Pattern** em indisponibilidades de ML.
  - Cobertura por testes unitários com mocks (`stretchr/testify`) 100% aprovados.
- [ ] **Etapa 5: Serviço Go — Handlers HTTP & Rotas (`go-chi`)**
  - Implementação dos endpoints REST para cadastro de usuários e fluxo completo da corrida.
- [ ] **Etapa 6: Conteinerização Multi-stage & Integração End-to-End**
  - `Dockerfile` otimizado para Go.
  - Testes de resiliência em tempo real simulando queda do microsserviço de ML.
- [ ] **Etapa 7: Documentação Final & Guia de Produção**
  - Justificativas arquiteturais detalhadas e roadmap de evolução do modelo preditivo em produção.

---

## 🚀 Como Rodar Localmente (Etapa Atual)

### Pré-requisitos
- Docker & Docker Compose instalados.

### 1. Clonar o Repositório
```bash
git clone https://github.com/seu-usuario/routeforge.git
cd routeforge
```

### 2. Iniciar o Banco de Dados PostgreSQL via Docker Compose
```bash
docker compose up -d postgres
```

### 3. Verificar o Status das Tabelas no Postgres
```bash
docker exec -it routeforge-postgres psql -U routeforge_user -d routeforge_db -c "\dt"
```

---

## 📄 Licença

Este projeto está sob a licença MIT. Sinta-se à vontade para utilizar e estudar a implementação.
