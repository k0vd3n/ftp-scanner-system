#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <number_of_processes>"
    exit 1
fi

NUM_PROCESSES=$1
BASE_PORT=2112
PIDS=()

for ((i=1; i<=NUM_PROCESSES; i++)); do
    INSTANCE="instance$i"
    PORT=$((BASE_PORT + i - 1))
    
    # Исправленные sed команды для точной замены
    sed -e "s/\(instance: \)\".*\"/\1\"$INSTANCE\"/" \
        -e "s/\(prom_http_port: \)\".*\"/\1\":$PORT\"/" \
        config/config.yaml > "config/config_$INSTANCE.yaml"


    echo "=== Config for $INSTANCE ==="
    grep -E '(instance|prom_http_port)' "config/config_$INSTANCE.yaml"
    echo "============================"

    go run cmd/directory-lister-service/main.go --config "config/config_$INSTANCE.yaml" &
    PIDS+=($!)
    echo -ne "\rStarted $i processes..."
done

echo ""

# Удаляем временные конфиги при завершении
trap 'kill ${PIDS[@]} 2>/dev/null; rm -f config/config_instance*.yaml' EXIT
wait "${PIDS[@]}"
echo "All processes have finished."