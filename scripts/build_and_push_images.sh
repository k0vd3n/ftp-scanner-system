#!/bin/bash
set -e

# Проверяем, передан ли аргумент с тегом
if [ -z "$1" ]; then
    echo "Использование: $0 <tag>"
    exit 1
fi

TAG="$1"
DOCKERHUB_USER="k0vd3n"

# Список сервисов и соответствующих путей к Dockerfile относительно корня репозитория.
declare -A services
services["main-service"]="dockerfiles/main-service/Dockerfile" 
services["report-service"]="dockerfiles/generate-report-service/Dockerfile"
services["get-report-service"]="dockerfiles/get-report-service/Dockerfile"
services["status-service"]="dockerfiles/status-service/Dockerfile"
services["file-scanner-service"]="dockerfiles/file-scanner-service/Dockerfile"
services["directory-lister-service"]="dockerfiles/directory-lister-service/Dockerfile"
services["counter-reducer-service"]="dockerfiles/counter-reducer-service/Dockerfile"
services["scan-result-reducer-service"]="dockerfiles/scan-result-reducer-service/Dockerfile"

echo "Запуск параллельной сборки Docker-образов..."

# Запускаем сборку образов в параллельных процессах
for service in "${!services[@]}"; do
    dockerfile="${services[$service]}"
    full_image_name="${DOCKERHUB_USER}/${service}:${TAG}"
    echo "Сборка ${full_image_name} (Dockerfile: ${dockerfile})..."
    docker build -f "${dockerfile}" -t "${full_image_name}" . &
done

# Ждем завершения всех сборок
wait
echo "Сборка Docker-образов завершена."

echo "Запуск параллельной отправки Docker-образов в Docker Hub..."

# Отправляем образы в параллельных процессах
for service in "${!services[@]}"; do
    full_image_name="${DOCKERHUB_USER}/${service}:${TAG}"
    echo "Отправка ${full_image_name}..."
    docker push "${full_image_name}" &
done

# Ждем завершения отправки
wait
echo "Отправка Docker-образов завершена."
