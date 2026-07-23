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