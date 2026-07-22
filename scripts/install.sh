#!/bin/bash
set -euo pipefail

# Debugging helper — print debug lines as normal colored install messages
debug_log() {
    log "[debug] $*"
}

# ============================================================================
# Configuration
# ============================================================================
DAYZ_HOME="${DAYZ_HOME:-/srv/dayz}"
STEAM_USER="${STEAM_USER:-kqkklan}"
REINSTALL="${REINSTALL:-0}"
DEFAULT_TEMPLATE_URL="https://raw.githubusercontent.com/kabroxiko/dayzops/main/scripts/server.yaml.tmpl"

# Prompt interactively for configuration values when running in a terminal
prompt_for_values() {
    if [ ! -t 0 ]; then
        debug_log "Non-interactive shell; skipping interactive prompts"
        return 0
    fi

    log "Interactive setup — press Enter to accept the default in brackets"
    input=""
    read -r -p "Installation directory (DAYZ_HOME) [${DAYZ_HOME}]: " input
    if [ -n "${input}" ]; then
        DAYZ_HOME="$input"
    fi

    input=""
    read -r -p "Steam username (STEAM_USER) [${STEAM_USER}]: " input
    if [ -n "${input}" ]; then
        STEAM_USER="$input"
    fi

}
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
    debug_log "mkdir -p $DAYZ_HOME {config,server,backups,workshop,state,steamcmd}"
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

    debug_log "running as dayz: curl -sSL https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz | tar -xz"
    runuser -u dayz -- sh -c "curl -sSL https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz | tar -xz" \
        || error "Failed to download/extract SteamCMD"

    debug_log "running as dayz: $DAYZ_HOME/steamcmd/steamcmd.sh +quit"
    runuser -u dayz -- "$DAYZ_HOME/steamcmd/steamcmd.sh" +quit \
        || error "SteamCMD installation test failed"
}

