.PHONY: help install dev test lint format type-check run validate

help:
	@echo "Available commands:"
	@echo "  install    Install dependencies"
	@echo "  dev        Run development server with hot reload"
	@echo "  test       Run tests"
	@echo "  lint       Run linter"
	@echo "  format     Format code"
	@echo "  run        Run production server"
	@echo "  validate   Run all checks"

install:
	poetry install

dev:
	poetry run python -m src.main

test:
	poetry run pytest

lint:
	poetry run ruff check src tests

format:
	poetry run black src tests
	poetry run isort src tests

type-check:
	poetry run mypy src

run:
	poetry run uvicorn src.main:app --host 0.0.0.0 --port 8000

validate: format lint test