# Usage

## Prerequisites
- Go 1.22+
- Terraform 1.6+
- Tart CLI installed and working (`tart --version`)
- API and Executor from this repo running, or use the in-process tests

## Pulling official Tart images from GHCR requires login
Many Cirrus Labs Tart images hosted on GitHub Container Registry (GHCR) require authentication. If you reference images like these:
- Debian 13 (ARM64): `ghcr.io/cirruslabs/tart-debian:13.20240922`
- Ubuntu 22.04: `ghcr.io/cirruslabs/tart-ubuntu:22.04.20240921`
- DietPi ARM64: `ghcr.io/cirruslabs/tart-dietpi:20240901-arm64`

you must log in to GHCR or you will see errors such as:
```
Error: AuthFailed(why: "received unexpected HTTP status code 403 while retrieving an authentication token", details: "{\"errors\":[{\"code\":\"DENIED\",\"message\":\"requested access to the resource is denied\"}]}")
```

### How to obtain a token and log in
1. Create a GitHub Personal Access Token (PAT) with the scope:
   - `read:packages`
   
   Steps:
   - GitHub → Settings → Developer settings → Personal access tokens
   - Create a classic token (or a fine-grained token with package read access to the Cirrus Labs org/repo)
   - Copy the token value

2. Log in to GHCR for Tart:
   ```bash
   tart login ghcr.io -u <github-username> -p <token-with-read:packages>
   ```

3. Verify you can pull an image:
   ```bash
   tart pull ghcr.io/cirruslabs/tart-debian:13.20240922
   ```

After this, the executor (which shells out to `tart`) will be able to pull the same images successfully.

### How to use `tart login` correctly
- Always specify the full registry hostname. For GHCR this is `ghcr.io`.

Command syntax:
```bash
tart login <host> [--username <username>] [--password-stdin] [--insecure] [--no-validate]
```

Examples:
```bash
# Interactive (you will be prompted for username and password/token)
tart login ghcr.io

# Non-interactive (useful for automation)
echo "$GITHUB_TOKEN" | tart login ghcr.io --username "$GITHUB_USER" --password-stdin
```

Optional flags:
- `--username <username>`: specify username explicitly
- `--password-stdin`: read the password/token from stdin (CI-friendly)
- `--insecure`: allow insecure connections (not recommended)
- `--no-validate`: skip server certificate validation

Reference: Tart CLI login docs
- https://docs.cirruslabs.org/tart/cli/login/

### Notes
- `tart login` persists credentials (e.g., in keychain on macOS). The executor uses the same Tart CLI context, so no additional configuration is needed for it.
- If you prefer not to authenticate, use local images already present on your host (shown by `tart list`) and set `image` to that local name. Our API will skip `pull` for names that look local.

## Running the sample Terraform configs
- Ensure API and Executor are running:
  ```bash
  make api
  make executor
  ```
- Initialize and apply:
  ```bash
  terraform init -upgrade -input=false
  terraform apply -auto-approve
  ```

See `test.tf` for examples using the GHCR images above.

## Quick start: install Tart and try a base image

If you just want to verify Tart itself and a base image end-to-end:

```bash
# 1) Install Tart (Homebrew)
brew install cirruslabs/cli/tart

# 2) (Optional but often required) Login to GHCR if the image is private/restricted
#    Many Cirrus Labs images require GHCR auth. If unauthenticated pulls fail with 403,
#    perform the login step described above.
# tart login ghcr.io

# 3) Clone a macOS base image and run it
tart clone ghcr.io/cirruslabs/macos-tahoe-base:latest tahoe-base
tart run tahoe-base
```

Notes:
- Whether credentials are required depends on the image’s visibility and your access. If `tart clone` fails with an auth error (e.g., 403 DENIED), login to GHCR as described in the "How to obtain a token and log in" section above and retry.
- After a successful clone, `tart run tahoe-base` should boot the VM.

## One-command console (tmux)

Launch a ready-to-use console with API, executor, and a shell in a single tmux session.

### Prerequisites
- Ensure tmux is installed:
  ```bash
  make deps.tmux
  ```

### Start the console
```bash
make console
```

This runs `addons/tmux-console.sh`, which creates a tmux session with three panes:

```
┌──────────────────────────────────────┬──────────────────────────────────────┐
│ Pane 0: API                          │ Pane 1: Executor                     │
│ (runs `make api`)                    │ (runs `make executor`)               │
├──────────────────────────────────────┴──────────────────────────────────────┤
│ Pane 2: Shell (your interactive pane)                                      │
└────────────────────────────────────────────────────────────────────────────┘
```

### Controls (tmux)
- Detach: `Ctrl-b` then `d`
- Switch panes: `Ctrl-b` then arrow keys
- Kill the session: `tmux kill-session -t <name>`

### Session name
- Configured via `.env` variable `BSS_TMUX_CONSOLE_TART_TF`.
- Default if not set: `beleganjur_services`.

Example `.env`:
```env
BSS_TMUX_CONSOLE_TART_TF=beleganjur_services
```
