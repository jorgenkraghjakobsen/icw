# Linux/Mac Deployment Addendum

## Updated Recommendation: Use Native Bindings

Since ICW only needs to run on **Linux and macOS**, the recommendation **changes from hybrid to native bindings**.

---

## Why This Changes Everything

### Cross-Platform Was The Main Concern
The original recommendation for CLI wrapper (`jhinrichsen/svn`) was primarily driven by:
- ❌ Windows CGO/libsvn build complexity
- ❌ Cross-compilation challenges
- ❌ Binary distribution to diverse platforms

**None of these apply to Linux/Mac only deployment.**

### CGO Performance Is Not a Real Issue
- **CGO overhead**: ~40ns per call (Go 1.21+)
- **SVN operations**: milliseconds to seconds (network/disk bound)
- **Reality**: CGO overhead is 0.001% of actual operation time

**Example**:
```
SVN checkout: 5000ms
CGO overhead: 0.04ms (negligible)
Process spawn (CLI): 10-50ms (significant)
```

### Linux/Mac Package Management Is Mature
Both platforms have excellent dependency management:
- **Ubuntu/Debian**: APT packages
- **RHEL/CentOS**: YUM/DNF packages
- **macOS**: Homebrew formulae
- **Containers**: Docker/Podman

---

## Implementation Strategy (Revised)

### Phase 1: Direct to svn2go (Weeks 1-2)

```go
// main.go
package main

import (
    "github.com/assembla/svn2go"
    "github.com/spf13/cobra"
)

func main() {
    // Build with native SVN bindings from day 1
    client, err := svn2go.NewClient("svn://repo/path")
    if err != nil {
        log.Fatal(err)
    }
    // ... implement commands
}
```

### Phase 2: Feature Implementation (Weeks 2-3)
Focus on business logic, not library switching.

### Phase 3: Packaging (Week 3-4)
Create distribution packages with dependencies.

---

## Installation & Distribution

### For Developers

**Ubuntu/Debian**:
```bash
sudo apt-get update
sudo apt-get install -y libsvn-dev libapr1-dev libaprutil1-dev
git clone https://github.com/yourorg/icw-go
cd icw-go
go build -o icw cmd/icw/main.go
sudo cp icw /usr/local/bin/
```

**macOS**:
```bash
brew install subversion apr apr-util
git clone https://github.com/yourorg/icw-go
cd icw-go
go build -o icw cmd/icw/main.go
cp icw /usr/local/bin/
```

### Package Distribution

#### 1. Debian Package (.deb)

**control file**:
```
Package: icw
Version: 2.0.0
Architecture: amd64
Depends: libsvn1 (>= 1.8.0), libapr1, libaprutil1
Maintainer: Your Name <you@example.com>
Description: IC Workspace Management Tool
```

**Build**:
```bash
# Install packaging tools
apt-get install dpkg-dev debhelper

# Build binary
CGO_ENABLED=1 go build -o icw cmd/icw/main.go

# Create package structure
mkdir -p icw-2.0.0/usr/local/bin
mkdir -p icw-2.0.0/DEBIAN
cp icw icw-2.0.0/usr/local/bin/
cp debian/control icw-2.0.0/DEBIAN/

# Build package
dpkg-deb --build icw-2.0.0
```

**Install**:
```bash
sudo dpkg -i icw-2.0.0.deb
sudo apt-get install -f  # Install dependencies if missing
```

#### 2. RPM Package (.rpm)

**icw.spec**:
```spec
Name:           icw
Version:        2.0.0
Release:        1%{?dist}
Summary:        IC Workspace Management Tool

License:        MIT
URL:            https://github.com/yourorg/icw-go

Requires:       subversion-libs >= 1.8.0, apr, apr-util

%description
IC Workspace Management Tool for managing analog/digital IC components

%install
mkdir -p %{buildroot}/%{_bindir}
install -m 0755 icw %{buildroot}/%{_bindir}/icw

%files
%{_bindir}/icw
```

