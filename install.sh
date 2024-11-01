#!/bin/sh
set -e
# Code generated by godownloader on 2024-10-21T13:02:39Z. DO NOT EDIT.
#

usage() {
  this=$1
  cat <<EOF
$this: download go binaries for bruin-data/bruin

Usage: $this [-b] bindir [-d] [tag]
  -b sets bindir or installation directory, Defaults to ./bin
  -d turns on debug logging
   [tag] is a tag from
   https://github.com/bruin-data/bruin/releases
   If tag is missing, then the latest will be used.

 Generated by godownloader
  https://github.com/goreleaser/godownloader

EOF
  exit 2
}

parse_args() {
  #BINDIR is ./bin unless set be ENV
  # over-ridden by flag below

  BINDIR=${BINDIR:-~/.local/bin}
  while getopts "b:dh?x" arg; do
    case "$arg" in
      b) BINDIR="$OPTARG" ;;
      d) log_set_priority 10 ;;
      h | \?) usage "$0" ;;
      x) set -x ;;
    esac
  done
  shift $((OPTIND - 1))
  TAG=$1
}

# this function wraps all the destructive operations
# if a curl|bash cuts off the end of the script due to
# network, either nothing will happen or will syntax error
# out preventing half-done work
execute() {
  
  tmpdir=$(mktemp -d)
  log_debug "downloading files into ${tmpdir}"
  http_download "${tmpdir}/${TARBALL}" "${TARBALL_URL}"
  srcdir="${tmpdir}"
  (cd "${tmpdir}" && untar "${TARBALL}")

  test ! -d "${BINDIR}" && install -d "${BINDIR}"
  log_info "installing to ${BINDIR}"
  for binexe in $BINARIES; do
    if [ "$OS" = "windows" ]; then
      binexe="${binexe}.exe"
    fi
    install "${srcdir}/${binexe}" "${BINDIR}/"
    echo "  ${binexe}"
    log_info "installed ${BINDIR}/${binexe}"
  done
  rm -rf "${tmpdir}"
  


  current_shell=$(basename "$SHELL")

 
  case "$current_shell" in
    bash|sh)
      export_command="export PATH=\"\$PATH:${BINDIR}\""
      eval "$export_command"
      echo "$export_command" >> "$HOME/.${current_shell}rc"
      export PATH="$PATH:${BINDIR}"

      log_info "The PATH has been updated in your current shell session. You can now use the installed binaries without restarting your shell."
      ;;
    zsh)
      export_command="export PATH=\"\$PATH:${BINDIR}\""
      eval "$export_command"
      echo "$export_command" >> "$HOME/.zshrc"
      # Export PATH in the current shell
      export PATH="$PATH:${BINDIR}"

      log_info "The PATH has been updated in your current shell session. You can now use the installed binaries without restarting your shell."
      ;;
    fish)
      export_command="set -gx PATH \$PATH ${BINDIR}"
      fish -c "$export_command"
      echo "$export_command" >> "$HOME/.config/fish/config.fish"
      # Export PATH in the current shell (for fish, this is already done by the fish -c command)

      log_info "The PATH has been updated in your current shell session. You can now use the installed binaries without restarting your shell."
      ;;
    *)
      export_command="export PATH=\"\$PATH:${BINDIR}\""
      eval "$export_command"
      log_info "# Unable to automatically add to your shell's configuration file."
      log_info "# Please manually add the appropriate line to your shell's configuration file to make it permanent."
      # Export PATH in the current shell
      export PATH="$PATH:${BINDIR}"
      ;;
  esac
  

  log_info "everything's installed!"
}


get_binaries() {
  case "$PLATFORM" in
    darwin/amd64) BINARIES="bruin" ;;
    darwin/arm64) BINARIES="bruin" ;;
    linux/amd64) BINARIES="bruin" ;;
    linux/arm64) BINARIES="bruin" ;;
    windows/386) BINARIES="bruin" ;;
    windows/amd64) BINARIES="bruin" ;;
    *)
      log_crit "platform $PLATFORM is not supported.  Make sure this script is up-to-date and file request at https://github.com/${PREFIX}/issues/new"
      exit 1
      ;;
  esac
}
tag_to_version() {
  if [ -z "${TAG}" ]; then
    log_info "checking GitHub for latest tag"
  else
    log_info "checking GitHub for tag '${TAG}'"
  fi
  log_debug "will get github for "$OWNER/$REPO" and tag '${TAG}'"
  REALTAG=$(github_release "$OWNER/$REPO" "${TAG}") && true
  log_debug "REALTAG is '${REALTAG}'"

  if test -z "$REALTAG"; then
    log_crit "unable to find '${TAG}' - use 'latest' or see https://github.com/${PREFIX}/releases for details"
    exit 1
  fi
  log_debug "will run for tag ${REALTAG}, version -> ${TAG#v}"
  # if version starts with 'v', remove it
  TAG="$REALTAG"
  VERSION=${TAG#v}
}
adjust_format() {
  # change format (tar.gz or zip) based on OS
  case ${OS} in
    windows) FORMAT=zip ;;
  esac
  true
}
adjust_os() {
  # adjust archive name based on OS
  case ${OS} in
    386) OS=i386 ;;
    amd64) OS=x86_64 ;;
    darwin) OS=Darwin ;;
    linux) OS=Linux ;;
    windows) OS=Windows ;;
  esac
  true
}
adjust_arch() {
  # adjust archive name based on ARCH
  case ${ARCH} in
    386) ARCH=i386 ;;
    amd64) ARCH=x86_64 ;;
    darwin) ARCH=Darwin ;;
    linux) ARCH=Linux ;;
    windows) ARCH=Windows ;;
  esac
  true
}

