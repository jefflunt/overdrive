#!/bin/bash

# scripts/setup-macos.sh - Setup script for Overdrive on macOS (Apple Silicon)

set -e

REQUIRED_GO_VERSION="1.24.5"
PORT=3281

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Overdrive macOS Setup ===${NC}"

# 1. Architecture Check
ARCH=$(uname -m)
if [ "$ARCH" != "arm64" ]; then
    echo -e "${RED}Error: This script is intended for Apple Silicon (arm64). Found: $ARCH${NC}"
    exit 1
fi

# 2. Discovery Phase
discover_dependencies() {
    # Homebrew
    if command -v brew &> /dev/null; then
        BREW_STATUS="${GREEN}[FOUND]${NC} $(brew --version | head -n 1)"
        NEED_BREW=false
    else
        BREW_STATUS="${RED}[MISSING]${NC} -> Will install via https://brew.sh"
        NEED_BREW=true
    fi
    echo -e "Homebrew: $BREW_STATUS"

    # Git
    if command -v git &> /dev/null; then
        GIT_STATUS="${GREEN}[FOUND]${NC} $(git --version)"
        NEED_GIT=false
    else
        GIT_STATUS="${RED}[MISSING]${NC} -> Will install via Homebrew"
        NEED_GIT=true
    fi
    echo -e "Git: $GIT_STATUS"

    # Go
    if command -v go &> /dev/null; then
        CURRENT_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        # Simple version comparison
        if [ "$(printf '%s\n' "$REQUIRED_GO_VERSION" "$CURRENT_GO_VERSION" | sort -V | head -n1)" = "$REQUIRED_GO_VERSION" ]; then
            GO_STATUS="${GREEN}[FOUND]${NC} v$CURRENT_GO_VERSION"
            NEED_GO=false
        else
            GO_STATUS="${YELLOW}[OUTDATED]${NC} v$CURRENT_GO_VERSION -> Will upgrade to v$REQUIRED_GO_VERSION+ via Homebrew"
            NEED_GO=true
        fi
    else
        GO_STATUS="${RED}[MISSING]${NC} -> Will install v$REQUIRED_GO_VERSION+ via Homebrew"
        NEED_GO=true
    fi
    echo -e "Go: $GO_STATUS"

    # Podman
    if command -v podman &> /dev/null; then
        PODMAN_VERSION=$(podman --version)
        if podman machine list --format "{{.Running}}" | grep -q "true"; then
            PODMAN_STATUS="${GREEN}[FOUND & RUNNING]${NC} $PODMAN_VERSION"
            NEED_PODMAN=false
            NEED_PODMAN_START=false
        else
            PODMAN_STATUS="${YELLOW}[FOUND BUT STOPPED]${NC} $PODMAN_VERSION -> Will start podman machine"
            NEED_PODMAN=false
            NEED_PODMAN_START=true
        fi
    else
        PODMAN_STATUS="${RED}[MISSING]${NC} -> Will install via Homebrew"
        NEED_PODMAN=true
        NEED_PODMAN_START=true
    fi
    echo -e "Podman: $PODMAN_STATUS"

    # auth.json
    if [ -f "auth.json" ]; then
        AUTH_STATUS="${GREEN}[FOUND]${NC} auth.json exists"
        NEED_AUTH=false
    else
        AUTH_STATUS="${RED}[MISSING]${NC} -> Will create template"
        NEED_AUTH=true
    fi
    echo -e "auth.json: $AUTH_STATUS"
}

echo -e "\n${BLUE}--- Discovery Phase ---${NC}"
discover_dependencies

# 3. Confirmation Phase
if [ "$NEED_BREW" = false ] && [ "$NEED_GIT" = false ] && [ "$NEED_GO" = false ] && [ "$NEED_PODMAN" = false ] && [ "$NEED_PODMAN_START" = false ] && [ "$NEED_AUTH" = false ]; then
    echo -e "\n${GREEN}Everything is already set up!${NC}"
else
    echo -e "\n${YELLOW}Action required to complete setup.${NC}"
    read -p "Do you want to proceed with these changes? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${RED}Setup cancelled.${NC}"
        exit 1
    fi
fi

# 4. Installation Phase
echo -e "\n${BLUE}--- Installation Phase ---${NC}"

if [ "$NEED_BREW" = true ]; then
    echo "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    # Add brew to path for the rest of the script if it was just installed
    eval "$(/opt/homebrew/bin/brew shellenv)"
fi

if [ "$NEED_GIT" = true ]; then
    echo "Installing Git via Homebrew..."
    brew install git
fi

if [ "$NEED_GO" = true ]; then
    echo "Installing/Upgrading Go via Homebrew..."
    brew install go
fi

if [ "$NEED_PODMAN" = true ]; then
    echo "Installing Podman via Homebrew..."
    brew install podman
    echo "Initializing Podman machine..."
    podman machine init --cpus 4 --memory 8192 --disk-size 40 || echo "Podman machine already exists"
    NEED_PODMAN_START=true
fi

if [ "$NEED_PODMAN_START" = true ]; then
    echo "Starting Podman machine..."
    podman machine start || echo "Podman machine already started"
fi

if [ "$NEED_AUTH" = true ]; then
    echo "Creating auth.json template..."
    cat > auth.json <<EOF
{
  "google": {
    "type": "api",
    "key": "YOUR_API_KEY_HERE"
  }
}
EOF
    echo -e "${YELLOW}Warning: auth.json created with placeholder. Please add your Google AI API key.${NC}"
fi

# 5. Re-discovery Phase
echo -e "\n${BLUE}--- Verification Phase ---${NC}"
discover_dependencies

# Final Summary
echo -e "\n${BLUE}--- Post-Install Steps ---${NC}"

# Check PATH for Go
if ! grep -q "GOPATH" ~/.zshrc 2>/dev/null; then
    echo -e "1. ${YELLOW}Recommended:${NC} Add Go bin to your PATH in ~/.zshrc:"
    echo -e "   ${BLUE}echo 'export PATH=\"\$(go env GOPATH)/bin:\$PATH\"' >> ~/.zshrc${NC}"
fi

if [ "$NEED_AUTH" = true ]; then
    echo -e "2. ${YELLOW}Action Required:${NC} Edit ${BLUE}auth.json${NC} and replace ${BLUE}YOUR_API_KEY_HERE${NC} with a valid Google AI API key."
fi

echo -e "\n${GREEN}Setup script finished!${NC}"
echo -e "To start Overdrive:"
echo -e "   1. ${BLUE}./scripts/rebuild-restart-all${NC}"
echo -e "   2. Open your browser to: ${BLUE}http://localhost:$PORT${NC}"
