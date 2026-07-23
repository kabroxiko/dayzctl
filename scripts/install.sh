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
    
    # Get version from checksums.txt filename pattern
    # Since checksums.txt is in the release, we can get the version from the tag
    LATEST_URL="https://github.com/kabroxiko/dayzctl/releases/latest"
    REDIRECT_URL=$(curl -fsSL -o /dev/null -w "%{url_effective}" "$LATEST_URL")
    VERSION=$(echo "$REDIRECT_URL" | grep -o 'v[0-9.]*$' | sed 's/^v//')
    
    if [ -z "$VERSION" ]; then
        error "Failed to get latest version from GitHub"
    fi
    
    ASSET="dayzctl_${VERSION}_${OS}_${ARCH}.tar.gz"
    DL_URL="https://github.com/kabroxiko/dayzctl/releases/download/v${VERSION}/${ASSET}"
    CHECKSUM_URL="https://github.com/kabroxiko/dayzctl/releases/download/v${VERSION}/checksums.txt"
    
    log "Installing dayzctl v${VERSION}"
    log "Downloading ${ASSET}..."
    
    # Download checksums
    CHECKSUMS=$(curl -fsSL "$CHECKSUM_URL")
    
    if [ -z "$CHECKSUMS" ]; then
        error "Failed to download checksums"
    fi
    
    # Get checksum for our asset
    EXPECTED_CHECKSUM=$(echo "$CHECKSUMS" | grep " ${ASSET}$" | awk '{print $1}')
    
    if [ -z "$EXPECTED_CHECKSUM" ]; then
        error "Checksum not found for ${ASSET}"
    fi
    
    # Download archive
    TMP_DIR=$(mktemp -d)
    TMP_FILE="${TMP_DIR}/${ASSET}"
    
    curl -fsSL -o "$TMP_FILE" "$DL_URL"
    
    if [ ! -f "$TMP_FILE" ]; then
        rm -rf "$TMP_DIR"
        error "Failed to download ${ASSET}"
    fi
    
    # Verify checksum
    ACTUAL_CHECKSUM=$(sha256sum "$TMP_FILE" | awk '{print $1}')
    
    if [ "$ACTUAL_CHECKSUM" != "$EXPECTED_CHECKSUM" ]; then
        rm -rf "$TMP_DIR"
        error "Checksum verification failed"
    fi
    
    # Extract and install
    tar -xzf "$TMP_FILE" -C "$TMP_DIR"
    mv "${TMP_DIR}/dayzctl" /usr/local/bin/dayzctl
    chmod 755 /usr/local/bin/dayzctl
    
    rm -rf "$TMP_DIR"
    
    # Verify installation
    if ! /usr/local/bin/dayzctl version > /dev/null 2>&1; then
        error "dayzctl verification failed"
    fi
    
    log "dayzctl v${VERSION} installed successfully"
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

    # Try multiple template URLs
    TEMPLATE_URLS=(
        "https://raw.githubusercontent.com/kabroxiko/dayzctl/main/scripts/server.yaml.tmpl"
        "https://raw.githubusercontent.com/kabroxiko/dayzops/main/scripts/server.yaml.tmpl"
    )

    TMP_TEMPLATE=""
    for url in "${TEMPLATE_URLS[@]}"; do
        log "Trying template URL: $url"
        TMP_TEMPLATE=$(mktemp /tmp/server.yaml.tmpl.XXXXXX 2>/dev/null || echo "")
        
        if [ -z "$TMP_TEMPLATE" ]; then
            log "Failed to create temp file, trying with mktemp without suffix..."
            TMP_TEMPLATE=$(mktemp)
        fi
        
        if [ -z "$TMP_TEMPLATE" ]; then
            error "Failed to create temporary file"
        fi

        if command -v curl >/dev/null 2>&1; then
            log "Using curl to download template..."
            if curl -fsSL "$url" -o "$TMP_TEMPLATE" 2>/tmp/curl_error.log; then
                log "Template downloaded successfully from $url"
                break
            else
                log "Failed to download from $url (curl error: $(cat /tmp/curl_error.log 2>/dev/null || echo 'unknown'))"
                rm -f "$TMP_TEMPLATE"
                TMP_TEMPLATE=""
            fi
        elif command -v wget >/dev/null 2>&1; then
            log "Using wget to download template..."
            if wget -qO "$TMP_TEMPLATE" "$url" 2>/tmp/wget_error.log; then
                log "Template downloaded successfully from $url"
                break
            else
                log "Failed to download from $url (wget error: $(cat /tmp/wget_error.log 2>/dev/null || echo 'unknown'))"
                rm -f "$TMP_TEMPLATE"
                TMP_TEMPLATE=""
            fi
        else
            error "Neither curl nor wget available to download template"
        fi
    done

    if [ -z "$TMP_TEMPLATE" ] || [ ! -f "$TMP_TEMPLATE" ]; then
        error "Failed to download template from all sources. Please check network connectivity and URLs."
    fi

    log "Template downloaded successfully, rendering config..."
    debug_log "Template file: $TMP_TEMPLATE"
    debug_log "DAYZ_HOME: $DAYZ_HOME"
    debug_log "STEAM_USER: $STEAM_USER"

    sed "s|%%DAYZ_HOME%%|$DAYZ_HOME|g; s|%%STEAM_USER%%|$STEAM_USER|g" "$TMP_TEMPLATE" > "$DAYZ_HOME/config/server.yaml" || {
        rm -f "$TMP_TEMPLATE"
        error "Failed to render config template"
    }

    rm -f "$TMP_TEMPLATE" || warn "Failed to remove temporary template"

    chown -R dayz:dayz "$DAYZ_HOME/config" || error "Failed to set ownership on config"
    log "config created at $DAYZ_HOME/config/server.yaml"
    
    # Display config content for debugging
    debug_log "Config content:"
    debug_log "$(cat $DAYZ_HOME/config/server.yaml 2>/dev/null | head -20 || echo 'Unable to read config')"
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
    
    prompt_for_values

    detect_distro
    create_structure
    create_user
    install_deps
    install_steamcmd
    install_dayzctl
    create_config
    set_ownership
    
    log "Applying configuration..."
    if ! /usr/local/bin/dayzctl apply; then
        error "dayzctl apply failed - check the error above"
    fi
    log "Configuration applied successfully"
    
    log "Downloading/updating DayZ server (this may take a while)..."
    if ! /usr/local/bin/dayzctl update; then
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
