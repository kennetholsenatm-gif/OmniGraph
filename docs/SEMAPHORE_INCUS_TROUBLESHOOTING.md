# Semaphore on Incus — troubleshooting

## Blank **entire** page (all white) at `http://localhost:3001`

If **nothing** renders (not even the shell), the SPA **JavaScript bundles did not load**. Common cause on this stack:

- **`SEMAPHORE_WEB_ROOT` was `http://127.0.0.1:3001`** → HTML has `<base href="http://127.0.0.1:3001/">` → the browser requests `js/*.js` from **`127.0.0.1`**. If **`http://127.0.0.1:3001` connection refused** but **`http://localhost:3001` works** (IPv6 / proxy only on `localhost`), **no JS runs** → white page.

**Fix (on the Incus host, from repo root):**

```bash
SEMAPHORE_WEB_ROOT=http://localhost:3001 ./scripts/fix-semaphore-incus-env.sh
```

Confirm:

```bash
curl -sS http://localhost:3001/ | grep -o '<base href="[^"]*"'
# expect: <base href="http://localhost:3001/">
```

Then **hard refresh** the browser (Ctrl+Shift+R). In **DevTools → Network**, `js/app.*.js` and `js/chunk-vendors.*.js` should be **200** from **`localhost`** (not failed to `127.0.0.1`).

## Default **`admin` / `admin_local_dev`** login does not work

The **Docker** image reads **`SEMAPHORE_ADMIN_*`** and creates the first user. **Native** (`semaphore server --no-config` + `/etc/semaphore/semaphore.env`) **does not** — those Ansible defaults are only used if we **bootstrap the user** via CLI.

**Fix (existing LXC):** from the repo on the Incus host:

```bash
./scripts/fix-semaphore-incus-env.sh
```

The script idempotently runs **`semaphore user add --no-config --admin ...`** when no matching login exists (defaults: `admin` / `admin_local_dev`). Override: `ADMIN_LOGIN=... ADMIN_PASS=... ./scripts/fix-semaphore-incus-env.sh`.

**Manual (inside the container):**

```bash
incus exec devsecops-semaphore -- systemd-run --pipe --wait \
  -p User=semaphore -p Group=semaphore -p WorkingDirectory=/var/lib/semaphore \
  -p EnvironmentFile=/etc/semaphore/semaphore.env \
  /usr/local/bin/semaphore user add --no-config --admin \
  --login admin --email admin@local.dev --name "Local Admin" --password 'admin_local_dev'
```

Then reset password if needed: `semaphore user change-by-login --help`.

New installs from an updated playbook run the same bootstrap at the end of **`semaphore-install.sh`**.

## Blank dashboard (sidebar/title OK, **white** main area)

Treat this as a **front-end + API** problem until proven otherwise — not “it’s always HTTP/URL”. Get **evidence** first, then fix what the evidence shows.

### 1) Browser (always do this)

- **DevTools → Console:** red errors (failed chunk load, JS exception, CSP) explain a lot of “white main panel” cases.
- **DevTools → Network:** reload, filter **Fetch/XHR** (or **All**). Note status codes for **`/api/`** and any **red** rows (4xx/5xx), **blocked**, or **(failed)**.
- Try **another browser** or **private window** (extensions / ad block sometimes break SPAs).

### 2) Server logs while you click around

```bash
incus exec devsecops-semaphore -- journalctl -u semaphore -f
```

Trigger the dashboard again and watch for **panic**, **401**, **500**, or DB errors.

### 3) Only if Network shows a **host / public URL** mismatch

Semaphore uses **`SEMAPHORE_WEB_ROOT`** for redirects and some links. If (and only if) responses or `Location` headers point at a **different host:port** than your tab, align env with what you actually use:

```bash
SEMAPHORE_WEB_ROOT=http://YOUR_TAB_HOST:3001 ./scripts/fix-semaphore-incus-env.sh
```

This is **one** knob — not a universal “blank UI” fix.

### 4) Other things people actually hit

