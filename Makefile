# Запустить все тесты пять раз с детектором гонок.
test-race:
	go test -race -count=5 ./...

# Удалить локальные данные разработки для чистого запуска.
dev-local-clean:
	# Удалить локальную БД, загрузки и опубликованные файлы.
	rm -rf playground/data.dev-local/dev-local.db playground/data.dev-local/downloads/* playground/data.dev-local/public/*
	# Удалить все данные Docker-окружения разработки.
	rm -rf playground/data.dev-docker


# Загрузить переменные из .env.dev-local и запустить бота через Go.
dev-local-run:
	@set -ae; . ./.env.dev-local; set +a; go run ./cmd/voxify-bot

# Пересобрать образ разработки, удалив предыдущие контейнер и образ.
dev-docker-build:
	# Остановить предыдущий контейнер, если он существует.
	docker stop voxify-dev || true
	# Удалить контейнер, чтобы освободить имя и ссылку на образ.
	docker rm voxify-dev || true
	# Удалить предыдущий образ разработки, если он существует.
	docker rmi voxify-dev:latest || true
	# Собрать новый образ из текущего исходного кода.
	docker build -t voxify-dev:latest .

# Запустить бота с .env.dev-docker и данными в playground; удалить контейнер при выходе.
dev-docker-run:
	docker run \
		--rm \
		--name voxify-dev \
		--env-file .env.dev-docker \
		-v ./playground/data.dev-docker:/data \
		voxify-dev:latest

# Перегенерировать моки интерфейсов для тестов.
mockery:
	# Удалить старые моки, включая файлы для удалённых интерфейсов.
	rm -rf internal/mocks
	# Создать моки по конфигурации .mockery.yaml.
	mockery
