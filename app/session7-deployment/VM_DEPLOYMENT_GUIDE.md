# Hướng Dẫn Deployment Lên VM với Nginx và Certbot

## Tổng Quan

Setup này sử dụng **2 lớp Nginx**:

- **Nginx trong Docker** (port 80 → map ra 3000): Serve frontend static files + proxy API calls đến backend container
- **Nginx trên VM**: Reverse proxy + SSL termination với Certbot
- **Docker** để chạy backend (Go API) và database (PostgreSQL)

### Tại sao 2 lớp Nginx?

1. **Nginx trong Docker container**:
   - Serve React static files (HTML, CSS, JS)
   - Proxy `/api/*` requests đến backend container
   - Portable và consistent trong mọi môi trường

2. **Nginx trên VM**:
   - SSL/TLS termination với Let's Encrypt (Certbot)
   - Rate limiting, DDoS protection
   - Load balancing (nếu scale nhiều containers)
   - Centralized logging và monitoring

## Kiến Trúc

```
Internet (HTTPS) → Nginx (VM) + Certbot SSL
                        ↓
                   (HTTP internal)
                        ↓
            ┌───────────┴───────────┐
            ↓                       ↓
    Frontend:3000              Backend:8080
    (Nginx container)          (Go API)
    - Static files                 ↓
    - Proxy /api/*          PostgreSQL:5432
```

**Flow request**:

1. Browser → `https://domain.com/` → Nginx VM (SSL) → Nginx container (port 3000) → Serve index.html
2. Browser → `https://domain.com/api/health` → Nginx VM → Nginx container `/api/*` → Backend:8080

## Bước 1: Chuẩn Bị VM

### Yêu Cầu

- Ubuntu 20.04+ hoặc Debian 11+
- Docker và Docker Compose đã cài đặt
- Domain name đã trỏ về VM (cho SSL)

### Cài Đặt Docker

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify
docker --version
docker-compose --version
```

### Cài Đặt Nginx

```bash
sudo apt install nginx -y
sudo systemctl enable nginx
sudo systemctl start nginx
```

### Cài Đặt Certbot

```bash
sudo apt install certbot python3-certbot-nginx -y
```

## Bước 2: Deploy Ứng Dụng

### Clone Repository

```bash
cd /opt
sudo git clone https://github.com/dinhmanhtan/cmc-intern-program.git
sudo chown -R $USER:$USER cmc-intern-program
cd cmc-intern-program/app/session7-deployment
```

### Cấu Hình Environment Variables

```bash
# Tạo .env cho backend (tùy chọn, có thể dùng docker-compose.yml)
cat > backend/.env << EOF
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=YOUR_SECURE_PASSWORD_HERE
DB_NAME=mini_asm
EOF
```

### Cập Nhật docker-compose.yml

```bash
# Sửa password trong docker-compose.yml
nano docker-compose.yml

# Đổi:
# POSTGRES_PASSWORD: postgres@123
# DB_PASSWORD: postgres@123
#
# Thành password mạnh hơn
```

### Start Services

```bash
# Build và start containers
docker-compose up -d

# Kiểm tra status
docker-compose ps

# Xem logs
docker-compose logs -f
```

## Bước 3: Cấu Hình Nginx

### Tạo Nginx Config

```bash
sudo nano /etc/nginx/sites-available/mini-asm
```

Nội dung file (thay `your-domain.com` bằng domain của bạn):

```nginx
# Mini ASM - Nginx Configuration
upstream frontend {
    server localhost:3000;
}

upstream backend {
    server localhost:8080;
}

