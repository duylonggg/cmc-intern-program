# Docker Deployment Guide

This directory contains the complete Docker setup for the Mini ASM application, including database, backend API, and frontend.

## 📁 Structure

```
session7-deployment/
├── docker-compose.yml          # Orchestrates all services
├── backend/
│   ├── Dockerfile             # Go backend container
│   └── .dockerignore         # Exclude files from build
└── frontend/
    ├── Dockerfile             # React frontend container
    ├── nginx.conf            # Nginx configuration
    └── .dockerignore         # Exclude files from build
```

## 🚀 Quick Start

### Prerequisites

- Docker (v20.10+)
- Docker Compose (v2.0+)

### Start All Services

```bash
# From the session7-deployment directory
docker-compose up -d
```

This will start:

- **PostgreSQL** database on port `5432`
- **Backend API** on port `8080`
- **Frontend** on port `3000`
- **pgAdmin** (optional) on port `5050`

### Check Service Status

```bash
docker-compose ps
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f db
```

### Stop All Services

```bash
docker-compose down
```

### Stop and Remove Volumes (Clean Reset)

```bash
docker-compose down -v
```

## 🌐 Access Points

| Service      | URL                          | Description             |
| ------------ | ---------------------------- | ----------------------- |
| Frontend     | http://localhost:3000        | React application       |
| Backend API  | http://localhost:8080        | Go REST API             |
| Health Check | http://localhost:8080/health | API health status       |
| pgAdmin      | http://localhost:5050        | Database GUI (optional) |
| PostgreSQL   | localhost:5432               | Direct database access  |

### pgAdmin Credentials

- **Email:** admin@miniasm.com
- **Password:** admin

## 🏗️ Build Details

### Backend (Go)

- **Base Image:** golang:1.24-alpine (builder), alpine:latest (runtime)
- **Build Process:** Multi-stage build for minimal image size
- **Port:** 8080
- **Environment Variables:**
  - `DB_HOST`: Database host (default: db)
  - `DB_PORT`: Database port (default: 5432)
  - `DB_USER`: Database user (default: postgres)
  - `DB_PASSWORD`: Database password (default: postgres)
  - `DB_NAME`: Database name (default: mini_asm)

### Frontend (React + Vite)

- **Base Image:** node:18-alpine (builder), nginx:alpine (runtime)
- **Build Process:** Vite build + Nginx static serving
- **Port:** 80 (mapped to 3000 on host)
- **Features:**
  - API proxy to backend at `/api/*`
  - SPA routing support
  - Gzip compression
  - Static asset caching

### Database (PostgreSQL)

- **Image:** postgres:15-alpine
- **Port:** 5432
- **Default Credentials:**
  - User: postgres
  - Password: postgres
  - Database: mini_asm
- **Migrations:** Auto-applied on first startup

## 🔧 Development Commands

### Rebuild Specific Service

```bash
# Rebuild backend
docker-compose build backend

# Rebuild frontend
docker-compose build frontend

# Rebuild all
docker-compose build
```

### Restart Service

```bash
docker-compose restart backend
docker-compose restart frontend
```

### Execute Commands in Container

```bash
# Backend shell
docker-compose exec backend sh

# Frontend shell
docker-compose exec frontend sh

# Database shell
docker-compose exec db psql -U postgres -d mini_asm
```

### Scale Services (if needed)

```bash
docker-compose up -d --scale backend=3
```

## 🧪 Testing the Deployment

### 1. Check Health

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "status": "healthy",
  "database": "connected"
}
```

### 2. Create an Asset

```bash
curl -X POST http://localhost:8080/assets \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "example.com",
    "type": "domain",
    "criticality": "high"
  }'
```

### 3. List Assets

```bash
curl http://localhost:8080/assets
```

### 4. Access Frontend

Open http://localhost:3000 in your browser

## 🐛 Troubleshooting

### Backend Can't Connect to Database

```bash
# Check database is ready
docker-compose logs db

# Verify backend env vars
docker-compose exec backend env | grep DB_
```

### Frontend Can't Reach Backend

```bash
# Check backend health
curl http://localhost:8080/health

# Check frontend logs
docker-compose logs frontend

# Verify nginx proxy config
docker-compose exec frontend cat /etc/nginx/conf.d/default.conf
```

### Port Already in Use

```bash
# Check what's using the port
sudo lsof -i :8080
sudo lsof -i :3000

# Stop the conflicting service or change port in docker-compose.yml
```

### Database Migration Issues

```bash
# Check migration files are copied
docker-compose exec backend ls -la migrations/

# Manually run migrations
docker-compose exec db psql -U postgres -d mini_asm -f /docker-entrypoint-initdb.d/001_create_assets.up.sql
```

### Clean Start (Nuclear Option)

```bash
# Stop everything
docker-compose down -v

# Remove all containers, images, and volumes
docker system prune -a --volumes

# Rebuild and start
docker-compose up -d --build
```

## 🔒 Production Considerations

### Security

1. **Change default passwords** in docker-compose.yml
2. **Use secrets management** for sensitive data
3. **Enable HTTPS** with reverse proxy (nginx/traefik)
4. **Implement rate limiting**
5. **Add authentication/authorization**

### Performance

1. **Set resource limits** in docker-compose.yml:

   ```yaml
   services:
     backend:
       deploy:
         resources:
           limits:
             cpus: "1"
             memory: 512M
   ```

2. **Use production-grade database** settings
3. **Enable connection pooling**
4. **Add caching layer** (Redis)

### Monitoring

1. **Add health checks** (already included)
2. **Implement logging** (ELK stack, Loki)
3. **Set up metrics** (Prometheus + Grafana)
4. **Add alerting** (AlertManager)

### Backup

```bash
# Database backup
docker-compose exec db pg_dump -U postgres mini_asm > backup.sql

# Restore
docker-compose exec -T db psql -U postgres mini_asm < backup.sql
```

## 📚 Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Nginx Documentation](https://nginx.org/en/docs/)
- [PostgreSQL Docker Hub](https://hub.docker.com/_/postgres)

## 🎓 Learning Objectives

This deployment setup demonstrates:

- ✅ Multi-stage Docker builds for optimization
- ✅ Service orchestration with Docker Compose
- ✅ Container networking and service discovery
- ✅ Volume management for data persistence
- ✅ Health checks for service reliability
- ✅ Reverse proxy configuration with Nginx
- ✅ Environment-based configuration
- ✅ Development vs production considerations
