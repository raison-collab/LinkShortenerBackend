#!/bin/bash

if [ -f .env ]; then
  export $(cat .env | sed 's/#.*//g' | xargs)
fi

docker compose down

docker compose up -d --build

if [ ! -z "$DOCKER_NETWORK_PROXY" ]; then
  echo "Подключаем link_shortener_frontend к сети ${DOCKER_NETWORK_PROXY}..."
  docker network connect ${DOCKER_NETWORK_PROXY} link_shortener_app
else
  echo "Переменная DOCKER_NETWORK_PROXY не задана. Пропускаем подключение к сети."
fi

echo "Перезапуск завершен."