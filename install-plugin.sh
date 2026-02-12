#!/bin/sh
set -e

REPO="0xbe1/aptly"
INSTALL_DIR="/usr/local/bin"
PLUGINS=""

usage() {
    cat <<'EOF'
Usage: install-plugin.sh [options] <plugin> [plugin...]

Install one or more aptly plugin binaries from GitHub releases.

Plugins:
  move-decompiler    (alias: decompiler)
  aptos-tracer       (alias: tracer)
  aptos-script-compose (alias: compose)

Options:
  --install-dir <dir>  Install destination (default: /usr/local/bin)
  --all                Install all plugins
  --list               Show available plugins
  -h, --help           Show this help

Examples:
  install-plugin.sh move-decompiler
  install-plugin.sh decompiler tracer
  install-plugin.sh --all
EOF
}

list_plugins() {
    cat <<'EOF'
move-decompiler
aptos-tracer
aptos-script-compose
EOF
}

canonical_plugin() {
    case "$1" in
        move-decompiler|decompiler) echo "move-decompiler" ;;
        aptos-tracer|tracer) echo "aptos-tracer" ;;
        aptos-script-compose|compose) echo "aptos-script-compose" ;;
        *) return 1 ;;
    esac
}

add_plugin() {
    plugin="$(canonical_plugin "$1")" || {
        echo "Unknown plugin: $1"
        echo "Run with --list to see supported plugins."
        exit 1
    }

    case " $PLUGINS " in
        *" $plugin "*) ;;
        *) PLUGINS="$PLUGINS $plugin" ;;
    esac
}

while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help)
            usage
            exit 0
            ;;
        --list)
            list_plugins
            exit 0
            ;;
        --all)
            add_plugin move-decompiler
            add_plugin aptos-tracer
            add_plugin aptos-script-compose
            shift
            ;;
        --install-dir)
            [ $# -ge 2 ] || { echo "Missing value for --install-dir"; exit 1; }
            INSTALL_DIR="$2"
            shift 2
            ;;
        --*)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
        *)
            add_plugin "$1"
            shift
            ;;
    esac
done

if [ -z "${PLUGINS# }" ]; then
    echo "No plugins selected."
    usage
    exit 1
fi

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *) echo "Unsupported OS: $OS" && exit 1 ;;
esac

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

TMP_DIR="$(mktemp -d)"
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

get_tag() {
    tag_prefix="$1"
    curl -fsSL "https://api.github.com/repos/$REPO/releases?per_page=100" \
      | grep -o "\"tag_name\":[[:space:]]*\"${tag_prefix}[^\"]*\"" \
      | head -n1 \
      | sed -E 's/.*"([^"]+)"/\1/'
}

install_plugin() {
    plugin="$1"
    case "$plugin" in
        move-decompiler|aptos-tracer)
            tag_prefix="aptly-cli-v"
            ;;
        aptos-script-compose)
            tag_prefix="aptos-script-compose-v"
            ;;
        *)
            echo "Unsupported plugin: $plugin"
            exit 1
            ;;
    esac

    tag="$(get_tag "$tag_prefix")"
    if [ -z "$tag" ]; then
        echo "Failed to fetch latest release tag for $plugin (${tag_prefix}*)"
        exit 1
    fi

    archive="${plugin}_${tag}_${OS}_${ARCH}.tar.gz"
    url="https://github.com/$REPO/releases/download/$tag/$archive"

    echo "Installing $plugin ${tag} for ${OS}/${ARCH}..."
    curl -fsSL "$url" | tar xz -C "$TMP_DIR"

    if [ ! -f "$TMP_DIR/$plugin" ]; then
        echo "Downloaded archive did not contain expected binary: $plugin"
        exit 1
    fi
    chmod +x "$TMP_DIR/$plugin"

    if [ -d "$INSTALL_DIR" ] && [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_DIR/$plugin" "$INSTALL_DIR/$plugin"
    else
        echo "Installing to $INSTALL_DIR (requires sudo)..."
        sudo mkdir -p "$INSTALL_DIR"
        sudo mv "$TMP_DIR/$plugin" "$INSTALL_DIR/$plugin"
    fi

    echo "Installed $plugin to $INSTALL_DIR/$plugin"
}

for plugin in $PLUGINS; do
    install_plugin "$plugin"
done