- **Service still restarting** — fix `semaphore.service` first (`systemctl status`, journal).
- **Incus `proxy` device** — if API or **WebSocket** traffic behaves oddly, compare against [Semaphore’s expected ports / paths](https://docs.semaphoreui.com) and your proxy settings.
- **Fresh install / no data** — some views look empty until projects exist; confirm you’re past any **setup / login** flow.

### 5) Proof: `<base href="...">` must match the URL you actually use

Semaphore injects **`<base href="...">`** from **`SEMAPHORE_WEB_ROOT`**. Check with **the same host:port you use in the browser** (not a different loopback):

```bash
curl -sS http://localhost:3001/ | grep -o '<base href="[^"]*"'
```

If the HTML says e.g. **`http://127.0.0.1:3001/`** but you only ever open **`http://localhost:3001`**, the browser can treat those as **different sites** (cookies / same-origin), and the SPA may stay blank even though `/` is **200** and `/api/ping` is **pong**.

**Fix:** set **`SEMAPHORE_WEB_ROOT=http://localhost:3001`** (or whatever host:port you actually type) and restart Semaphore — see `SEMAPHORE_WEB_ROOT=... ./scripts/fix-semaphore-incus-env.sh`. **Do not** switch to **`http://127.0.0.1:3001`** if that host **refuses** on your machine (see **IPv4 vs IPv6 loopback** below).

## `Cannot Find configuration! Use --config parameter... semaphore setup`

Native installs that use **only** `/etc/semaphore/semaphore.env` must start the process with **`--no-config`**. Otherwise Semaphore looks for `config.json` on disk and exits immediately.

**Fix:** ensure the unit runs `semaphore server --no-config` (the install template and `scripts/fix-semaphore-incus-env.sh` do this). Manual override:

```bash
sudo mkdir -p /etc/systemd/system/semaphore.service.d
sudo tee /etc/systemd/system/semaphore.service.d/20-no-config.conf >/dev/null <<'EOF'
[Service]
ExecStart=
ExecStart=/usr/local/bin/semaphore server --no-config
EOF
sudo systemctl daemon-reload
sudo systemctl reset-failed semaphore
sudo systemctl restart semaphore
```

## `semaphore.service` **Active: failed** / exit code **1** (inside the container)

Semaphore **requires** `SEMAPHORE_COOKIE_HASH`, `SEMAPHORE_COOKIE_ENCRYPTION`, and `SEMAPHORE_ACCESS_KEY_ENCRYPTION` to be **valid base64** (it runs `base64.StdEncoding.DecodeString` on them). Older versions of the install script used `openssl rand -hex 32`, which is **not** base64 and causes startup to fail after `validateConfig` / cookie setup.

**1) Confirm in logs**

```bash
incus exec devsecops-semaphore -- journalctl -u semaphore -b --no-pager -n 80
```

**2) Fix secrets without rebuilding the LXC** (regenerates the three keys and restarts)

```bash
incus exec devsecops-semaphore -- bash -c '
  set -euo pipefail
  test -f /etc/semaphore/semaphore.env
  for k in SEMAPHORE_ACCESS_KEY_ENCRYPTION SEMAPHORE_COOKIE_HASH SEMAPHORE_COOKIE_ENCRYPTION; do
    val="$(openssl rand -base64 32 | tr -d "\n\r")"
    sed -i "/^${k}=/d" /etc/semaphore/semaphore.env
    echo "${k}=${val}" >> /etc/semaphore/semaphore.env
  done
  chmod 0600 /etc/semaphore/semaphore.env
  systemctl restart semaphore
'
```

**3) Permanent fix in repo:** reinstall script uses `openssl rand -base64 32` — re-run **`deploy-semaphore-incus.yml`** after pulling the update if you prefer Ansible to own the file.

### `FATAL: Ident authentication failed for user "semaphore"`

PostgreSQL reads **`pg_hba.conf` top-to-bottom**; the **first matching rule wins**. Many stock configs include **`host ... 127.0.0.1/32 ... ident`** (or similar) **before** any `md5`/`scram-sha-256` lines you append, so TCP connections to `127.0.0.1:5432` hit **`ident`** and fail.

**Fix:** prepend password-based rules for localhost TCP (same as the install script), then restart Postgres:

```bash
./scripts/fix-semaphore-incus-env.sh
```

Or manually **inside the LXC** (as root):

```bash
sudo bash -c 'tmp=$(mktemp); {
  echo "# tcp localhost password auth (must be before host ... ident lines)"
  echo "host all all 127.0.0.1/32 scram-sha-256"
  echo "host all all ::1/128 scram-sha-256"
  cat /var/lib/pgsql/data/pg_hba.conf
} > "$tmp"; mv -f "$tmp" /var/lib/pgsql/data/pg_hba.conf; chown postgres:postgres /var/lib/pgsql/data/pg_hba.conf; chmod 0600 /var/lib/pgsql/data/pg_hba.conf'
systemctl restart postgresql
```

Then re-run `./scripts/fix-semaphore-incus-env.sh` (or your `psql` test).

### Still failing after base64 fix? (Postgres SSL / `sslmode`)

Go’s `lib/pq` defaults to **`sslmode=prefer`**. A typical local Postgres on **127.0.0.1** without TLS can still make the client fail negotiation so **`semaphore server` exits immediately** with code **1** (often with little useful output unless you raise log level).

**One-shot patch (from repo root on the Incus host):**

```bash
chmod +x scripts/fix-semaphore-incus-env.sh
./scripts/fix-semaphore-incus-env.sh
# If you also need to regenerate the three base64 secrets again:
REGEN_SECRETS=1 ./scripts/fix-semaphore-incus-env.sh
```

This adds:

`SEMAPHORE_DB_OPTIONS={"sslmode":"disable"}`

to `/etc/semaphore/semaphore.env` and restarts the unit.

**Capture a useful log (recommended):**

```bash
incus exec devsecops-semaphore -- sed -i 's/^SEMAPHORE_LOG_LEVEL=.*/SEMAPHORE_LOG_LEVEL=debug/' /etc/semaphore/semaphore.env
# If the line is missing:
incus exec devsecops-semaphore -- grep -q '^SEMAPHORE_LOG_LEVEL=' /etc/semaphore/semaphore.env || \
  incus exec devsecops-semaphore -- sh -c 'echo SEMAPHORE_LOG_LEVEL=debug >> /etc/semaphore/semaphore.env'
incus exec devsecops-semaphore -- systemctl restart semaphore
incus exec devsecops-semaphore -- journalctl -u semaphore -b -n 200 --no-pager
```

