#!/bin/sh

# ==============================================================================
#  tadev Installer Script (v1.0.0)
#  Repository: https://github.com/rugved-danej/termux-android-development
# ==============================================================================

set -e

# 1. Verify and detect Termux environment
if [ -n "$PREFIX" ] && [ -d "$PREFIX" ] && echo "$PREFIX" | grep -q "com.termux"; then
    TARGET_PREFIX="$PREFIX"
elif [ -d "/data/data/com.termux/files/usr" ]; then
    TARGET_PREFIX="/data/data/com.termux/files/usr"
else
    printf "\033[31m✕ Error: tadev is engineered exclusively for Termux on Android.\033[0m\n"
    printf "Please run this installer inside the Termux app on any supported architecture (ARM64, ARM, x86_64, 386).\n"
    exit 1
fi

if [ ! -d "$TARGET_PREFIX/bin" ]; then
    mkdir -p "$TARGET_PREFIX/bin" 2>/dev/null || {
        printf "\033[31m✕ Error: Unable to access or create %s/bin. Please run with appropriate permissions.\033[0m\n" "$TARGET_PREFIX"
        exit 1
    }
fi

SCRIPT_NAME="tadev"
ALT_NAME="termux-android-dev"
DESTINATION="$TARGET_PREFIX/bin/$SCRIPT_NAME"
ALT_DESTINATION="$TARGET_PREFIX/bin/$ALT_NAME"
REPO_OWNER="rugved-danej"
REPO_NAME="termux-android-development"
TAG="v1.0.0"

printf "\033[35m•\033[0m Installing %s 1.0.0...\n" "$SCRIPT_NAME"

# 2. Clean up any existing binary and symlinks before installing to prevent "same file" errors
rm -f "$DESTINATION" "$ALT_DESTINATION"

# 3. Detect device CPU architecture
ARCH=$(uname -m)
case "$ARCH" in
    aarch64|arm64) GOARCH="arm64" ;;
    armv7l|armv8l|arm) GOARCH="arm" ;;
    x86_64|amd64) GOARCH="amd64" ;;
    i686|386) GOARCH="386" ;;
    *)
        printf "\033[31m✕ Error: Unsupported architecture %s\033[0m\n" "$ARCH"
        exit 1
        ;;
esac

SCRIPT_DIR="$(cd "$(dirname "$0")" 2>/dev/null && pwd || echo "")"

# 4. Install binary (local precompiled dist, local Go build, or remote GitHub Release)
if [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/dist/tadev" ]; then
    printf "\033[35m•\033[0m Installing precompiled local binary from dist/tadev...\n"
    cp -f "$SCRIPT_DIR/dist/tadev" "$DESTINATION"
    chmod +x "$DESTINATION"
elif [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/main.go" ] && command -v go >/dev/null 2>&1; then
    printf "\033[35m•\033[0m Source files and Go detected. Building locally...\n"
    (cd "$SCRIPT_DIR" && go build -ldflags="-s -w -X 'main.Author=Rugved Danej' -X 'main.Version=1.0.0'" -o "$DESTINATION" main.go)
    chmod +x "$DESTINATION"
else
    printf "\033[35m•\033[0m Downloading standalone precompiled binary for %s (%s)...\n" "$GOARCH" "$ARCH"
    DOWNLOAD_URL_V100="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$TAG/tadev-android-$GOARCH"
    DOWNLOAD_URL_LATEST="https://github.com/$REPO_OWNER/$REPO_NAME/releases/latest/download/tadev-android-$GOARCH"
    
    DOWNLOAD_SUCCESS=false
    if curl -sL --fail "$DOWNLOAD_URL_V100" -o "$DESTINATION" || curl -sL --fail "$DOWNLOAD_URL_LATEST" -o "$DESTINATION"; then
        DOWNLOAD_SUCCESS=true
    elif command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
        printf "\033[35m•\033[0m Fetching release asset via GitHub CLI (authenticated)...\n"
        if gh release download "$TAG" -p "tadev-android-$GOARCH" -O "$DESTINATION" --repo "$REPO_OWNER/$REPO_NAME" 2>/dev/null; then
            DOWNLOAD_SUCCESS=true
        fi
    fi

    if [ "$DOWNLOAD_SUCCESS" != "true" ]; then
        printf "\033[33m! Note: Remote release asset not found. Trying local Go build...\033[0m\n"
        if command -v go >/dev/null 2>&1 && [ -n "$SCRIPT_DIR" ] && [ -f "$SCRIPT_DIR/main.go" ]; then
            (cd "$SCRIPT_DIR" && go build -ldflags="-s -w -X 'main.Author=Rugved Danej' -X 'main.Version=1.0.0'" -o "$DESTINATION" main.go)
        else
            printf "\033[31m✕ Error: Unable to download binary. Please install Go ('pkg install golang') or check repository visibility.\033[0m\n"
            exit 1
        fi
    fi
    chmod +x "$DESTINATION"
fi

# 5. Create canonical symlink safely: termux-android-dev -> tadev
rm -f "$ALT_DESTINATION"
(cd "$TARGET_PREFIX/bin" && rm -f "$ALT_NAME" && ln -sf "$SCRIPT_NAME" "$ALT_NAME")

# 6. Completion message
printf "\033[32m✓\033[0m Installation successful!\n"
printf "You can now run: \033[35m%s\033[0m (or \033[35m%s\033[0m)\n" "$SCRIPT_NAME" "$ALT_NAME"
