# Especificação de Requisitos: Routeforge

Este documento especifica os Requisitos Funcionais (RF), Requisitos Não-Funcionais (RNF) e Regras de Negócio (RN) do sistema de simulação de corridas urbanas **Routeforge**.

---

## 🎯 1. Requisitos Funcionais (RF)

| ID | Nome | Descrição | Atores |
|---|---|---|---|
| **RF01** | Cadastrar Usuário | O sistema deve permitir o cadastro de passageiros e motoristas especificando nome, email e tipo de usuário. | Passageiro, Motorista |
| **RF02** | Solicitar Estimativa de Corrida | O sistema deve receber as coordenadas de origem e destino, calcular a distância em km (Haversine), solicitar a estimativa de tempo (ETA) e o multiplicador de preço ao microsserviço de ML (ou acionar fallback) e retornar o orçamento ao passageiro. | Passageiro |
| **RF03** | Criar/Confirmar Corrida | O sistema deve registrar a solicitação formal de corrida com status inicial `requested` e gravar o histórico inicial de preços. | Passageiro |
| **RF04** | Aceitar Corrida | O sistema deve permitir que um motorista aceite uma corrida que esteja com status `requested`, alterando o status para `accepted` e vinculando o motorista. | Motorista |
| **RF05** | Finalizar Corrida | O sistema deve permitir que a corrida transicione para `in_progress` e posteriormente para `completed`, gravando o valor final cobrado. | Motorista |
| **RF06** | Consultar Detalhes da Corrida | O sistema deve permitir consultar o estado atual de uma corrida, incluindo dados de localização, motorista, valores e se o preço utilizou modelo preditivo ou cálculo de contingência (*fallback*). | Passageiro, Motorista |

---

## ⚡ 2. Requisitos Não-Funcionais (RNF)

| ID | Categoria | Descrição |
|---|---|---|
| **RNF01** | **Resiliência / Alta Disponibilidade** | Se a chamada HTTP síncrona para o microsserviço de ML falhar (timeout > 2000ms, erro 5xx ou conexão recusada), a API em Go deve acionar automaticamente a estratégia de *Fallback* offline sem falhar a requisição do usuário (*Graceful Degradation*). |
| **RNF02** | **Arquitetura & SOLID** | O serviço principal em Go deve ser estruturado em camadas bem definidas (Handler, Service, Repository, Client), utilizando interfaces para injeção de dependência e desacoplamento total de componentes. |
| **RNF03** | **Desempenho** | O tempo de resposta P95 da API em Go para estimativas de corrida em fluxo normal deve ser inferior a 100ms (excluindo a latência da chamada externa de ML). |
| **RNF04** | **Testabilidade** | As regras de negócio críticas (cálculo de distância Haversine, transição de estados de corrida e motor de fallback) devem possuir cobertura de testes unitários com mocks. |
| **RNF05** | **Conteinerização & Portabilidade** | Toda a solução (PostgreSQL, Go API e Python ML Service) deve ser executável de forma isolada através de um único comando `docker compose up --build`. |
| **RNF06** | **Persistência & Integridade** | O banco de dados PostgreSQL deve utilizar migrações versionadas, chaves primárias baseadas em UUID e índices estratégicos para consultas frequentes por status e usuários. |

---

## 📜 3. Regras de Negócio (RN)

| ID | Nome | Regra |
|---|---|---|
| **RN01** | Cálculo Geodésico | A distância percorrida é estimada em linha reta corrigida através da **Fórmula de Haversine** considerando a curvatura média da Terra ($R = 6371\text{ km}$). |
| **RN02** | Ciclo de Vida da Corrida | O status da corrida deve obrigatoriamente seguir a sequência de transição: `requested` $\rightarrow$ `accepted` $\rightarrow$ `in_progress` $\rightarrow$ `completed`. Uma corrida só pode ser alterada para `canceled` se estiver no status `requested` ou `accepted`. |
| **RN03** | Precificação Normal (ML) | O valor total estimado da corrida é calculado pela fórmula: $\text{Preço} = \text{Tarifa Base (\$2.50)} + (\text{Distância em km} \times \text{\$1.80} \times \text{Multiplicador Dinâmico ML})$. |
| **RN04** | Precificação em Contingência (*Fallback*) | Na indisponibilidade do serviço de ML, o multiplicador dinâmico é fixado em $1.00$ (sem taxa de pico), o ETA é estimado considerando velocidade média de $30\text{ km/h}$, e o evento é registrado na tabela `price_history` com a flag `is_fallback = TRUE`. |
| **RN05** | Exclusividade de Motorista | Um motorista não pode aceitar uma nova corrida se já possuir uma corrida ativa nos status `accepted` ou `in_progress`. |
