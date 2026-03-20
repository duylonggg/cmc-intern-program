# 🚀 Quick Reference - Docker Commands

## Essential Commands

### Start/Stop Services

```bash
# Start all services
docker-compose up -d
# or
make up

# Stop all services
docker-compose down
# or
make down

# Restart specific service
docker-compose restart backend
```

### View Status & Logs

```bash
# Check status
docker-compose ps
make status

# View all logs
docker-compose logs -f
make logs

# View specific service logs
docker-compose logs -f backend
make logs-backend
```

### Build & Rebuild

```bash
# Build images
docker-compose build

# Rebuild from scratch
docker-compose build --no-cache
make rebuild

# Build specific service
docker-compose build backend
```

## Access Points

| Service      | URL                          | Purpose       |
| ------------ | ---------------------------- | ------------- |
| Frontend     | http://localhost:3000        | Web UI        |
| Backend API  | http://localhost:8080        | REST API      |
| Health Check | http://localhost:8080/health | API Status    |
| pgAdmin      | http://localhost:5050        | DB Management |
| PostgreSQL   | localhost:5432               | Direct DB     |

## Common Tasks

### Test the API

```bash
# Health check
curl http://localhost:8080/health

# List assets
curl http://localhost:8080/assets

# Create asset
curl -X POST http://localhost:8080/assets \
  -H "Content-Type: application/json" \
  -d '{"domain":"example.com","type":"domain","criticality":"high"}'
```

### Access Containers

```bash
# Backend shell
docker-compose exec backend sh

# Frontend shell
docker-compose exec frontend sh

# Database shell
docker-compose exec db psql -U postgres -d mini_asm
```

### Database Operations

```bash
# Backup database
make backup
# or
docker-compose exec -T db pg_dump -U postgres mini_asm > backup.sql

# Restore database
make restore FILE=backup.sql
# or
docker-compose exec -T db psql -U postgres mini_asm < backup.sql

# Check database status
docker-compose exec db pg_isready -U postgres
```

### Troubleshooting

```bash
# Check which ports are in use
sudo lsof -i :8080
sudo lsof -i :3000

# Clean up everything
docker-compose down -v
docker system prune -a

# Rebuild everything
make rebuild
```

## Environment Variables

### Backend (.env or docker-compose.yml)

```env
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=mini_asm
```

### Frontend (nginx.conf)

API proxy is configured in `nginx.conf`:

```nginx
location /api {
    proxy_pass http://backend:8080;
}
```

## Quick Diagnostics

### Backend Not Starting

```bash
# Check logs
docker-compose logs backend

# Check database connection
docker-compose exec backend env | grep DB_

# Restart backend
docker-compose restart backend
```

### Frontend Not Loading

```bash
# Check logs
docker-compose logs frontend

# Check nginx config
docker-compose exec frontend cat /etc/nginx/conf.d/default.conf

# Test backend from frontend container
docker-compose exec frontend wget -O- http://backend:8080/health
```

### Database Issues

```bash
# Check database logs
docker-compose logs db

# Test connection
docker-compose exec db pg_isready -U postgres

# Check tables
docker-compose exec db psql -U postgres -d mini_asm -c "\dt"
```

## Docker Compose File Structure

```yaml
services:
  db: # PostgreSQL database
  backend: # Go API server
  frontend: # React + Nginx
  pgadmin: # Database GUI (optional)

volumes:
  pgdata: # Database persistence

networks:
  mini-asm-network: # Service communication
```

## Useful Docker Commands

```bash
# List running containers
docker ps

# List all containers
docker ps -a

# View container logs
docker logs mini-asm-backend

# Execute command in container
docker exec -it mini-asm-backend sh

# Remove all stopped containers
docker container prune

# Remove all unused images
docker image prune -a

# View disk usage
docker system df

# Remove everything (careful!)
docker system prune -a --volumes
```

## Production Checklist

- [ ] Change default passwords
- [ ] Use environment files for secrets
- [ ] Enable HTTPS/SSL
- [ ] Set up reverse proxy
- [ ] Configure resource limits
- [ ] Set up monitoring
- [ ] Configure automated backups
- [ ] Add rate limiting
- [ ] Enable authentication

## Get Help

```bash
# Make commands
make help

# Deploy script commands
./deploy.sh help

# Docker Compose help
docker-compose --help

# View specific command help
docker-compose up --help
```
