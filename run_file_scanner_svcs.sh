#!/bin/bash

# Проверяем, передан ли аргумент
if [ -z "$1" ]; then
    echo "Usage: $0 <number_of_processes>"
    exit 1
fi

NUM_PROCESSES=$1
PIDS=()  # Массив для хранения PID запущенных процессов

# Запускаем процессы в фоне и обновляем строку с их количеством
for ((i=0; i<NUM_PROCESSES; i++)); do
    go run cmd/file-scanner-service/main.go &  # Запуск в фоне
    PIDS+=($!)  # Сохраняем PID процесса

    echo -ne "\rStarted $((i+1)) processes..."
    # sleep 0.1   Короткая пауза, чтобы избежать мерцания вывода
done

echo ""  # Перенос строки после завершения запуска

# Ждем завершения всех процессов
for pid in "${PIDS[@]}"; do
    wait "$pid"
done

echo "All processes have finished."