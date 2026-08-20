---
sidebar_position: 1
title: Начало работы
---

# Начало работы

## Установка

Установка одной строкой:

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/install.sh | sh
```

Удаление:

```bash
curl -fsSL https://raw.githubusercontent.com/dopejs/gozen/main/uninstall.sh | sh
```

## Первый запуск

Добавьте первого провайдера:

```bash
zen config add provider
```

Запуск с профилем по умолчанию:

```bash
zen
```

Запуск с определённым профилем:

```bash
zen -p work
```

Запуск с определённой CLI:

```bash
zen --cli codex
```
