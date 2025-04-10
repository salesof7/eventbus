# EventBus - Sistema de Gerenciamento de Eventos

Este repositório contém uma implementação de um **EventBus** em Go, projetado para facilitar a publicação, o consumo e o processamento de eventos em sistemas distribuídos. O EventBus é uma ferramenta robusta que suporta integração com brokers externos, gerenciamento de filas, tratamento de erros e oferece suporte a tracing e métricas via OpenTelemetry para monitoramento detalhado.

Este documento explica cada parte do código e como a aplicação funciona, oferecendo uma visão clara para desenvolvedores que desejam utilizá-la ou contribuir com o projeto.

---

## Funcionalidades Principais

- **Publicação de eventos**: Permite enviar eventos de forma assíncrona para uma fila de requisições.
- **Consumo de eventos**: Processa eventos recebidos de uma fila de respostas com base em handlers registrados.
- **Registro de eventos**: Gerencia eventos e suas ações associadas (handlers).
- **Suporte a sagas**: Inclui eventos de compensação para lidar com falhas em fluxos complexos.
- **Telemetria**: Integração com OpenTelemetry para tracing e métricas, permitindo monitoramento em tempo real.
- **Otimização para alta carga**: Suporte a batching, worker pools e controle de backpressure para grandes volumes de eventos.

---

## Estrutura do Projeto

O projeto é organizado em componentes modulares, cada um com uma responsabilidade específica:

1. **EventBus**: Núcleo do sistema, responsável por gerenciar filas, publicar e processar eventos.
2. **EventRegistry**: Gerencia o registro e a busca de eventos no sistema.
3. **EventFlow**: Facilita a construção de fluxos de eventos e sagas.
4. **Event**: Define a estrutura de um evento, incluindo nome, handler e eventos relacionados.

---

## Como Funciona

### 1. EventBus

O `EventBus` é o componente central da aplicação. Ele coordena a publicação de eventos, o consumo de mensagens de um broker externo (se configurado) e o processamento dos eventos com base nos handlers registrados.

#### Componentes Principais

- **Filas**:
  - `requestQueue`: Armazena eventos a serem publicados.
  - `responseQueue`: Armazena eventos a serem processados.
  - `errorCallback`: Fila para erros encontrados durante o processamento.

- **Worker Pool**: Controla o número de goroutines simultâneas para evitar sobrecarga no sistema.
- **Batching**: Eventos são agrupados em lotes antes da publicação, otimizando o uso de recursos.
- **Telemetria**:
  - **Tracing**: Registra spans para monitoramento de publicação, processamento e erros.
  - **Métricas**: Monitora taxas de eventos, latências e tamanhos de filas.

#### Fluxo de Funcionamento

1. **Publicação**: Um evento é enviado ao `EventBus` via `Publish` e colocado na `requestQueue`.
2. **Processamento de Requisições**: O `EventBus` retira eventos da `requestQueue`, os agrupa em lotes e os publica usando `publishBatch`.
3. **Consumo**: Se um broker externo estiver configurado, eventos consumidos são colocados na `responseQueue`.
4. **Processamento de Respostas**: Eventos da `responseQueue` são processados pelo método `ProcessEvent`, que executa os handlers registrados em goroutines gerenciadas pelo worker pool.
5. **Tratamento de Erros**: Erros são enviados para a `errorCallback` e registrados na telemetria.

### 2. EventRegistry

O `EventRegistry` organiza e disponibiliza os eventos registrados no sistema. Ele utiliza um mapa onde a chave é o nome do evento e o valor é uma lista de eventos associados.

#### Funcionalidades

- **Register**: Adiciona uma lista de eventos ao registro.
- **Import**: Importa eventos de outro `EventRegistry`.
- **Get**: Recupera eventos registrados pelo nome.

### 3. EventFlow

O `EventFlow` ajuda a criar fluxos de eventos, incluindo sequências e sagas (eventos de compensação para falhas).

#### Funcionalidades

- **Next**: Adiciona um evento à sequência atual.
- **Saga**: Define um evento de compensação para o último evento da sequência.
- **Flat**: Retorna uma lista plana de todos os eventos no fluxo, incluindo sagas.

### 4. Event

A estrutura `Event` representa um evento individual no sistema. Ela contém:

- **Name**: Nome único do evento.
- **Saga**: Nome do evento de compensação (opcional).
- **Next**: Próximo evento na sequência (opcional).
- **Handler**: Função que processa o payload do evento.

---

## Telemetria

A telemetria é essencial para o monitoramento do `EventBus`, utilizando OpenTelemetry para tracing e métricas.

### Tracing

- **Spans**:
  - `PublishBatch`: Monitora a publicação de lotes de eventos.
  - `ProcessEvent`: Acompanha o processamento de um evento.
  - `EventHandler`: Registra a execução de um handler específico.
  - `errorCallback`: Rastreia o tratamento de erros.

- **Atributos**:
  - `event_name`: Nome do evento.
  - `batch_size`: Tamanho do lote publicado.
  - `error`: Detalhes de erros, se ocorrerem.

- **Eventos**:
  - "Starting event processing" e "Finished event processing" nos spans de `EventHandler`.

### Métricas

- **Contadores**:
  - `eventbus.publish.count`: Total de eventos publicados.
  - `eventbus.process.count`: Total de eventos processados.
  - `eventbus.errors`: Total de erros, categorizados por tipo.

- **Histogramas**:
  - `eventbus.publish.latency`: Latência de publicação.
  - `eventbus.process.latency`: Latência de processamento.

- **Gauges**:
  - `eventbus.queue.size`: Tamanho atual das filas (`request`, `response`, `error`).

---

## Configuração

O `EventBus` é configurado por meio da estrutura `EventBusConfig`, que define:

- **Tamanhos das filas**: `RequestQueueSize`, `ResponseQueueSize`, `ErrorQueueSize`.
- **Worker Pool**: `WorkerPoolSize` define o número máximo de goroutines simultâneas.
- **Tamanho do Batch**: `BatchSize` especifica quantos eventos são agrupados antes da publicação.
- **Timeout**: `Timeout` define o tempo máximo para operações.

---

## Como Usar

Siga os passos abaixo para integrar o `EventBus` ao seu projeto:

### 1. Crie um EventBus

```go
config := EventBusConfig{
    RequestQueueSize:  100,
    ResponseQueueSize: 100,
    ErrorQueueSize:    100,
    WorkerPoolSize:    10,
    BatchSize:         10,
    Timeout:           5 * time.Second,
}
eventBus, err := NewEventBus(broker, tracer, config)
if err != nil {
    log.Fatal(err)
}
```

### 2. Registre Eventos

```go
event := &Event{
    Name: "my_event",
    Handler: func(payload interface{}) (interface{}, error) {
        // Lógica do handler
        return nil, nil
    },
}
eventBus.Register([]*Event{event})
```

### 3. Inicie o evento

```go
eventBus.Start()
```

### 4. Publique Eventos

```go
err := eventBus.Publish("my_event", payload)
if err != nil {
    log.Println("Erro ao publicar:", err)
}
```

### 5. Pare o EventBus

```go
eventBus.Stop()
```

---

## Considerações Finais

O EventBus é uma solução escalável e monitorável para gerenciamento de eventos em sistemas distribuídos. Com suporte a alta carga, integração com telemetria e flexibilidade para fluxos complexos (como sagas), ele é ideal para aplicações que exigem processamento assíncrona robusto e confiável. Sinta-se à vontade para explorar o código, adaptá-lo às suas necessidades e contribuir com melhorias!

