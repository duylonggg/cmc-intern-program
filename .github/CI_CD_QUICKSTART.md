# 🚀 CI/CD Quick Start Guide

## Overview

This project uses **GitHub Actions** for automated testing and deployment of Session 7 (Mini ASM application).

## ✅ What Gets Tested

When you create a PR to `day3` or `main` branch that modifies Session 7 files:

1. ✅ **Backend** - Go build and tests
2. ✅ **Frontend** - React build and validation
3. ✅ **Docker** - Image builds
4. ✅ **Integration** - Full stack testing
5. ✅ **Security** - Vulnerability scanning

## 🔧 Setting Up

### Prerequisites

Ensure your development environment has:

- Git
- Go 1.24+
- Node.js 18+
- Docker & Docker Compose

### First Time Setup

```bash
# Clone the repository
git clone https://github.com/dinhmanhtan/cmc-intern-program.git
cd cmc-intern-program

# Create a feature branch
git checkout -b feature/my-changes

# Make your changes to session7-deployment
cd app/session7-deployment
```

## 📝 Workflow for Contributors

### 1. Create Feature Branch

```bash
# Always branch from day3 or main
git checkout day3
git pull origin day3
git checkout -b feature/add-new-endpoint
```

### 2. Make Changes

```bash
# Work on Session 7
cd app/session7-deployment

# Test locally before committing
make test  # or docker compose up -d
```

### 3. Test Locally

```bash
# Backend tests
cd backend
go test -v ./...
go build -v -o main ./cmd/server

# Frontend tests
cd ../frontend
npm ci
npm run build

# Docker build
cd ..
docker compose build
docker compose up -d

# Manual API testing
curl http://localhost:8080/health
curl http://localhost:8080/assets

# Clean up
docker compose down -v
```

### 4. Commit and Push

```bash
git add .
git commit -m "feat: add new endpoint for asset filtering"
git push origin feature/add-new-endpoint
```

### 5. Create Pull Request

1. Go to GitHub repository
2. Click "Compare & pull request"
3. Select base branch: `day3` or `main`
4. Fill in PR description
5. Create PR

### 6. Wait for CI Checks

The workflow will automatically:

- Build backend and frontend
- Run tests
- Build Docker images
- Run integration tests
- Scan for vulnerabilities

**Expected time**: 5-10 minutes

### 7. Review Results

Check the PR page for:

- ✅ Green checkmarks = All tests passed
- ❌ Red X = Some tests failed

Click "Details" to see logs if tests fail.

### 8. Fix Issues (if needed)

If tests fail:

```bash
# Check the error in GitHub Actions logs
# Fix the issue locally
# Test again
make test

# Commit fix
git add .
git commit -m "fix: resolve test failure"
git push origin feature/add-new-endpoint

# CI will automatically re-run
```

### 9. Get Approval & Merge

Once all checks pass:

1. Request review from team member
2. Address review comments
3. Get approval
4. Merge PR

## 🎯 Common Scenarios

### Scenario 1: Adding New Backend Endpoint

```bash
# 1. Create feature branch
git checkout -b feature/new-endpoint

# 2. Add handler in backend/internal/handler/
# 3. Add route in backend/cmd/server/main.go
# 4. Add tests

# 5. Test locally
cd app/session7-deployment/backend
go test -v ./internal/handler/

# 6. Commit and push
git add .
git commit -m "feat: add new endpoint"
git push origin feature/new-endpoint

# 7. Create PR → CI runs automatically
```

### Scenario 2: Updating Frontend UI

```bash
# 1. Create feature branch
git checkout -b feature/ui-update

# 2. Modify frontend/src/pages/*.jsx
# 3. Test locally
cd app/session7-deployment/frontend
npm run dev  # Check at localhost:3000

# 4. Build to verify
npm run build

# 5. Commit and push
git add .
git commit -m "feat: improve UI layout"
git push origin feature/ui-update

# 6. Create PR → CI runs automatically
```

### Scenario 3: Fixing Docker Configuration