# ============================================================================
# Install dayzctl binary (runs as root only)
# ============================================================================
install_dayzctl() {
    log "installing dayzctl binary (latest release)"
    
    ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    
    if [ -z "$ARCH" ]; then
        error "Failed to detect architecture"
    fi
    
    rm -f /usr/local/bin/dayzctl || warn "Failed to remove old binary (may not exist)"
    
    # Try local build first
    LOCAL_BINARY="./build/dayzctl-${OS}-${ARCH}"
    if [ -f "$LOCAL_BINARY" ]; then
        log "found local binary: $LOCAL_BINARY"
        cp "$LOCAL_BINARY" /usr/local/bin/dayzctl || error "Failed to copy local binary"
        chmod 755 /usr/local/bin/dayzctl || error "Failed to make dayzctl executable"
        log "dayzctl installed successfully from local build"
        return 0
    elif [ -f "./build/dayzctl" ]; then
        log "found local binary: ./build/dayzctl"
        cp ./build/dayzctl /usr/local/bin/dayzctl || error "Failed to copy local binary"
        chmod 755 /usr/local/bin/dayzctl || error "Failed to make dayzctl executable"
        log "dayzctl installed successfully from local build"
        return 0
    fi
    
    # Download manifest.json (SBOM) from latest release
    log "Downloading manifest from GitHub..."
    MANIFEST_URL="https://github.com/kabroxiko/dayzctl/releases/latest/download/manifest.json"
    
    if command -v curl >/dev/null 2>&1; then
        MANIFEST=$(curl -fsSL "$MANIFEST_URL" 2>/dev/null || echo "")
    elif command -v wget >/dev/null 2>&1; then
        MANIFEST=$(wget -qO- "$MANIFEST_URL" 2>/dev/null || echo "")
    else
        error "Neither curl nor wget available"
    fi
    
    if [ -z "$MANIFEST" ]; then
        error "Failed to download manifest from $MANIFEST_URL"
    fi
    
    # Extract version from SPDX manifest
    VERSION=$(echo "$MANIFEST" | grep -o '"versionInfo":"[^"]*"' | head -1 | sed 's/"versionInfo":"\([^"]*\)"/\1/')
    
    # Find the asset for our OS/ARCH from SPDX packages
    ASSET_NAME="dayzctl-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        ASSET_NAME="${ASSET_NAME}.exe"
    fi
    
    # Verify the asset exists in the manifest
    if ! echo "$MANIFEST" | grep -q "\"name\":\"${ASSET_NAME}\""; then
        error "Asset ${ASSET_NAME} not found in manifest"
    fi
    
    # Get checksum from manifest for verification
    CHECKSUM=$(echo "$MANIFEST" | grep -A 10 "\"name\":\"${ASSET_NAME}\"" | grep -o '"checksumValue":"[^"]*"' | head -1 | sed 's/"checksumValue":"\([^"]*\)"/\1/')
    
    if [ -z "$VERSION" ]; then
        error "Failed to parse version from manifest"
    fi
    
    log "Latest version: $VERSION"
    log "Asset: $ASSET_NAME"
    [ -n "$CHECKSUM" ] && log "Expected checksum: $CHECKSUM"
    
    # Download the binary
    DL_URL="https://github.com/kabroxiko/dayzctl/releases/download/v${VERSION}/${ASSET_NAME}"
    log "Downloading from: $DL_URL"
    
    # Download to temp file for checksum verification
    TMP_FILE=$(mktemp)
    
    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL -o "$TMP_FILE" "$DL_URL"; then
            rm -f "$TMP_FILE"
            error "Failed to download dayzctl binary from $DL_URL"
        fi
    else
        if ! wget -qO "$TMP_FILE" "$DL_URL"; then
            rm -f "$TMP_FILE"
            error "Failed to download dayzctl binary from $DL_URL"
        fi
    fi
    
    # Verify checksum if available
    if [ -n "$CHECKSUM" ]; then
        log "Verifying checksum..."
        ACTUAL_CHECKSUM=$(sha256sum "$TMP_FILE" | cut -d' ' -f1)
        if [ "$ACTUAL_CHECKSUM" != "$CHECKSUM" ]; then
            rm -f "$TMP_FILE"
            error "Checksum verification failed. Expected: $CHECKSUM, Got: $ACTUAL_CHECKSUM"
        fi
        log "Checksum verified successfully"
    fi
    
    # Move to final location
    mv "$TMP_FILE" /usr/local/bin/dayzctl || error "Failed to move binary"
    chmod 755 /usr/local/bin/dayzctl || error "Failed to make dayzctl executable"
    
    if ! /usr/local/bin/dayzctl version &>/dev/null; then
        error "dayzctl binary verification failed"
    fi
    
    log "dayzctl v${VERSION} installed successfully: $(/usr/local/bin/dayzctl version)"
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
# Create default server.yaml config (always download template)
# ============================================================================
create_config() {
    if [ -f "$DAYZ_HOME/config/server.yaml" ] && [ "$REINSTALL" != "1" ]; then
        log "config already exists in $DAYZ_HOME/config/server.yaml (preserved)"
        return 0
    fi

    log "creating default config"

    log "Downloading default template"
    TMP_TEMPLATE="$(mktemp /tmp/server.yaml.tmpl.XXXXXX)" || error "Failed to create temp file for template"

    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL "$DEFAULT_TEMPLATE_URL" -o "$TMP_TEMPLATE"; then
            rm -f "$TMP_TEMPLATE"
            error "Failed to download template"
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -qO "$TMP_TEMPLATE" "$DEFAULT_TEMPLATE_URL"; then
            rm -f "$TMP_TEMPLATE"
            error "Failed to download template"
        fi
    else
        rm -f "$TMP_TEMPLATE"
        error "Neither curl nor wget available to download template"
    fi

    debug_log "Downloaded template to $TMP_TEMPLATE"

    sed "s|%%DAYZ_HOME%%|$DAYZ_HOME|g; s|%%STEAM_USER%%|$STEAM_USER|g" "$TMP_TEMPLATE" > "$DAYZ_HOME/config/server.yaml" || {
        rm -f "$TMP_TEMPLATE"
        error "Failed to render config template"
    }

    rm -f "$TMP_TEMPLATE" || warn "Failed to remove temporary template"

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
    # Ask interactively for configuration when running in a TTY
    prompt_for_values

    detect_distro
    debug_log "DAYZ_HOME=$DAYZ_HOME STEAM_USER=$STEAM_USER REINSTALL=$REINSTALL"
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
    
    log "Downloading/updating DayZ server (this may take a while)..."
    if ! runuser -u dayz -- /usr/local/bin/dayzctl update; then
        error "dayzctl update failed - check the error above"
    fi
    log "DayZ server downloaded/updated successfully"
    
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
    log "  sudo -u dayz /usr/local/bin/dayzctl steam-login"
    log ""
    log "📋 View logs:"
    log "  journalctl -u dayz@main -f"
    log "  journalctl -u dayz-update -f"
    log ""
    log "done."
}

main "$@"
