#!/usr/bin/env bash

# Bootstrap a fresh Ubuntu Desktop VM for the LGTM project.
#
# Intended workflow:
#   1. Install Ubuntu Desktop in a VM.
#   2. Copy this file into the VM.
#   3. Run it inside the VM with:
#        sudo bash bootstrap-lgtm-vm.sh
#
# The Ubuntu Desktop image already provides the graphical desktop and browser.
# This script only installs the project infrastructure we need on top of it:
# SSH, development tools, Podman, Go, and k3s.

set -euo pipefail

# apt can ask interactive questions. In a bootstrap script we want predictable,
# non-interactive package installation instead.
export DEBIAN_FRONTEND=noninteractive

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

# Pin k3s so rerunning the script later does not silently install a newer
# Kubernetes version. Override when needed:
#   sudo K3S_VERSION=v1.37.0+k3s1 bash bootstrap-lgtm-vm.sh
readonly K3S_VERSION="${K3S_VERSION:-v1.37.0+k3s1}"

# Set INSTALL_K3S=0 when you only want the base VM packages:
#   sudo INSTALL_K3S=0 bash bootstrap-lgtm-vm.sh
readonly INSTALL_K3S="${INSTALL_K3S:-1}"

# The k3s installer is downloaded only temporarily and removed on exit.
readonly K3S_INSTALLER="/tmp/install-k3s.sh"

# Packages needed by this project on top of a normal Ubuntu Desktop install.
# declare -ra means: this is a read-only array.
# We intentionally do not install Firefox, GNOME, XFCE, or a display manager:
# Ubuntu Desktop already provides the desktop experience.
declare -ra APT_PACKAGES=(
    ca-certificates
    curl
    git
    make
    openssh-server
    rsync
    golang-go
    podman
    uidmap
    slirp4netns
)

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

fail() {
    echo "ERROR: $*" >&2
    exit 1
}

log_step() {
    echo
    echo "==> $*"
}

cleanup() {
    rm -f "${K3S_INSTALLER}"
}

require_root() {
    if [[ "${EUID}" -ne 0 ]]; then
        fail "Please run this script as root: sudo bash bootstrap-lgtm-vm.sh"
    fi
}

require_ubuntu() {
    if [[ ! -r /etc/os-release ]]; then
        fail "Cannot detect operating system: /etc/os-release is missing."
    fi

    # /etc/os-release defines variables such as ID and PRETTY_NAME.
    # shellcheck disable=SC1091
    . /etc/os-release

    if [[ "${ID:-}" != "ubuntu" ]]; then
        fail "This bootstrap script expects Ubuntu. Detected: ${PRETTY_NAME:-unknown}"
    fi
}

# ---------------------------------------------------------------------------
# Installation steps
# ---------------------------------------------------------------------------

install_apt_packages() {
    log_step "Refreshing package index"
    apt-get update

    log_step "Installing project packages"
    apt-get install -y --no-install-recommends "${APT_PACKAGES[@]}"
}

enable_ssh_service() {
    # openssh-server installs the ssh service. Enabling it makes SSH available
    # after reboots; --now also starts it immediately.
    log_step "Enabling SSH service"
    systemctl enable --now ssh
}

installed_k3s_version() {
    if ! command -v k3s >/dev/null 2>&1; then
        return 1
    fi

    k3s --version | awk 'NR == 1 { print $3 }'
}

install_k3s() {
    local current_version=""

    if [[ "${INSTALL_K3S}" != "1" ]]; then
        log_step "Skipping k3s installation"
        return
    fi

    if current_version="$(installed_k3s_version)"; then
        if [[ "${current_version}" == "${K3S_VERSION}" ]]; then
            log_step "k3s ${K3S_VERSION} is already installed"
            return
        fi

        fail "k3s ${current_version} is already installed; expected ${K3S_VERSION}."
    fi

    log_step "Downloading k3s installer"
    curl \
        --fail \
        --silent \
        --show-error \
        --location \
        https://get.k3s.io \
        --output "${K3S_INSTALLER}"

    chmod 0755 "${K3S_INSTALLER}"

    log_step "Installing k3s ${K3S_VERSION}"
    INSTALL_K3S_VERSION="${K3S_VERSION}" "${K3S_INSTALLER}"
}

# ---------------------------------------------------------------------------
# Final output
# ---------------------------------------------------------------------------

print_summary() {
    local vm_user="${SUDO_USER:-${USER:-unknown}}"

    log_step "Bootstrap complete"
    echo "User: ${vm_user}"
    echo "SSH service: enabled"
    echo "k3s version: ${K3S_VERSION}"

    echo
    echo "Useful checks inside the VM:"
    echo "  ip addr"
    echo "  sudo kubectl get nodes"
    echo "  go version"
    echo "  podman version"
}

main() {
    trap cleanup EXIT

    require_root
    require_ubuntu
    install_apt_packages
    enable_ssh_service
    install_k3s
    print_summary
}

main "$@"