## `Unit semaphore.service could not be found` (on the host)

**Expected.** Semaphore is installed **inside** the Incus LXC **`devsecops-semaphore`**, not on your Alma/WSL **host**. `systemctl status semaphore` on the host will always say the unit is missing.

Check the service **inside** the container:

```bash
incus list -c n,s | grep -E 'NAME|devsecops-semaphore'
incus exec devsecops-semaphore -- systemctl status semaphore --no-pager
```

If the instance does not exist or the install never finished, provision it:

```bash
cd ansible
ansible-playbook -i inventory/lxc.example.yml playbooks/deploy-semaphore-incus.yml \
  -e lxd_become=false -e lxd_manage_daemon=false -e lxd_ensure_idmap=false \
  -e lxd_incus_socket=/run/incus/unix.socket \
  -e 'lxd_apply_names=["devsecops-semaphore"]'
```

If **`systemctl status semaphore`** fails **inside** the container too, see logs:

```bash
incus exec devsecops-semaphore -- journalctl -u semaphore -n 100 --no-pager
```

## `ERR_EMPTY_RESPONSE` / **connection refused** on one of `localhost` vs `127.0.0.1`

**Cause:** **`localhost` and `127.0.0.1` are not the same socket.** Your browser/OS may use **IPv6** (`::1`) for `localhost`, while **`127.0.0.1`** is **IPv4-only**. The Incus **proxy** may listen on **only one** of them (or only on `0.0.0.0` / only on `[::]`).

| Symptom | Typical meaning |
|--------|------------------|
| **`http://localhost:3001` fails**, **`http://127.0.0.1:3001` works** | Often **IPv4 proxy only** — add an IPv6 listener (or fix `localhost` resolution). |
| **`http://localhost:3001` works**, **`http://127.0.0.1:3001` refuses** | Often **IPv6 / `::1` only** — **keep using `localhost`**; do not “fix” by switching to 127.0.0.1. |

### Fix (when `localhost` does not work but you need it)

1. If **`localhost` fails** and **`127.0.0.1` works**, add the IPv6 proxy (or re-run Ansible so **`semaphore_http6`** exists):

   ```bash
   incus config device add devsecops-semaphore semaphore_http6 proxy \
     listen=tcp:[::]:3001 connect=tcp:127.0.0.1:3000
   ```

   *(If the device already exists, skip this.)*

2. Re-run the deploy playbook so the **IPv6 proxy** is managed by Ansible (`semaphore_http6` → `[::]:3001`):
   ```bash
   cd ansible
   ansible-playbook -i inventory/lxc.example.yml playbooks/deploy-semaphore-incus.yml \
     -e lxd_become=false -e lxd_manage_daemon=false -e lxd_ensure_idmap=false \
     -e lxd_incus_socket=/run/incus/unix.socket \
     -e 'lxd_apply_names=["devsecops-semaphore"]'
   ```

### Manual check (inside WSL / Linux host)

```bash
# Instance running?
incus list | grep devsecops-semaphore

# Proxies present?
incus config device show devsecops-semaphore | sed -n '/semaphore_http/,/^$/p'

# Semaphore listening inside the container?
incus exec devsecops-semaphore -- ss -tlnp | grep 3000

# Service healthy?
incus exec devsecops-semaphore -- systemctl status semaphore --no-pager
```

### If the service is down

```bash
incus exec devsecops-semaphore -- journalctl -u semaphore -n 100 --no-pager
```

Typical causes: Postgres not ready, bad `SEMAPHORE_*` env, or port already in use.

### `SEMAPHORE_WEB_ROOT` / redirects

`semaphore_web_host` must match **exactly** what you type in the address bar (scheme + host + port), e.g. **`http://localhost:3001`** if that is what works on your host. Mismatch can break cookies or redirects. If **`127.0.0.1` refuses** on your machine, **do not** set WEB_ROOT to `http://127.0.0.1:3001`. Override when running the playbook, for example:

```bash
-e semaphore_web_host=http://localhost:3001
```

## Windows → WSL port forwarding

If Semaphore runs in **Incus inside WSL2**, use whichever loopback URL actually connects from Windows (**`localhost` or `127.0.0.1`** — they can differ). If one refuses, try the other, or confirm WSL is forwarding the port (recent WSL versions forward listening ports on `0.0.0.0` in the WSL VM). Align **`SEMAPHORE_WEB_ROOT`** with the URL that works, not an arbitrary loopback name.

## See also

- [LEAN_LOCAL_CONTROL_PLANE.md](opennebula-gitea-edge/LEAN_LOCAL_CONTROL_PLANE.md)
- [SEMAPHORE_POPULATE.md](SEMAPHORE_POPULATE.md) — remote Git + API template seed (Phase A)
- `scripts/start-semaphore.sh` / `scripts/start-semaphore.ps1`
