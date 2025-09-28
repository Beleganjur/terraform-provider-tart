#!/bin/bash
# tart-debootstrap-dietpi-vm.sh
# End-to-end: Checks dependencies, downloads debootstrap if missing, builds a minimal Debian arm64 rootfs, preps for Tart VM, adds DietPi installer.

set -euo pipefail

# --- Configurable Variables ---
IMAGE_SIZE="4G"
IMAGE_NAME="dietpi-tart-arm64"
WORKDIR="$HOME/tart-dietpi-tmp"
MOUNTDIR="$WORKDIR/mnt"
DEBIAN_SUITE="bookworm"
DEBIAN_MIRROR="http://deb.debian.org/debian"

# --- Dependency Installation ---
BREW_DEPS=(qemu coreutils tart)
echo "--- Checking Homebrew dependencies ---"
for pkg in "${BREW_DEPS[@]}"; do
    if ! brew list --formula | grep -qw "$pkg"; then
        echo ">> Installing $pkg..."
        brew install "$pkg"
    else
        echo ">> $pkg already installed."
    fi
done

# --- Working Directory Setup ---
mkdir -p "$WORKDIR"
mkdir -p "$MOUNTDIR"
cd "$WORKDIR"

# --- Check for debootstrap, download portable if missing ---
if ! command -v debootstrap >/dev/null; then
    echo "debootstrap not found, downloading portable version..."
    DEBOOTSTRAP_DEB="debootstrap_1.0.134_all.deb"
    curl -LO "http://ftp.debian.org/debian/pool/main/d/debootstrap/${DEBOOTSTRAP_DEB}"
    ar x "${DEBOOTSTRAP_DEB}"
    tar --strip-components=2 -xf data.tar.xz ./usr/sbin/debootstrap
    chmod +x debootstrap
    DEBOOTSTRAP_CMD="$PWD/debootstrap"
else
    DEBOOTSTRAP_CMD="$(command -v debootstrap)"
fi

# --- Create Blank Disk Image ---
echo "--- Creating blank disk image ---"
qemu-img create -f raw "${IMAGE_NAME}.img" "$IMAGE_SIZE"

# --- Format Disk Image ---
echo "--- Attaching and formatting disk image ---"
DEVLOOP=$(hdiutil attach -imagekey diskimage-class=CRawDiskImage -nomount "${IMAGE_NAME}.img" | grep '/dev/' | tail -1 | awk '{print $1}')
diskutil quiet eraseDisk JHFS+ "${IMAGE_NAME}" $DEVLOOP
hdiutil detach "$DEVLOOP"

# --- Mount Partition for debootstrap ---
DEVLOOP=$(hdiutil attach -mountpoint "$MOUNTDIR" "${IMAGE_NAME}.img" | grep '/dev/' | head -1 | awk '{print $1}')

# --- Bootstrap Debian ---
echo "--- Bootstrapping minimal ARM64 Debian ---"
sudo "$DEBOOTSTRAP_CMD" --arch=arm64 "$DEBIAN_SUITE" "$MOUNTDIR" "$DEBIAN_MIRROR"

# --- Configure root/ssh/network for first boot ---
echo "--- Setting up root user, SSH, & network ---"
sudo chroot "$MOUNTDIR" bash -c "
passwd -d root
apt-get update
apt-get install -y openssh-server sudo curl ca-certificates
echo 'PermitRootLogin yes' >> /etc/ssh/sshd_config
systemctl enable ssh
"

# --- Download DietPi installer into VM image ---
echo "--- Downloading DietPi installer into root home ---"
sudo chroot "$MOUNTDIR" bash -c "
curl -Lso /root/dietpi-installer 'https://raw.githubusercontent.com/MichaIng/DietPi/master/.build/images/dietpi-installer'
chmod +x /root/dietpi-installer
"

# --- Basic fstab and host config ---
echo "--- Writing fstab and hostname ---"
sudo bash -c "echo '/dev/vda1 / ext4 defaults 0 1' > '$MOUNTDIR/etc/fstab'"
echo "dietpi-tart" | sudo tee "$MOUNTDIR/etc/hostname"

# --- Cleanup and unmount ---
echo "--- Cleaning up and unmounting ---"
sudo umount "$MOUNTDIR" || hdiutil detach "$DEVLOOP"

# --- Convert image for Tart VM ---
echo "--- Importing image to Tart VM ---"
tart create "$IMAGE_NAME" --from-disk "${WORKDIR}/${IMAGE_NAME}.img"

echo
echo "-- SUCCESS! --"
echo "You can now launch your VM with:"
echo "  tart run $IMAGE_NAME"
echo "Log in as root (no password), then run:"
echo "  /root/dietpi-installer"
echo "on VM console to install DietPi."
echo

# Optionally: rm -rf "$WORKDIR"   # Clean up temp files if desired
