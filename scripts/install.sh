#!/bin/bash

# GitPatrol - Tactical Asset Monitoring
# Unified Installation Script (Standard Linux / Proxmox LXC)
# Standard: Binary-First Installation (Fast & Lightweight)

set -e

# --- Configuration ---
INSTALL_DIR="/opt/gitpatrol"
DATA_DIR="/var/lib/gitpatrol"
DB_DIR="$DATA_DIR/db"
BACKEND_PORT=8080
REPO_OWNER="efinityhub"
REPO_NAME="gitpatrol"
GITHUB_REPO="https://github.com/$REPO_OWNER/$REPO_NAME"

# Efinity Branding
echo "⚡ GitPatrol | Tactical Asset Monitoring"
echo "----------------------------------------"

# --- Root Check ---
if [ "$EUID" -ne 0 ]; then
  echo "Error: Please run as root (sudo)."
  exit 1
fi

# --- Architecture Check ---
ARCH=$(uname -m)
case $ARCH in
    x86_64) BIN_ARCH="amd64" ;;
    aarch64) BIN_ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# --- Dependency Check ---
echo "📦 Checking dependencies..."
DEPS="curl git wget jq"
apt-get update -qq
apt-get install -y -qq $DEPS

# --- Directory Setup ---
echo "📂 Setting up directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$DATA_DIR"
mkdir -p "$DB_DIR"

# --- Binary Installation (Preferred) ---
echo "🚀 Fetching latest release info..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | jq -r .tag_name)

if [ "$LATEST_RELEASE" != "null" ]; then
    echo "Found latest version: $LATEST_RELEASE"
    BINARY_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_RELEASE/gitpatrol-linux-$BIN_ARCH"
    
    echo "📥 Downloading pre-built binary..."
    if wget -q --spider "$BINARY_URL"; then
        wget -q -O "$INSTALL_DIR/gitpatrol" "$BINARY_URL"
        chmod +x "$INSTALL_DIR/gitpatrol"
        echo "✅ Binary installed successfully."
    else
        echo "⚠️  No pre-built binary found for $BIN_ARCH at $LATEST_RELEASE. Falling back to source build..."
        BUILD_FROM_SOURCE=true
    fi
else
    echo "⚠️  No GitHub releases found. Falling back to source build (main branch)..."
    BUILD_FROM_SOURCE=true
fi

# --- Source Build Fallback (Only if needed) ---
if [ "$BUILD_FROM_SOURCE" = true ]; then
    echo "🏗️  Starting source build (this will take longer)..."
    
    # Install build-only dependencies
    apt-get install -y -qq build-essential
    
    # Go (if missing)
    if ! command -v go &> /dev/null; then
        GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -n 1)
        wget -q "https://go.dev/dl/${GO_VERSION}.linux-amd64.tar.gz"
        tar -C /usr/local -xzf "${GO_VERSION}.linux-amd64.tar.gz"
        export PATH=$PATH:/usr/local/go/bin
        rm "${GO_VERSION}.linux-amd64.tar.gz"
        WAS_GO_INSTALLED=true
    fi

    # Node.js (if missing)
    if ! command -v npm &> /dev/null; then
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y -qq nodejs
        WAS_NODE_INSTALLED=true
    fi

    # Clone & Build
    if [ ! -d "$INSTALL_DIR/.git" ]; then
        git clone "$GITHUB_REPO" "$INSTALL_DIR/source"
    else
        cd "$INSTALL_DIR/source" && git pull origin main
    fi

    cd "$INSTALL_DIR/source/frontend"
    npm install --silent && npm run build --silent
    
    cd "$INSTALL_DIR/source/backend"
    mkdir -p cmd/gitpatrol/build
    cp -r ../frontend/build/* cmd/gitpatrol/build/
    go build -o "$INSTALL_DIR/gitpatrol" ./cmd/gitpatrol/main.go
    
    # --- Cleanup Build Deps (Surgical) ---
    if [ "$WAS_NODE_INSTALLED" = true ]; then
        echo "🧹 Cleaning up Node.js..."
        apt-get remove -y -qq nodejs &> /dev/null
    fi
    # We keep Go if it was manually installed by script because it's in /usr/local/go (not apt)
    # but we could remove the source folder
    rm -rf "$INSTALL_DIR/source"
fi

# --- Environment & DB ---
if [ ! -f "$DB_DIR/gitpatrol.env" ]; then
    echo "⚙️  Creating default configuration..."
    cat > "$DB_DIR/gitpatrol.env" <<EOF
DB_PATH=$DB_DIR/gitpatrol.db
WORKER_COUNT=3
PORT=$BACKEND_PORT
EOF
fi

# --- Systemd Integration ---
echo "🔧 Registering Systemd service..."
cat > /etc/systemd/system/gitpatrol.service <<EOF
[Unit]
Description=GitPatrol - Tactical Asset Monitoring
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/gitpatrol
Restart=always
RestartSec=5
EnvironmentFile=$DB_DIR/gitpatrol.env

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable gitpatrol &> /dev/null
systemctl restart gitpatrol

# --- Finalization ---
echo "----------------------------------------"
echo "✅ Installation Complete!"
echo "📍 Access GitPatrol at: http://$(hostname -I | awk '{print $1}'):$BACKEND_PORT"
echo "📊 Monitoring tactical assets in the shadows..."
