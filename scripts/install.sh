#!/usr/bin/env sh
set -eu

repo="${FZU_JWCH_REPO:-Seeridia/fzu-jwch-cli}"
binary="fzu-jwch"
install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
version="${FZU_JWCH_VERSION:-latest}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required command not found: $1" >&2
    exit 1
  fi
}

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux) echo "linux" ;;
    *)
      echo "error: unsupported OS: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) echo "amd64" ;;
    arm64 | aarch64) echo "arm64" ;;
    *)
      echo "error: unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

profile_file() {
  shell_name="$(basename "${SHELL:-sh}")"
  case "$shell_name" in
    zsh) echo "$HOME/.zshrc" ;;
    bash) echo "$HOME/.bashrc" ;;
    fish) echo "$HOME/.config/fish/config.fish" ;;
    *) echo "$HOME/.profile" ;;
  esac
}

add_path_hint() {
  case ":$PATH:" in
    *":$install_dir:"*) return 0 ;;
  esac

  shell_name="$(basename "${SHELL:-sh}")"
  profile="$(profile_file)"
  mkdir -p "$(dirname "$profile")"

  if [ "$shell_name" = "fish" ]; then
    line="fish_add_path \"$install_dir\""
  else
    line="export PATH=\"$install_dir:\$PATH\""
  fi

  if [ ! -f "$profile" ] || ! grep -F "$install_dir" "$profile" >/dev/null 2>&1; then
    {
      printf "\n# fzu-jwch\n"
      printf "%s\n" "$line"
    } >> "$profile"
    echo "Added $install_dir to PATH in $profile"
    echo "Restart your shell or run: . $profile"
  else
    echo "$install_dir is not in PATH. Check $profile and restart your shell."
  fi
}

need curl
need tar
need mktemp
need install

os="$(detect_os)"
arch="$(detect_arch)"

if [ "$version" = "latest" ]; then
  version="$(curl -fsSL "https://api.github.com/repos/$repo/releases/latest" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
fi

if [ -z "$version" ]; then
  echo "error: could not resolve latest release version" >&2
  exit 1
fi

asset="${binary}_${version}_${os}_${arch}.tar.gz"
url="https://github.com/$repo/releases/download/$version/$asset"
tmp="$(mktemp -d)"

cleanup() {
  rm -rf "$tmp"
}
trap cleanup EXIT INT TERM

echo "Downloading $url"
curl -fsSL "$url" -o "$tmp/$asset"

tar -xzf "$tmp/$asset" -C "$tmp"
mkdir -p "$install_dir"
install "$tmp/$binary" "$install_dir/$binary"

echo "Installed $binary to $install_dir/$binary"
add_path_hint

if command -v "$binary" >/dev/null 2>&1; then
  "$binary" --help >/dev/null
  echo "Run: $binary --help"
else
  echo "Run now: $install_dir/$binary --help"
fi
