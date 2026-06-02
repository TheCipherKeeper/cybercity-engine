# CyberCity — Engine

[![Part of CyberCity](https://img.shields.io/badge/CyberCity-composition-blueviolet)](https://github.com/TheCipherKeeper/cybercity)
[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/code-MIT-green)](LICENSE)
[![Docs: CC BY 4.0](https://img.shields.io/badge/docs-CC%20BY%204.0-lightgrey)](LICENSE-DOCS)

Go-каркас кибер-полигона **CyberCity**: событийное ядро, сетевая модель,
валидатор, движок симуляции, K8s-рендер манифестов. Это **оркестратор**
и «компилятор» города. Обложка проекта — [cybercity](https://github.com/TheCipherKeeper/cybercity).

> Канонический вход: [`master.md`](master.md) — философия и дорожная карта.
> Операционные правила для AI-агентов: [`AGENTS.md`](AGENTS.md).

## Статус

MVP-скелет. Все пакеты — заглушки с doc-комментариями. Реализация
идёт поэтапно по `master.md`. Текущая цель: `validate-network` поверх
`network.yml` (этап 2 дорожной карты).

## Что внутри

```
cmd/                           точки входа
  validate-network/            валидатор network.yml (заглушка)
  render-manifests/            рендер K8s-манифестов (заглушка)
internal/
  events/                      событийное ядро (ADR-0004)        — пакет-заглушка
  network/                     модель network.yml + валидатор     — пакет-заглушка
  sim/                         движок симуляции                   — пакет-заглушка
  runtime/                     K8s-адаптеры                       — пакет-заглушка
  scenario/                    пакеты сценариев                   — пакет-заглушка
network.yml                    канонический YAML (сеть + decoys)
network.md                     человекочитаемая проекция
docs/adr/                      архитектурные decision records
master.md                      философия и дорожная карта
AGENTS.md                      операционные правила для AI-агентов
```

## Быстрый старт

```bash
# сейчас (MVP)
go run ./cmd/validate-network
# → validate-network: not implemented yet

go run ./cmd/render-manifests
# → render-manifests: not implemented yet

# когда реализуем
go run ./cmd/validate-network            # прогнать network.yml
go run ./cmd/render-manifests            # сгенерить K8s-манифесты в network/generated/
kubectl --dry-run=client apply -f network/generated/
```

## Принципы

- **События — единственный источник истины.** Состояние = проекция потока.
- **Сеть декларативна.** Один YAML описывает всё. K8s — его проекция.
- **Безопасность по умолчанию.** Сегменты изолированы, каналы — явные.
- **LLM — помощник, не хозяин.** LLM пишет YAML, код валидирует, человек решает.
- **Воспроизводимость.** Один вход → один выход, детерминированный режим.

Подробности — в [`master.md`](master.md) и серии ADR в [`docs/adr/`](docs/adr/).

## Композиция CyberCity

| Слой | Репозиторий |
|---|---|
| Обложка / витрина | [cybercity](https://github.com/TheCipherKeeper/cybercity) |
| **Engine (этот репо)** | **cybercity-engine** |
| Данные | [cybercity-data](https://github.com/TheCipherKeeper/cybercity-data) |
| UI | [cybercity-ui](https://github.com/TheCipherKeeper/cybercity-ui) |
| Агенты | [cybercity-agents](https://github.com/TheCipherKeeper/cybercity-agents) |
| Blueprints | [cybercity-blueprints](https://github.com/TheCipherKeeper/cybercity-blueprints) |

## Лицензия

- Код: [MIT](LICENSE)
- Документация (`master.md`, ADR, комментарии): [CC BY 4.0](LICENSE-DOCS)
