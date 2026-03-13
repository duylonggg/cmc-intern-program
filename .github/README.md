# CI/CD Pipeline for Session 7 Deployment

This directory contains GitHub Actions workflows for automated testing and deployment of the Mini ASM application.

## Workflow: `session7-deployment.yml`

### Triggers

The workflow runs on:

- **Pull Requests** to `day3` or `main` branches
- **Push** to `day3` or `main` branches
- Only when files in `app/session7-deployment/**` are changed

### Pipeline Stages

```
┌─────────────────────────────────────────────────────────┐
│  1. Backend Build & Test                                │
│     - Go 1.24 setup                                     │
│     - Dependency verification                           │
│     - Build binary                                      │
│     - Run tests with coverage                           │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  2. Frontend Build & Test                               │
│     - Node.js 18 setup                                  │
│     - Install dependencies                              │
│     - Build production bundle                           │
│     - Verify build output                               │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  3. Docker Build                                        │
│     - Build backend Docker image                        │
│     - Build frontend Docker image                       │
│     - Use layer caching for speed                       │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  4. Integration Tests                                   │
│     - Start all services (Docker Compose)               │
│     - Wait for health checks                            │
│     - Test backend API endpoints                        │
│     - Test frontend accessibility                       │
│     - Test end-to-end functionality                     │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  5. Security Scan                                       │
│     - Trivy vulnerability scanning                      │
│     - Check for critical/high severity issues           │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  6. Build Report                                        │
│     - Generate summary of all stages                    │
│     - Display pass/fail status                          │
└─────────────────────────────────────────────────────────┘
```

### Jobs Breakdown

#### 1. **Backend Build & Test**

- Sets up Go 1.24 environment
- Verifies Go module dependencies
- Builds the backend binary
- Runs unit tests
- Generates code coverage report
- Uploads coverage to Codecov (optional)

**Working Directory**: `app/session7-deployment/backend`

#### 2. **Frontend Build & Test**

- Sets up Node.js 18 environment
- Installs dependencies with `npm ci` (clean install)
- Builds production bundle with Vite
- Verifies dist directory exists

**Working Directory**: `app/session7-deployment/frontend`

#### 3. **Docker Build**

- Uses Docker Buildx for multi-platform support
- Builds both backend and frontend Docker images
- Implements layer caching with GitHub Actions cache
- Does not push images (test build only)

**Images Built**:

- `mini-asm-backend:test`
- `mini-asm-frontend:test`

#### 4. **Integration Tests**

- Starts all services using Docker Compose
- Waits for services to be healthy (60s timeout)
- Tests backend endpoints:
  - Health check: `GET /health`
  - List assets: `GET /assets`
  - Create asset: `POST /assets`
- Tests frontend accessibility: `GET /`
- Shows logs on failure for debugging
- Cleans up with `docker compose down -v`

**Working Directory**: `app/session7-deployment`

#### 5. **Security Scan**

- Runs Trivy vulnerability scanner
- Scans both backend and frontend codebases
- Reports CRITICAL and HIGH severity issues
- Non-blocking (continues on error)

#### 6. **Build Report**

- Aggregates results from all jobs
- Generates GitHub summary report
- Shows pass/fail status for each stage
- Provides overall merge recommendation

### Environment Variables

The workflow uses these environment variables during integration tests:

```env
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=mini_asm
```

### Required Secrets

No secrets are required for the basic workflow. However, for optional features:

- `CODECOV_TOKEN` - For uploading coverage reports (optional)
- `DOCKER_HUB_USERNAME` - For pushing to Docker Hub (if needed)
- `DOCKER_HUB_TOKEN` - For Docker Hub authentication (if needed)

### Caching Strategy

The workflow uses GitHub Actions caching to speed up builds:

1. **Go Dependencies**: Cached based on `go.sum`
2. **Node Modules**: Cached based on `package-lock.json`
3. **Docker Layers**: Cached using GitHub Actions cache

### Testing Strategy

#### Unit Tests

```bash
# Backend
cd app/session7-deployment/backend
go test -v ./...
```

#### Integration Tests

- Database connectivity
- API endpoint functionality
- Frontend-backend communication
- Full request-response cycle

### Workflow Outputs

The workflow provides:

1. **GitHub Checks**: ✅ or ❌ for each job
2. **Summary Report**: Displayed in PR/commit summary
3. **Coverage Reports**: Uploaded to Codecov (optional)
4. **Logs**: Available for each job in GitHub Actions UI

### Viewing Results

1. **In Pull Request**:
   - Check status badges at the bottom
   - Click "Details" to see full logs
   - View summary in the "Checks" tab

2. **In Actions Tab**:
   - Navigate to repository > Actions
   - Select the workflow run
   - View each job's logs

### Local Testing

Before pushing, you can test locally:

```bash
# Test backend
cd app/session7-deployment/backend
go test -v ./...
go build -v -o main ./cmd/server

# Test frontend
cd app/session7-deployment/frontend
npm ci
npm run build

# Test Docker build
cd app/session7-deployment
docker compose build
docker compose up -d
# Run manual tests
docker compose down -v
```

### Troubleshooting

#### Job Fails: Backend Build

```bash
# Check Go version
go version  # Should be 1.24+

# Verify dependencies
cd app/session7-deployment/backend
go mod verify
go mod tidy
```

#### Job Fails: Frontend Build

```bash
# Check Node version
node --version  # Should be 18+

# Clean install
cd app/session7-deployment/frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

#### Job Fails: Docker Build

- Check Dockerfile syntax
- Verify .dockerignore doesn't exclude needed files
- Ensure package-lock.json exists for frontend

#### Job Fails: Integration Tests

- Check docker-compose.yml syntax
- Verify all environment variables are set
- Check port conflicts (8080, 3000, 5432)
- Review service logs in the workflow output

### Performance Optimization

Current optimization strategies:

1. **Parallel Execution**: Backend and frontend builds run in parallel
2. **Dependency Caching**: Go modules and npm packages cached
3. **Docker Layer Caching**: Reduces rebuild time
4. **Conditional Execution**: Only runs on relevant file changes

### Future Enhancements

Potential improvements:

- [ ] Add end-to-end tests with Playwright/Cypress
- [ ] Deploy to staging environment on merge
- [ ] Automated performance testing
- [ ] Database migration testing
- [ ] Load testing with k6
- [ ] Add SonarQube code quality checks
- [ ] Implement canary deployments
- [ ] Add rollback capability

### Badge for README

Add this to your README.md to show build status:

```markdown
[![Session 7 - Build and Test](https://github.com/dinhmanhtan/cmc-intern-program/actions/workflows/session7-deployment.yml/badge.svg)](https://github.com/dinhmanhtan/cmc-intern-program/actions/workflows/session7-deployment.yml)
```

## Branch Strategy

- **Feature branches** → PR to `day3` → triggers workflow
- **day3 branch** → PR to `main` → triggers workflow
- **main branch** → production-ready code

## Contributing

When creating a PR that affects Session 7:

1. Ensure all tests pass locally
2. Update documentation if needed
3. Wait for CI checks to pass
4. Address any security warnings
5. Get review approval before merging

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Trivy Security Scanner](https://github.com/aquasecurity/trivy-action)
