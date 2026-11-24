#! /bin/sh

# Figure out OS and ARCH
OS="`uname`"
ARCH="`uname -m`"
VERSION=v0.1.0
OSARCH=
FORMAT=tar.gz
case $OS in
  'Linux')
    case $ARCH in
        'x86_64')
            OSARCH='linux-amd64'
            ;;
        'armv8')
            OSARCH='linux-arm64'
            ;;
        'armv7l')
            OSARCH='linux-arm64'
            ;;
        'aarch64')
            OSARCH='linux-arm64'
            ;;
    esac
    ;;
  'Darwin')
    case $ARCH in
        'x86_64')
            OSARCH='apple-amd64'
            ;;
        'arm64')
            OSARCH='apple-arm64'
            ;;
    esac
    ;;
  *) ;;
esac


# Download
echo "Installing berglas9 ${VERSION}..."
echo "   for ${OSARCH} system"
curl -o /tmp/berglas9.${FORMAT} -fsSL https://github.com/otcom-dev/berglas9/releases/download/${VERSION}/berglas9-${VERSION}-${OSARCH}.${FORMAT}

# Check if the user is root and adjust the installation directory accordingly
INSTALL_DIR='~/.local/bin'
if [ "$(id -u)" -eq 0 ]; then
  INSTALL_DIR='/usr/local/bin'
fi
echo "   into ${INSTALL_DIR}"

# Install
if [ -e /tmp/berglas9.${FORMAT} ]; then
    mkdir -p ${INSTALL_DIR}
    rm -f ${INSTALL_DIR}/berglas9 && \
      echo "    * Removing existing executable" && \
      tar xfz /tmp/berglas9.${FORMAT} -C ${INSTALL_DIR}/ && \
      echo "    * Extracting" && \
      rm /tmp/berglas9.${FORMAT} && \
      echo "    * Success!"
    exit $?
fi

echo "  * Failed due to some reason. Please try manually downloading berglas9 and copy the binary to ~/.local/bin"
exit 1
