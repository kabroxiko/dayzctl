#!/bin/bash
set -euo pipefail

# ============================================================================
# Configuration
# ============================================================================
DAYZ_HOME="${DAYZ_HOME:-/srv/dayz}"
DAYZCTL_VERSION="${DAYZCTL_VERSION:-v1.0.0}"
STEAM_USER="${STEAM_USER:-kqkklan}"
REINSTALL="${REINSTALL:-0}"

# ============================================================================
# Colors for output
# ============================================================================
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[install]${NC} $*"; }
warn() { echo -e "${YELLOW}[install] WARNING:${NC} $*" >&2; }
error() { echo -e "${RED}[install] ERROR:${NC} $*" >&2; exit 1; }

# ============================================================================
# Check root privileges
# ============================================================================
if [ "$EUID" -ne 0 ]; then
    error "Please run as root"
fi

# ============================================================================
# Detect distribution
# ============================================================================
detect_distro() {
    if [ ! -f /etc/os-release ]; then
        error "Cannot detect distribution: /etc/os-release not found"
    fi
    
    . /etc/os-release
    
    DISTRO_ID="${ID:-unknown}"
    
    if [ -n "${ID_LIKE:-}" ]; then
        DISTRO_FAMILY="$ID_LIKE"
    else
        case "$ID" in
            ubuntu|debian) DISTRO_FAMILY="apt" ;;
            centos|rhel|fedora) DISTRO_FAMILY="yum" ;;
            alpine) DISTRO_FAMILY="apk" ;;
            *) DISTRO_FAMILY="unknown" ;;
        esac
    fi
    
    if [ -z "$DISTRO_FAMILY" ] || [ "$DISTRO_FAMILY" = "unknown" ]; then
        case "$ID" in
            ubuntu|debian) DISTRO_FAMILY="apt" ;;
            centos|rhel|fedora) DISTRO_FAMILY="yum" ;;
            alpine) DISTRO_FAMILY="apk" ;;
            *) error "Unsupported distribution: $ID (family: $DISTRO_FAMILY)" ;;
        esac
    fi
    
    log "detected distro: $ID (family $DISTRO_FAMILY)"
}

# ============================================================================
# Create directory structure
# ============================================================================
create_structure() {
    log "creating structure in $DAYZ_HOME"
    mkdir -p "$DAYZ_HOME" \
             "$DAYZ_HOME/config" \
             "$DAYZ_HOME/server" \
             "$DAYZ_HOME/backups" \
             "$DAYZ_HOME/workshop" \
             "$DAYZ_HOME/state" \
             "$DAYZ_HOME/steamcmd" || error "Failed to create directory structure"
}

# ============================================================================
# Create dayz user (runs the actual server)
# ============================================================================
create_user() {
    if id "dayz" &>/dev/null; then
        log "user dayz already exists"
        return 0
    fi
    
    log "creating user dayz"
    useradd -m -d "$DAYZ_HOME" -s /bin/bash dayz || error "Failed to create dayz user"
    chown -R dayz:dayz "$DAYZ_HOME" || error "Failed to set ownership for dayz user"
}

# ============================================================================
# Install system dependencies (only what's needed)
# ============================================================================
install_deps() {
    log "enabling i386 architecture and installing dependencies ($DISTRO_FAMILY)"
    
    case "$DISTRO_FAMILY" in
        apt|debian)
            dpkg --add-architecture i386 || error "Failed to add i386 architecture"
            apt-get update -qq || error "Failed to update package lists"
            apt-get install -y -qq \
                curl \
                tar \
                gzip \
                rsync \
                ca-certificates \
                lib32gcc-s1 \
                util-linux || error "Failed to install dependencies"
            ;;
        yum|fedora|rhel|centos)
            yum install -y -q \
                curl \
                tar \
                gzip \
                rsync \
                ca-certificates \
                glibc.i686 \
                util-linux || error "Failed to install dependencies"
            ;;
        apk|alpine)
            apk add --no-cache \
                curl \
                tar \
                gzip \
                rsync \
                ca-certificates \
                libgcc \
                util-linux || error "Failed to install dependencies"
            ;;
        *)
            error "Unsupported package manager: $DISTRO_FAMILY"
            ;;
    esac
}

