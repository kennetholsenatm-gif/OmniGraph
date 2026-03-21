# Bootstrap USB bundle

Offline-first payload for bringing up a mini PC **without** relying on lab Git, Semaphore, or full routing.

**Full narrative:** [docs/BOOTSTRAP_USB_BUNDLE.md](../../docs/BOOTSTRAP_USB_BUNDLE.md)

## Layout (after build)

```
out/devsecops-bootstrap-YYYYMMDD/
  repo/           # copy of devsecops-pipeline
  collections/    # ansible-galaxy collection install (ansible_collections/)
  images/         # optional: populate with docker save (*.tar)
  MANIFEST.txt    # git SHA, date, optional image list
  bootstrap-on-target.sh  # symlink or copy from scripts/
```

## Build (Linux, connected)

```bash
cd scripts
chmod +x build-bundle.sh
./build-bundle.sh
```

Output directory is **`out/devsecops-bootstrap-<date>`** under this folder (gitignored).

## Run (target Alma host)

Mount USB, then:

```bash
sudo BUNDLE_ROOT=/path/to/devsecops-bootstrap-YYYYMMDD ./bootstrap-on-target.sh
```

**With Ansible LXC bootstrap** (inventory on target, `ansible-core` installed, `lxd_hosts` in inventory):

```bash
sudo RUN_ANSIBLE_BOOTSTRAP=1 \
  BOOTSTRAP_INVENTORY=/root/mini-pc-iam.yml \
  ANSIBLE_PLAYBOOK_EXTRA_ARGS='-K' \
  BUNDLE_ROOT=/path/to/devsecops-bootstrap-YYYYMMDD \
  ./bootstrap-on-target.sh
```

Copy `scripts/bootstrap-on-target.sh` into the bundle root when publishing to USB, or run it from the repo copy inside `repo/deployments/bootstrap-usb-bundle/scripts/` and set `BUNDLE_ROOT` to the parent of `repo/`.