server {
    listen 80;
    listen [::]:80;
    server_name your-domain.com www.your-domain.com;

    # Redirect HTTP to HTTPS (sẽ kích hoạt sau khi có SSL)
    # return 301 https://$server_name$request_uri;

    # Client max body size
    client_max_body_size 10M;

    # Frontend
    location / {
        # Proxy tất cả requests đến Nginx container
        # Nginx container sẽ serve static files HOẶC proxy /api/* đến backend
        proxy_pass http://frontend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # KHÔNG CẦN proxy /api/ riêng ở đây
    # Vì Nginx container (frontend:3000) đã handle /api/* và proxy đến backend
    # Nginx VM chỉ cần forward TẤT CẢ requests đến frontend container

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1000;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    # Logging
    access_log /var/log/nginx/mini-asm-access.log;
    error_log /var/log/nginx/mini-asm-error.log;
}
```

### Kích Hoạt Site

```bash
# Tạo symbolic link
sudo ln -s /etc/nginx/sites-available/mini-asm /etc/nginx/sites-enabled/

# Xóa default site (tùy chọn)
sudo rm /etc/nginx/sites-enabled/default

# Test config
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

## Bước 4: Cấu Hình SSL với Certbot

### Lấy SSL Certificate

```bash
# Chạy Certbot
sudo certbot --nginx -d your-domain.com -d www.your-domain.com

# Làm theo hướng dẫn:
# 1. Nhập email
# 2. Đồng ý terms of service
# 3. Chọn redirect HTTP to HTTPS (option 2)
```

Certbot sẽ tự động:

- Lấy SSL certificate từ Let's Encrypt
- Cập nhật Nginx config
- Setup auto-renewal

### Verify SSL

```bash
# Kiểm tra certificate
sudo certbot certificates

# Test renewal
sudo certbot renew --dry-run
```

### Auto-renewal

Certbot tự động tạo cron job hoặc systemd timer. Kiểm tra:

```bash
# Systemd timer
sudo systemctl status certbot.timer

# Hoặc cron
sudo crontab -l
```

## Bước 5: Cập Nhật Frontend Config

Cập nhật frontend để sử dụng `/api` thay vì direct URL:

```bash
# File: frontend/src/services/api.js
# Đảm bảo API calls sử dụng /api prefix:
# const API_BASE = '/api'
```

Rebuild frontend nếu cần:

```bash
cd /opt/cmc-intern-program/app/session7-deployment
docker-compose build frontend
docker-compose up -d frontend
```

## Bước 6: Testing

### Test Local

```bash
# Health check
curl http://localhost:8080/health

# Frontend
curl http://localhost:3000
```

### Test qua Nginx

```bash
# HTTP (trước khi có SSL)
curl http://your-domain.com
curl http://your-domain.com/api/health

# HTTPS (sau khi có SSL)
curl https://your-domain.com
curl https://your-domain.com/api/health
```

### Test trên Browser

1. Mở `https://your-domain.com`
2. Kiểm tra SSL certificate (icon khóa trên address bar)
3. Test các chức năng của ứng dụng

## Bước 7: Monitoring & Logs

### Docker Logs

```bash
# Xem tất cả logs
cd /opt/cmc-intern-program/app/session7-deployment
docker-compose logs -f

# Log specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f db
```

### Nginx Logs

```bash
# Access logs
sudo tail -f /var/log/nginx/mini-asm-access.log

# Error logs
sudo tail -f /var/log/nginx/mini-asm-error.log
```

### System Resources

```bash
# Container stats
docker stats

# Disk usage
docker system df
df -h
```

## Bước 8: Backup & Maintenance

### Database Backup

```bash
# Tạo backup directory
mkdir -p /opt/backups/mini-asm

# Backup database
docker-compose exec -T db pg_dump -U postgres mini_asm > /opt/backups/mini-asm/backup_$(date +%Y%m%d_%H%M%S).sql

# Hoặc dùng script tự động
cat > /opt/backup-db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/backups/mini-asm"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
cd /opt/cmc-intern-program/app/session7-deployment
docker-compose exec -T db pg_dump -U postgres mini_asm > "$BACKUP_DIR/backup_$TIMESTAMP.sql"
# Keep only last 7 days
find "$BACKUP_DIR" -name "backup_*.sql" -mtime +7 -delete
EOF

chmod +x /opt/backup-db.sh

# Add to crontab (daily at 2 AM)
echo "0 2 * * * /opt/backup-db.sh" | sudo crontab -
```

### Update Application

```bash
cd /opt/cmc-intern-program/app/session7-deployment

# Pull latest code
git pull origin main

# Rebuild và restart
docker-compose build
docker-compose up -d

# Verify
docker-compose ps
```

### SSL Certificate Renewal

Certbot tự động renew, nhưng có thể manual:

```bash
sudo certbot renew
sudo systemctl reload nginx
```

## Troubleshooting

### Lỗi: "502 Bad Gateway"

```bash
# Kiểm tra containers có chạy không
docker-compose ps

# Kiểm tra logs
docker-compose logs backend
docker-compose logs frontend

# Restart containers
docker-compose restart
```

### Lỗi: "Connection Refused"

```bash
# Kiểm tra port đang lắng nghe
sudo netstat -tulpn | grep -E ':3000|:8080'

# Kiểm tra firewall
sudo ufw status
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

### Lỗi SSL Certificate

```bash
# Kiểm tra certificate
sudo certbot certificates

# Renew manually
sudo certbot renew --force-renewal

# Restart nginx
sudo systemctl restart nginx
```

### Database Connection Issues

```bash
# Kiểm tra database
docker-compose exec db psql -U postgres -d mini_asm -c "SELECT 1;"

# Restart database
docker-compose restart db

# Check logs
docker-compose logs db
```

## Security Checklist

- [ ] Đổi default passwords
- [ ] Setup firewall (ufw)
- [ ] Enable fail2ban
- [ ] Regular security updates
- [ ] Backup database định kỳ
- [ ] Monitor logs
- [ ] Limit SSH access
- [ ] Use strong SSL settings

## Firewall Setup

```bash
# Install ufw
sudo apt install ufw -y

# Allow SSH (QUAN TRỌNG!)
sudo ufw allow ssh

# Allow HTTP & HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

## Performance Tuning

### Nginx

```bash
# Edit nginx.conf
sudo nano /etc/nginx/nginx.conf

# Tăng worker_connections
events {
    worker_connections 2048;
}

# Enable caching
http {
    proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m max_size=1g inactive=60m;
}
```

### Docker Resources

```bash
# Giới hạn resources trong docker-compose.yml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
```

## Useful Commands

```bash
# Restart all services
docker-compose restart

# Stop all services
docker-compose down

# Start with rebuild
docker-compose up -d --build

# View container IPs
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' mini-asm-backend

# Clean up old images
docker system prune -a

# Nginx reload
sudo systemctl reload nginx

# Nginx restart
sudo systemctl restart nginx
```

## Support

Nếu gặp vấn đề:

1. Kiểm tra logs: `docker-compose logs` và `/var/log/nginx/`
2. Verify config: `sudo nginx -t`
3. Check ports: `sudo netstat -tulpn`
4. Review firewall: `sudo ufw status`

---

**Lưu ý**: Thay thế `your-domain.com` bằng domain thật của bạn trong tất cả các config!