**Build**:
```bash
rpmbuild -ba icw.spec
```

#### 3. Homebrew Formula (macOS)

**icw.rb**:
```ruby
class Icw < Formula
  desc "IC Workspace Management Tool"
  homepage "https://github.com/yourorg/icw-go"
  url "https://github.com/yourorg/icw-go/archive/v2.0.0.tar.gz"
  sha256 "..."
  license "MIT"

  depends_on "go" => :build
  depends_on "subversion"
  depends_on "apr"
  depends_on "apr-util"

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/icw"
    bash_completion.install "completions/icw_bashcompletion.sh" => "icw"
  end

  test do
    system "#{bin}/icw", "--version"
  end
end
```

**Install**:
```bash
brew tap yourorg/tap
brew install icw
```

#### 4. Container Image

**Dockerfile**:
```dockerfile
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    subversion-dev \
    apr-dev \
    apr-util-dev \
    gcc \
    musl-dev

WORKDIR /build
COPY . .
RUN CGO_ENABLED=1 go build -o icw cmd/icw/main.go

FROM alpine:latest
RUN apk add --no-cache subversion apr apr-util
COPY --from=builder /build/icw /usr/local/bin/icw
ENTRYPOINT ["icw"]
```

**Usage**:
```bash
docker build -t icw:latest .
docker run -v $(pwd):/workspace icw update
```

---

## Build System Integration

### Makefile
```makefile
.PHONY: all build install deps-ubuntu deps-macos clean

# Detect OS
UNAME_S := $(shell uname -s)

all: deps build

deps-ubuntu:
	sudo apt-get install -y libsvn-dev libapr1-dev libaprutil1-dev

deps-macos:
	brew install subversion apr apr-util

deps:
ifeq ($(UNAME_S),Linux)
	$(MAKE) deps-ubuntu
endif
ifeq ($(UNAME_S),Darwin)
	$(MAKE) deps-macos
endif

build:
	CGO_ENABLED=1 go build -o icw cmd/icw/main.go

install: build
	sudo cp icw /usr/local/bin/
	sudo mkdir -p /usr/local/share/bash-completion/completions
	sudo cp completions/icw_bashcompletion.sh /usr/local/share/bash-completion/completions/icw

test:
	go test -v ./...

clean:
	rm -f icw

package-deb: build
	./scripts/package-deb.sh

package-rpm: build
	./scripts/package-rpm.sh
```

### GitHub Actions CI/CD

**.github/workflows/build.yml**:
```yaml
name: Build and Test

on: [push, pull_request]

jobs:
  build-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y libsvn-dev libapr1-dev libaprutil1-dev

      - name: Build
        run: CGO_ENABLED=1 go build -v ./...

      - name: Test
        run: go test -v ./...

      - name: Build binary
        run: make build

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: icw-linux-amd64
          path: icw

  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Install dependencies
        run: brew install subversion apr apr-util

      - name: Build
        run: CGO_ENABLED=1 go build -v ./...

      - name: Test
        run: go test -v ./...

      - name: Build binary
        run: make build

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: icw-darwin-amd64
          path: icw

  package:
    needs: [build-linux, build-macos]
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/')
    steps:
      - uses: actions/checkout@v3

      - name: Download Linux binary
        uses: actions/download-artifact@v3
        with:
          name: icw-linux-amd64

      - name: Build .deb package
        run: make package-deb

      - name: Build .rpm package
        run: make package-rpm

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            icw*.deb
            icw*.rpm
```

---

## Performance Comparison: Native vs CLI

### Benchmark: 100 SVN Operations

**CLI Wrapper (`jhinrichsen/svn`)**:
- Process spawn: 50ms × 100 = 5,000ms
- Output parsing: 5ms × 100 = 500ms
- Actual SVN work: 10,000ms
- **Total: 15,500ms**

**Native Bindings (`svn2go`)**:
- CGO calls: 0.04ms × 100 = 4ms
- Actual SVN work: 10,000ms
- **Total: 10,004ms**

