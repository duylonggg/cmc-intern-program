.PHONY: help build up down restart status logs logs-backend logs-frontend logs-db clean rebuild test exec-backend exec-frontend exec-db backup restore

# Default target
help:
	@echo "Mini ASM Docker Deployment Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build all Docker images"
	@echo "  up             Start all services"
	@echo "  down           Stop all services"
	@echo "  restart        Restart all services"
	@echo "  status         Show service status"
	@echo "  logs           Show logs for all services"
	@echo "  logs-backend   Show backend logs"
	@echo "  logs-frontend  Show frontend logs"
	@echo "  logs-db        Show database logs"
	@echo "  clean          Remove all containers and volumes"
	@echo "  rebuild        Rebuild all images from scratch"
	@echo "  test           Test API endpoints"
	@echo "  exec-backend   Open shell in backend container"
	@echo "  exec-frontend  Open shell in frontend container"
	@echo "  exec-db        Open PostgreSQL shell"
	@echo "  backup         Backup database"
	@echo ""
	@echo "Examples:"
	@echo "  make up         # Start all services"
	@echo "  make logs       # View all logs"
	@echo "  make rebuild    # Rebuild everything"

# Build all images
build:
	docker compose build

# Start all services
up:
	docker compose up -d
	@echo "Waiting for services to start..."
	@sleep 5
	@$(MAKE) status

# Stop all services
down:
	docker compose down

# Restart all services
restart:
	docker compose restart
	@sleep 3
	@$(MAKE) status

# Show service status
status:
	@echo "=== Service Status ==="
	@docker compose ps
	@echo ""
	@echo "=== Health Checks ==="
	@curl -s http://localhost:8080/health 2>/dev/null && echo "✅ Backend: http://localhost:8080" || echo "❌ Backend not responding"
	@curl -s http://localhost:3000 2>/dev/null > /dev/null && echo "✅ Frontend: http://localhost:3000" || echo "❌ Frontend not responding"
	@docker compose exec -T db pg_isready -U postgres 2>/dev/null && echo "✅ Database: localhost:5432" || echo "❌ Database not responding"

# Show logs for all services
logs:
	docker compose logs -f

# Show backend logs
logs-backend:
	docker compose logs -f backend

# Show frontend logs
logs-frontend:
	docker compose logs -f frontend

# Show database logs
logs-db:
	docker compose logs -f db

# Clean everything
clean:
	docker compose down -v

# Rebuild everything
rebuild:
	docker compose build --no-cache
	docker compose up -d
	@sleep 5
	@$(MAKE) status

# Test API
test:
	@echo "=== Testing Health Endpoint ==="
	@curl -s http://localhost:8080/health | jq '.' || echo "Failed"
	@echo ""
	@echo "=== Testing List Assets ==="
	@curl -s http://localhost:8080/assets | jq '.' || echo "Failed"
	@echo ""
	@echo "=== Testing Frontend ==="
	@curl -s http://localhost:3000 > /dev/null && echo "✅ Frontend accessible" || echo "❌ Frontend not accessible"

# Execute shell in backend
exec-backend:
	docker compose exec backend sh

# Execute shell in frontend
exec-frontend:
	docker compose exec frontend sh

# Execute PostgreSQL shell
exec-db:
	docker compose exec db psql -U postgres -d mini_asm

# Backup database
backup:
	@mkdir -p backups
	docker compose exec -T db pg_dump -U postgres mini_asm > backups/backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup created in backups/"

# Restore database (usage: make restore FILE=backups/backup.sql)
restore:
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Please specify FILE=path/to/backup.sql"; \
		exit 1; \
	fi
	docker compose exec -T db psql -U postgres mini_asm < $(FILE)
	@echo "✅ Database restored from $(FILE)"
