# Release Guide

## 1. Create the GitHub Repository

The Go module is declared as `github.com/archforge/cli` in `go.mod`, so the repository must be exactly:

```
https://github.com/archforge/cli
```

Steps:
1. Create the `archforge` organization on GitHub (or use a personal account and rename the module)
2. Create the `cli` repository (public)
3. Push the code:
   ```bash
   git init
   git remote add origin git@github.com:archforge/cli.git
   git push
   ```

Two additional repositories are required — GoReleaser will write to them automatically at release time:
- `archforge/homebrew-tap` — where the Homebrew formula lives
- `archforge/scoop-bucket` — where the Scoop manifest lives

---

## 2. Configure GitHub Actions Secrets

The `.goreleaser.yaml` references these environment variables. Set them in `Settings → Secrets → Actions` on the repository:

| Secret | Purpose |
|---|---|
| `GITHUB_TOKEN` | Create the GitHub Release and upload binaries |
| `HOMEBREW_TAP_TOKEN` | Push the formula to the `homebrew-tap` repo |
| `SCOOP_BUCKET_TOKEN` | Push the manifest to the `scoop-bucket` repo |
| `DOCKER_USERNAME` | Publish images to Docker Hub |
| `DOCKER_PASSWORD` | Publish images to Docker Hub |

---

## 3. Tag and Release

GoReleaser is triggered by a semver tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

To test locally before publishing:

```bash
# Simulate everything without publishing
make release-dry-run

# Build real binaries without publishing
make release-snapshot
```

In CI, when the tag reaches GitHub, the workflow runs:

```bash
make release   # → goreleaser release --clean
```

---

## 4. What GoReleaser Does

In order:

1. Compiles 6 binaries (linux/darwin/windows × amd64/arm64) with the correct ldflags
2. Packages each one as `.tar.gz` or `.zip`
3. Generates `checksums.txt` with the SHA-256 of each file
4. Builds Docker images for amd64 and arm64 and creates the multi-arch manifest
5. Uploads everything to the GitHub Release as downloadable assets
6. Pushes the updated Homebrew formula (with real URLs and checksums) to the `homebrew-tap` repo
7. Pushes the Scoop manifest to the `scoop-bucket` repo

After the release, users can install with:

```bash
# macOS / Linux
brew install archforge/tap/arch-forge

# Windows
scoop bucket add archforge https://github.com/archforge/scoop-bucket
scoop install arch-forge

# Docker
docker run --rm archforge/cli:1.0.0 --help
```

---

## 5. Before You Release

The most critical check: **the Go module path must match the actual repository URL.**

If using a personal account instead of an organization, update `go.mod` and all internal imports from:
```
github.com/archforge/cli
```
to:
```
github.com/yourusername/arch_forge
```

Run `make release-dry-run` to catch any GoReleaser configuration issues before tagging.