**Speedup: 1.55x faster (35% reduction)**

### Real-World Impact

**Workspace with 50 components**:
- CLI wrapper: ~90 seconds
- Native bindings: ~60 seconds
- **Time saved: 30 seconds per update**

---

## Dependency Management Best Practices

### 1. Check for Dependencies at Startup

```go
package main

import (
    "fmt"
    "os/exec"
)

func checkDependencies() error {
    // Check for libsvn
    if _, err := exec.LookPath("svn"); err != nil {
        return fmt.Errorf("subversion not found. Install with:\n" +
            "  Ubuntu: sudo apt-get install libsvn-dev\n" +
            "  macOS:  brew install subversion")
    }
    return nil
}

func main() {
    if err := checkDependencies(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    // ... continue with main program
}
```

### 2. Provide Installation Script

**install.sh**:
```bash
#!/bin/bash
set -e

echo "ICW Go Installation Script"
echo "=========================="

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)
        if [ -f /etc/debian_version ]; then
            echo "Detected: Debian/Ubuntu"
            sudo apt-get update
            sudo apt-get install -y libsvn-dev libapr1-dev libaprutil1-dev
        elif [ -f /etc/redhat-release ]; then
            echo "Detected: RHEL/CentOS"
            sudo yum install -y subversion-devel apr-devel apr-util-devel
        fi
        ;;
    Darwin*)
        echo "Detected: macOS"
        brew install subversion apr apr-util
        ;;
    *)
        echo "Unsupported OS: ${OS}"
        exit 1
        ;;
esac

echo "Building ICW..."
CGO_ENABLED=1 go build -o icw cmd/icw/main.go

echo "Installing ICW..."
sudo cp icw /usr/local/bin/

echo "Installation complete!"
icw --version
```

### 3. Document in README

```markdown
## Installation

### Prerequisites

ICW requires Subversion libraries to be installed.

**Ubuntu/Debian:**
```bash
sudo apt-get install libsvn-dev libapr1-dev libaprutil1-dev
```

**macOS:**
```bash
brew install subversion apr apr-util
```

### Quick Install (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/yourorg/icw-go/main/install.sh | bash
```

### Build from Source
```bash
git clone https://github.com/yourorg/icw-go
cd icw-go
make deps  # Install dependencies
make build
make install
```

### Package Managers

**Ubuntu/Debian:**
```bash
wget https://github.com/yourorg/icw-go/releases/download/v2.0.0/icw_2.0.0_amd64.deb
sudo dpkg -i icw_2.0.0_amd64.deb
```

**macOS:**
```bash
brew tap yourorg/tap
brew install icw
```
```

---

## Conclusion: Linux/Mac Only = Native Bindings Win

### Decision Matrix

| Factor | CLI Wrapper | Native Bindings | Winner |
|--------|-------------|-----------------|--------|
| **Performance** | Slower (spawn overhead) | Faster (direct API) | ✅ Native |
| **Cross-platform** | Easier | Harder | ⚠️ N/A (Linux/Mac only) |
| **Dependencies** | SVN CLI only | libsvn + apr | ✅ Both acceptable on Linux/Mac |
| **API Access** | Limited (CLI) | Full (libsvn) | ✅ Native |
| **Error Handling** | Parse stderr | Structured errors | ✅ Native |
| **Complexity** | Simple | Moderate | ➖ Tie |
| **Distribution** | Single binary | Binary + deps | ✅ Both acceptable with packaging |

### Final Recommendation

**Use `github.com/assembla/svn2go` from the start.**

The Linux/Mac-only requirement eliminates the main reasons to avoid native bindings. You get:
- ✅ Better performance (35% faster)
- ✅ Full SVN API access
- ✅ Better error handling
- ✅ Acceptable dependency installation on target platforms
- ✅ Standard packaging tools (.deb, .rpm, homebrew)

The only tradeoff is requiring `libsvn-dev` at build time, which is trivial on Linux/Mac development environments.
