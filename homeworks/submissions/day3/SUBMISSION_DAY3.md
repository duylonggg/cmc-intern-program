# Homework Submission - Day 3

**Họ tên:** Hà Duy Long

## Các bài đã hoàn thành

- [ ] Bài 1: Mở rộng Scan API
- [ ] Bài 2: Viết Unit Tests
- [x] Bài 3: Tích hợp Frontend
- [x] Bài 4: CI/CD với GitHub Actions
- [x] Bài 5: Deploy với Docker Compose
- [ ] Bài 6: Tính năng EASM mới (Bonus)
- [ ] Bài 7: Deploy lên Cloud VM (Bonus)
- [ ] Bài 8: Domain & TLS/HTTPS (Bonus)
- [ ] Bài 9: Auto Deploy on Merge (Bonus)

## Link Repository

https://github.com/duylonggg/cmc-intern-program

## Link Demo (nếu có)

Chưa có public demo URL.

---

## Bằng chứng theo từng bài

### Bài 1: Mở rộng Scan API — **Chưa hoàn thành đầy đủ**

Đã có mở rộng một phần scan type trong model:

- `dns`, `whois`, `subdomain`, `cert_trans`, `asn`, `port`, `ssl`, `all`
- File: `app/session7-deployment/backend/internal/model/scan.go`

Nhưng chưa thấy triển khai đầy đủ theo đề Day3:

- Chưa có `scan_type: ip`
- Chưa có `scan_type: tech`
- `port_scanner_template.go` vẫn là template và trả về `not implemented`
- Luồng service scan chủ yếu xử lý `all/dns/whois/subdomain`

### Bài 2: Viết Unit Tests — **Chưa hoàn thành đầy đủ**

Đã có:

- Model tests:
  - `app/session7-deployment/backend/internal/model/asset_test.go`
  - `app/session7-deployment/backend/internal/model/scan_test.go`
- Validator tests:
  - `app/session7-deployment/backend/internal/validator/asset_validator_test.go`

Chưa có:

- Scanner tests bắt buộc theo đề (không có file `*_test.go` trong `internal/scanner`)

### Bài 3: Tích hợp Frontend — **Đã hoàn thành**

Đã có frontend tại:

- `app/session7-deployment/frontend`

Có các script build/dev:

```bash
npm ci
npm run build
```

Kết quả build local (đã chạy):

```text
vite v5.4.21 building for production...
✓ built in 1.65s
```

### Bài 4: CI/CD với GitHub Actions — **Đã hoàn thành**

Đã có workflow:

- `.github/workflows/session7-deployment.yml`

Bao gồm các job chính:

- Backend build + test + coverage
- Frontend build
- Docker image build
- Docker compose integration test

### Bài 5: Deploy với Docker Compose — **Đã hoàn thành**

Đã có cấu hình deploy stack:

- `app/session7-deployment/docker-compose.yml`

Bao gồm các service:

- `db` (PostgreSQL)
- `backend` (Go API)
- `frontend` (React + Nginx)

---

## Command outputs đã xác minh trong repo

### Backend tests

```bash
cd app/session7-deployment/backend
go test ./...
```

Kết quả:

```text
ok  	mini-asm/internal/model      0.003s
ok  	mini-asm/internal/validator  0.005s
```

### Frontend build

```bash
cd app/session7-deployment/frontend
npm ci
npm run build
```

Kết quả:

```text
vite v5.4.21 building for production...
✓ built in 1.65s
```