```bash
# 1. Create feature branch
git checkout -b fix/docker-config

# 2. Modify docker-compose.yml or Dockerfile
# 3. Test locally
cd app/session7-deployment
docker compose build
docker compose up -d

# 4. Verify services work
curl http://localhost:8080/health
curl http://localhost:3000

# 5. Clean up
docker compose down -v

# 6. Commit and push
git add .
git commit -m "fix: update Docker health checks"
git push origin fix/docker-config

# 7. Create PR → CI runs automatically
```

## 🐛 Troubleshooting CI Failures

### Backend Build Fails

**Error**: "go: module not found" or "build failed"

**Solution**:

```bash
cd app/session7-deployment/backend
go mod tidy
go mod verify
git add go.mod go.sum
git commit -m "fix: update go modules"
git push
```

### Frontend Build Fails

**Error**: "npm ci failed" or "build command failed"

**Solution**:

```bash
cd app/session7-deployment/frontend
rm -rf node_modules package-lock.json
npm install
git add package-lock.json
git commit -m "fix: update package-lock.json"
git push
```

### Docker Build Fails

**Error**: "Dockerfile syntax error" or "COPY failed"

**Solution**:

1. Check Dockerfile syntax
2. Verify files exist before COPY
3. Check .dockerignore isn't excluding needed files
4. Test locally: `docker compose build`

### Integration Tests Fail

**Error**: "Service not responding" or "Health check timeout"

**Solution**:

1. Check service logs in GitHub Actions
2. Verify environment variables
3. Test locally with: `docker compose up -d`
4. Check backend logs: `docker compose logs backend`

### Security Scan Warnings

**Error**: "High severity vulnerability found"

**Solution**:

```bash
# Update dependencies
cd app/session7-deployment/backend
go get -u ./...
go mod tidy

cd ../frontend
npm audit fix
npm update

# Test and commit
git add .
git commit -m "chore: update dependencies for security"
git push
```

## 📊 Understanding CI Status

### Status Badges Meaning

- **⏳ Yellow (In Progress)**: Tests are running
- **✅ Green (Success)**: All tests passed
- **❌ Red (Failure)**: Some tests failed
- **⚪ Gray (Skipped)**: Tests skipped (no changes)

### Job Status in PR

| Job                   | What It Tests               | Typical Duration |
| --------------------- | --------------------------- | ---------------- |
| Backend Build & Test  | Go compilation + unit tests | 1-2 min          |
| Frontend Build & Test | React build + validation    | 1-2 min          |
| Docker Build          | Multi-stage image builds    | 2-3 min          |
| Integration Tests     | Full stack + API testing    | 2-3 min          |
| Security Scan         | Vulnerability check         | 1-2 min          |

**Total**: ~5-10 minutes

## 🔑 Best Practices

### Before Committing

1. ✅ Test locally first
2. ✅ Run unit tests
3. ✅ Build Docker images
4. ✅ Check code quality
5. ✅ Update documentation

### Writing Commit Messages

Follow conventional commits:

```bash
feat: add new scanning endpoint
fix: resolve database connection issue
docs: update API documentation
chore: update dependencies
test: add integration tests
refactor: improve error handling
```

### Pull Request Checklist

- [ ] All tests pass locally
- [ ] Docker build succeeds
- [ ] No console errors/warnings
- [ ] Documentation updated
- [ ] Commit messages are clear
- [ ] PR description explains changes

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Session 7 README](../app/session7-deployment/README.md)
- [Docker Guide](../app/session7-deployment/DOCKER_GUIDE.md)
- [Contributing Guidelines](../CONTRIBUTING.md)

## 💡 Tips

1. **Push early, push often** - CI catches issues faster
2. **Read the logs** - Error messages are usually clear
3. **Test locally first** - Saves CI time
4. **Keep PRs small** - Easier to review and test
5. **Ask for help** - Check with team if stuck

## 🆘 Getting Help

If CI issues persist:

1. Check [GitHub Actions Logs](https://github.com/dinhmanhtan/cmc-intern-program/actions)
2. Review [Troubleshooting Guide](README.md#troubleshooting)
3. Ask in team chat/slack
4. Create an issue with CI logs attached

---

**Remember**: Green checks = Happy merge! ✅
