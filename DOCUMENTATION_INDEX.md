# 📚 Documentation Index - Storage Control Plane

Welcome to the Storage Control Plane documentation! This index helps you find the right guide for your needs.

## 🚀 Getting Started

| Guide | Description | When to Use |
|-------|-------------|-------------|
| **[QUICKSTART.md](QUICKSTART.md)** | 2-minute setup guide | First time setup, want to get running fast |
| **[README.md](README.md)** | Complete development guide | Full development workflow, comprehensive setup |
| **[LINUX_SETUP.md](LINUX_SETUP.md)** | Linux/Unix/WSL setup guide | Linux development, WSL, Ubuntu, Fedora, etc. |
| **[WINDOWS_SETUP.md](WINDOWS_SETUP.md)** | Windows-specific setup | Windows development, PowerShell, Chocolatey |

## 🧪 Testing & Development

| Guide | Description | When to Use |
|-------|-------------|-------------|
| **[TESTING.md](TESTING.md)** | Complete testing guide | Running tests, debugging test failures |
| **[AIR_SETUP.md](AIR_SETUP.md)** | Hot reload configuration | Air troubleshooting, custom configurations |

## 📋 Planning & Architecture

| Guide | Description | When to Use |
|-------|-------------|-------------|
| **[ROADMAP.md](ROADMAP.md)** | Implementation roadmap | Understanding project phases, planning work |

## 🔧 Quick Reference

### Essential Commands
```bash
# Development
air                    # Start with hot reload
go run ./cmd/api       # Start without hot reload
make dev              # Alternative hot reload start

# Testing  
make test             # Run all tests
make test-unit        # Unit tests only
make test-e2e         # End-to-end tests
.\test_e2e.ps1        # Windows E2E tests

# Building
make build            # Build binary
make clean            # Clean artifacts
```

### Key Configuration Files
```
.env                  # Environment variables
.env.example          # Environment template
.air.toml             # Hot reload config
Makefile              # Build automation
go.mod                # Go dependencies
```

### Directory Structure
```
storage_control_plane/
├── cmd/api/          # Application entry point
├── internal/         # Private application code
├── pkg/models/       # Shared data models
├── data/             # Local development data
├── tmp/              # Build artifacts (gitignored)
└── docs/             # Documentation
```

## 🎯 Use Case Guide

### "I want to start developing immediately"
1. Read [QUICKSTART.md](QUICKSTART.md)
2. Run `air`
3. Test with `.\test_e2e.ps1`

### "I want to understand the full system"
1. Read [README.md](README.md) - Development guide
2. Read [ROADMAP.md](ROADMAP.md) - Architecture overview
3. Read [TESTING.md](TESTING.md) - Testing strategy

### "I'm having issues with hot reload"
1. Read [AIR_SETUP.md](AIR_SETUP.md)
2. Check troubleshooting section
3. Verify `.air.toml` configuration

### "I want to run tests"
1. Read [TESTING.md](TESTING.md) 
2. Start with unit tests: `make test-unit`
3. Progress to E2E tests: `make test-e2e`

### "I want to contribute"
1. Read [README.md](README.md) - Development workflow
2. Read [ROADMAP.md](ROADMAP.md) - What needs building
3. Read [TESTING.md](TESTING.md) - How to test your changes

## 🆘 Troubleshooting Quick Index

| Problem | Check This Guide | Section |
|---------|------------------|---------|
| Air not working | [AIR_SETUP.md](AIR_SETUP.md) | Troubleshooting |
| Tests failing | [TESTING.md](TESTING.md) | Troubleshooting Tests |
| Build errors | [README.md](README.md) | Debugging |
| Port conflicts | [README.md](README.md) | Common Issues |
| Import errors | [README.md](README.md) | Common Issues |

## 📖 Documentation Standards

All documentation follows these principles:
- **Action-oriented** - Tell you what to do
- **Example-heavy** - Show actual commands and code
- **Troubleshooting-focused** - Address common issues
- **Windows-friendly** - PowerShell examples included
- **Copy-paste ready** - Commands work as-is

## 🔄 Document Updates

| Document | Last Updated | Next Review |
|----------|--------------|-------------|
| QUICKSTART.md | 2025-06-29 | Weekly |
| README.md | 2025-06-29 | Weekly |
| TESTING.md | 2025-06-29 | Bi-weekly |
| AIR_SETUP.md | 2025-06-29 | Monthly |
| ROADMAP.md | 2025-06-29 | Monthly |

## 💡 Documentation Tips

1. **Start with QUICKSTART** if you're new
2. **Use the search function** in your editor to find specific topics
3. **Copy commands exactly** - they're tested and work
4. **Check multiple guides** if you're stuck
5. **Update documentation** when you find issues

---

**Choose your guide and start building! 🚀**

*The Storage Control Plane team believes great documentation makes great software. Happy coding!*