# ============================================================================
# Install SteamCMD (runs as dayz user)
# ============================================================================
install_steamcmd() {
    log "installing SteamCMD in $DAYZ_HOME/steamcmd"
    
    if [ -f "$DAYZ_HOME/steamcmd/steamcmd.sh" ] && [ "$REINSTALL" != "1" ]; then
        log "SteamCMD already installed"
        return 0
    fi
    
    mkdir -p "$DAYZ_HOME/steamcmd" || error "Failed to create steamcmd directory"
    chown dayz:dayz "$DAYZ_HOME/steamcmd" || error "Failed to set ownership on steamcmd directory"
    
    cd "$DAYZ_HOME/steamcmd" || error "Failed to cd to steamcmd directory"
    
    runuser -u dayz -- sh -c "curl -sSL https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz | tar -xz" \
        || error "Failed to download/extract SteamCMD"
    
    runuser -u dayz -- "$DAYZ_HOME/steamcmd/steamcmd.sh" +quit \
        || error "SteamCMD installation test failed"
}

# ============================================================================
# Install dayzctl binary (runs as root only)
# ============================================================================
install_dayzctl() {
    log "installing dayzctl binary"
    
    ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
    if [ -z "$ARCH" ]; then
        error "Failed to detect architecture"
    fi
    
    rm -f /usr/local/bin/dayzctl || warn "Failed to remove old binary (may not exist)"
    
    # Try local build first
    LOCAL_BINARY="./build/dayzctl-linux-${ARCH}"
    if [ -f "$LOCAL_BINARY" ]; then
        log "found local binary: $LOCAL_BINARY"
        cp "$LOCAL_BINARY" /usr/local/bin/dayzctl || error "Failed to copy local binary"
    elif [ -f "./build/dayzctl" ]; then
        log "found local binary: ./build/dayzctl"
        cp ./build/dayzctl /usr/local/bin/dayzctl || error "Failed to copy local binary"
    else
        # Download from GitHub releases
        log "downloading dayzctl-linux-${ARCH} version ${DAYZCTL_VERSION}"
        if ! curl -fsSL -o /usr/local/bin/dayzctl \
            "https://github.com/kabroxiko/dayzctl/releases/download/${DAYZCTL_VERSION}/dayzctl-linux-${ARCH}"; then
            error "Failed to download dayzctl binary from GitHub"
        fi
    fi
    
    chmod 755 /usr/local/bin/dayzctl || error "Failed to make dayzctl executable"
    
    if ! /usr/local/bin/dayzctl version &>/dev/null; then
        error "dayzctl binary verification failed"
    fi
    
    log "dayzctl installed successfully: $(/usr/local/bin/dayzctl version)"
}

# ============================================================================
# Install bercon-cli (RCON client)
# ============================================================================
install_bercon_cli() {
    log "installing bercon-cli"
    
    ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
    if [ -z "$ARCH" ]; then
        warn "Failed to detect architecture, skipping bercon-cli install"
        return 0
    fi
    
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    if [ "$OS" != "linux" ]; then
        warn "bercon-cli only available for Linux, skipping"
        return 0
    fi
    
    # Try latest release first
    BINARY_URL="https://github.com/WoozyMasta/bercon-cli/releases/latest/download/bercon-cli-${OS}-${ARCH}"
    log "downloading bercon-cli from $BINARY_URL"
    
    if ! curl -fsSL -o /usr/bin/bercon-cli "$BINARY_URL"; then
        # Fallback to specific version
        warn "Failed to download latest, trying v0.4.4..."
        BINARY_URL="https://github.com/WoozyMasta/bercon-cli/releases/download/v0.4.4/bercon-cli-${OS}-${ARCH}"
        if ! curl -fsSL -o /usr/bin/bercon-cli "$BINARY_URL"; then
            warn "Failed to download bercon-cli, RCON will not be available"
            return 1
        fi
    fi
    
    chmod 755 /usr/bin/bercon-cli || warn "Failed to make bercon-cli executable"
    
    # Verify bercon-cli works
    if /usr/bin/bercon-cli -v &>/dev/null; then
        log "bercon-cli installed successfully: $(/usr/bin/bercon-cli -v 2>&1 | head -1)"
    else
        warn "bercon-cli installed but verification failed"
    fi
}

