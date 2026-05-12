#!/usr/bin/env bash
set -euo pipefail

# Получаем команду из аргументов
COMMAND="$1"

# Выполняем поиск модулей и запуск переданной команды в каждом
while IFS= read -r dir; do
  echo "==> ${dir#$PWD/}"
  (
    cd "$dir"
    # Выполняем команду, переданную в скрипт
    eval "$COMMAND"
  )
done < <(go list -m -f '{{.Dir}}')

echo "==> Total time: $SECONDS sec."