cat /dev/null <<EOF
------------------------------------------------------------------------
https://github.com/client9/shlib - portable posix shell functions
Public domain - http://unlicense.org
https://github.com/client9/shlib/blob/master/LICENSE.md
but credit (and pull requests) appreciated.
------------------------------------------------------------------------
EOF
is_command() {
  command -v "$1" >/dev/null
}
echoerr() {
  echo "$@" 1>&2
}
log_prefix() {
  echo "$0"
}
_logp=6
log_set_priority() {
  _logp="$1"
}
log_priority() {
  if test -z "$1"; then
    echo "$_logp"
    return
  fi
  [ "$1" -le "$_logp" ]
}
log_tag() {
  case $1 in
    0) echo "emerg" ;;
    1) echo "alert" ;;
    2) echo "crit" ;;
    3) echo "err" ;;
    4) echo "warning" ;;
    5) echo "notice" ;;
    6) echo "info" ;;
    7) echo "debug" ;;
    *) echo "$1" ;;
  esac
}
log_debug() {
  log_priority 7 || return 0
  echoerr "$(log_prefix)" "$(log_tag 7)" "$@"
}
log_info() {
  log_priority 6 || return 0
  echoerr "$(log_prefix)" "$(log_tag 6)" "$@"
}
log_err() {
  log_priority 3 || return 0
  echoerr "$(log_prefix)" "$(log_tag 3)" "$@"
}
log_crit() {
  log_priority 2 || return 0
  echoerr "$(log_prefix)" "$(log_tag 2)" "$@"
}
uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    cygwin_nt*) os="windows" ;;
    mingw*) os="windows" ;;
    msys_nt*) os="windows" ;;
  esac
  echo "$os"
}
uname_arch() {
  arch=$(uname -m)
  case $arch in
    x86_64) arch="amd64" ;;
    x86) arch="386" ;;
    i686) arch="386" ;;
    i386) arch="386" ;;
    aarch64) arch="arm64" ;;
    armv5*) arch="armv5" ;;
    armv6*) arch="armv6" ;;
    armv7*) arch="armv7" ;;
  esac
  echo ${arch}
}
uname_os_check() {
  os=$(uname_os)
  case "$os" in
    darwin) return 0 ;;
    dragonfly) return 0 ;;
    freebsd) return 0 ;;
    linux) return 0 ;;
    android) return 0 ;;
    nacl) return 0 ;;
    netbsd) return 0 ;;
    openbsd) return 0 ;;
    plan9) return 0 ;;
    solaris) return 0 ;;
    windows) return 0 ;;
  esac
  log_crit "uname_os_check '$(uname -s)' got converted to '$os' which is not a GOOS value. Please file bug at https://github.com/client9/shlib"
  return 1
}
uname_arch_check() {
  arch=$(uname_arch)
  case "$arch" in
    386) return 0 ;;
    amd64) return 0 ;;
    arm64) return 0 ;;
    armv5) return 0 ;;
    armv6) return 0 ;;
    armv7) return 0 ;;
    ppc64) return 0 ;;
    ppc64le) return 0 ;;
    mips) return 0 ;;
    mipsle) return 0 ;;
    mips64) return 0 ;;
    mips64le) return 0 ;;
    s390x) return 0 ;;
    amd64p32) return 0 ;;
  esac
  log_crit "uname_arch_check '$(uname -m)' got converted to '$arch' which is not a GOARCH value.  Please file bug report at https://github.com/client9/shlib"
  return 1
}
untar() {
  tarball=$1
  case "${tarball}" in
    *.tar.gz | *.tgz) tar --no-same-owner -xzf "${tarball}" ;;
    *.tar) tar --no-same-owner -xf "${tarball}" ;;
    *.zip) unzip "${tarball}" ;;
    *)
      log_err "untar unknown archive format for ${tarball}"
      return 1
      ;;
  esac
}
http_download_curl() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    log_debug "Executing: curl  -sL -o \"$local_file\" \"$source_url\""
    code=$(curl  -sL -o "$local_file" "$source_url")
  else
    log_debug "Executing: curl  -sL -H \"$header\" -o \"$local_file\" \"$source_url\""
    code=$(curl  -sL -H "$header" -o "$local_file" "$source_url")
  fi
  log_debug "http_download_curl received HTTP status $code, return code $?"
  if [ $? -ne 0 ]; then
    log_err "Failed to download from $source_url"
    return 1  # Return 1 for failure
  fi
  return 0
}
http_download_wget() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    wget -q -O "$local_file" "$source_url"
  else
    wget -q --header "$header" -O "$local_file" "$source_url"
  fi
}
http_download() {
  log_debug "http_download $2"
  if is_command curl; then
    http_download_curl "$@"
    return
  elif is_command wget; then
    http_download_wget "$@"
    return
  fi
  log_crit "http_download unable to find wget or curl"
  return 1
}
http_copy() {
  tmp=$(mktemp)
  http_download "${tmp}" "$1" "$2" || return 1
  body=$(cat "$tmp")
  rm -f "${tmp}"
  echo "$body"
}
github_release() {
  owner_repo=$1
  version=$2
  test -z "$version" && version="latest"
  giturl="https://github.com/${owner_repo}/releases/${version}"
  log_debug "will get github releases from ${giturl}"
  json=$(http_copy "$giturl" "Accept:application/json")
  log_debug "json release info is $json"
  test -z "$json" && return 1
  version=$(echo "$json" | tr -s '\n' ' ' | sed 's/.*"tag_name":"//' | sed 's/".*//')
  log_debug "Version found : '$version'"
  test -z "$version" && return 1
  echo "$version"
}
hash_sha256() {
  TARGET=${1:-/dev/stdin}
  if is_command gsha256sum; then
    hash=$(gsha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command sha256sum; then
    hash=$(sha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command shasum; then
    hash=$(shasum -a 256 "$TARGET" 2>/dev/null) || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command openssl; then
    hash=$(openssl -dst openssl dgst -sha256 "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f a
  else
    log_crit "hash_sha256 unable to find command to compute sha-256 hash"
    return 1
  fi
}
hash_sha256_verify() {
  TARGET=$1
  checksums=$2
  if [ -z "$checksums" ]; then
    log_err "hash_sha256_verify checksum file not specified in arg2"
    return 1
  fi
  BASENAME=${TARGET##*/}
  want=$(grep "${BASENAME}" "${checksums}" 2>/dev/null | tr '\t' ' ' | cut -d ' ' -f 1)
  if [ -z "$want" ]; then
    log_err "hash_sha256_verify unable to find checksum for '${TARGET}' in '${checksums}'"
    return 1
  fi
  got=$(hash_sha256 "$TARGET")
  if [ "$want" != "$got" ]; then
    log_err "hash_sha256_verify checksum for '$TARGET' did not verify ${want} vs $got"
    return 1
  fi
}
cat /dev/null <<EOF
------------------------------------------------------------------------
End of functions from https://github.com/client9/shlib
------------------------------------------------------------------------
EOF

PROJECT_NAME="bruin"
OWNER=bruin-data
REPO="bruin"
BINARY=bruin
FORMAT=tar.gz
OS=$(uname_os)
ARCH=$(uname_arch)
PREFIX="$OWNER/$REPO"

# use in logging routines
log_prefix() {
	echo "$PREFIX"
}
PLATFORM="${OS}/${ARCH}"
log_debug "Running for ${PLATFORM}"
GITHUB_DOWNLOAD=https://github.com/${OWNER}/${REPO}/releases/download

uname_os_check "$OS"
uname_arch_check "$ARCH"

parse_args "$@"

get_binaries

tag_to_version

adjust_format
log_debug "Will download ${FORMAT} format"

adjust_os
log_debug "OS adjusted to ${OS}"

adjust_arch
log_debug "ARCH adjusted to ${ARCH}"

log_info "found version: ${VERSION} for ${TAG}/${OS}/${ARCH}"

NAME=${PROJECT_NAME}_${OS}_${ARCH}
TARBALL=${NAME}.${FORMAT}
TARBALL_URL=${GITHUB_DOWNLOAD}/${TAG}/${TARBALL}

log_info "Starting the dowload of ${TARBALL_URL}"



execute