# ============================================================================
# Create default server.yaml config
# ============================================================================
create_config() {
    if [ -f "$DAYZ_HOME/config/server.yaml" ] && [ "$REINSTALL" != "1" ]; then
        log "config already exists in $DAYZ_HOME/config/server.yaml (preserved)"
        return 0
    fi
    
    log "creating default config"
    
    cat > "$DAYZ_HOME/config/server.yaml" << EOF || error "Failed to write config file"
# DayZ Server Configuration
# ============================================================================
# This config is managed by dayzctl (root tool)
# The dayz user only runs steamcmd, not dayzctl
# ============================================================================

# Steam credentials (used by dayz user via steamcmd)
steam:
  username: "$STEAM_USER"

# Directory paths
paths:
  install_dir: "$DAYZ_HOME/server"
  workshop_dir: "$DAYZ_HOME/workshop"
  mods_dir: "$DAYZ_HOME/server"
  backups_dir: "$DAYZ_HOME/backups"
  state_dir: "$DAYZ_HOME/state"
  storage_dir: null
  steamcmd_bin: "$DAYZ_HOME/steamcmd/steamcmd.sh"

# Server instances (managed by root via systemd)
instances:
  - name: main
    instanceId: 1
    port: 2302
    steam_query_port: 27016
    template: dayzOffline.chernarusplus
    hostname: "DayZ Server"
    max_players: 60
    enabled: true
    rcon:
      enabled: true
      port: 2306
      password: rcon123
    mods: []
    servermods: []

  - name: solo
    instanceId: 2
    port: 2303
    steam_query_port: 27017
    template: dayzOffline.chernarusplus
    hostname: "DayZ Solo Server"
    max_players: 30
    enabled: true
    rcon:
      enabled: true
      port: 2307
      password: rcon123
    mods: []
    servermods: []

# Backup settings (managed by root)
backup:
  enabled: true
  retention_days: 14
  keep_daily: 7
  keep_weekly: 4

# Update settings (root runs dayzctl update, which calls steamcmd as dayz)
updates:
  enabled: true
  schedule: "04:00"

# Healthcheck settings
healthcheck:
  enabled: true
  startup_timeout: 300
EOF
    
    chown -R dayz:dayz "$DAYZ_HOME/config" || error "Failed to set ownership on config"
    log "config created at $DAYZ_HOME/config/server.yaml"
}

# ============================================================================
# Set ownership of all files
# ============================================================================
set_ownership() {
    log "adjusting ownership of $DAYZ_HOME to dayz:dayz"
    chown -R dayz:dayz "$DAYZ_HOME" 2>/dev/null || warn "Failed to set ownership on some files (continuing)"
}

# ============================================================================
# Main installation
# ============================================================================
main() {
    log "=== DayZ Server Installation ==="
    log "dayzctl: root tool for server management"
    log "dayz user: runs steamcmd and server processes"
    log ""
    
    detect_distro
    create_structure
    create_user
    install_deps
    install_steamcmd
    install_dayzctl
    install_bercon_cli
    create_config
    set_ownership
    
    log "Applying configuration..."
    if ! /usr/local/bin/dayzctl apply; then
        error "dayzctl apply failed - check the error above"
    fi
    log "Configuration applied successfully"
    
    log ""
    log "=== Installation Complete ==="
    log ""
    log "📁 Configuration: $DAYZ_HOME/config/server.yaml"
    log ""
    log "🔧 dayzctl commands (run as root):"
    log "  dayzctl apply              # Generate systemd units (does NOT restart services)"
    log "  dayzctl status             # Check server status"
    log "  dayzctl update             # Update server (calls steamcmd as dayz)"
    log "  dayzctl instance start main  # Start main instance"
    log "  dayzctl instance stop main   # Stop main instance"
    log "  dayzctl mods list          # List mods"
    log "  dayzctl rcon send main status  # Send RCON command"
    log ""
    log "🔑 Steam login (run as dayz user):"
    log "  sudo -u dayz dayzctl steam-login"
    log ""
    log "📋 View logs:"
    log "  journalctl -u dayz@main -f"
    log "  journalctl -u dayz-update -f"
    log ""
    log "done."
}

# ============================================================================
# Parse command line arguments
# ============================================================================
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            DAYZCTL_VERSION="$2"
            shift 2
            ;;
        --user)
            STEAM_USER="$2"
            shift 2
            ;;
        --home)
            DAYZ_HOME="$2"
            shift 2
            ;;
        --reinstall)
            REINSTALL=1
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --version VERSION  Specify version to install (default: v1.0.0)"
            echo "  --user USER        Steam username to use"
            echo "  --home PATH        Installation directory (default: /srv/dayz)"
            echo "  --reinstall        Force reinstall even if files exist"
            echo "  --help, -h         Show this help"
            echo ""
            echo "Architecture:"
            echo "  dayzctl runs as root (system management)"
            echo "  dayz user runs steamcmd and server processes"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

main "$@"
