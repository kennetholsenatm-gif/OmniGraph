# ISR serial console → Ansible (socat + telnet transport)

The Cisco **async serial** console is **not** SSH. To run `cisco.ios` playbooks over the cable:

1. **socat** listens on `127.0.0.1:PORT` and forwards bytes **bidirectionally** to **COM3** (Windows) or **`/dev/ttyUSB0`** (Linux/WSL).
2. Ansible uses **`ansible_connection: ansible.netcommon.telnet`** to `127.0.0.1:PORT`. Despite the name, Ansible opens a **TCP** session; if the far end is raw serial (via socat), there is **no real Telnet server** on the router — you are piping bytes. The `telnet` **connection plugin** is still the usual choice for “TCP to a dumb CLI” on equipment that does not support SSH.

If you see **garbled prompts** or **stuck IAC** bytes, try a different Ansible/Python version or use a **hardware console server** that offers **SSH** to a logical console port (cleaner than raw socat in some cases).

## Windows + WSL2 + USB serial

1. Install **usbipd-win** and attach the USB–serial adapter to WSL so a device appears (often **`/dev/ttyUSB0`**).
2. Install **socat** in the distro: `sudo apt update && sudo apt install -y socat`.
3. From repo:  
   `.\scripts\socat-console-bridge.ps1 -WslDevice /dev/ttyUSB0 -Port 3322`
4. Verify: in WSL, `nc -v 127.0.0.1 3322` — you should see the **Router>/#** prompt.
5. Run playbooks with [`ansible/inventory/isr-console-bridge.example.yml`](../ansible/inventory/isr-console-bridge.example.yml) (copy to `isr-console-bridge.yml`).

**COM3 only visible in Windows (not WSL):** you must **usbipd attach** the adapter into WSL, or run Ansible from a machine that has the serial port natively (Linux laptop). Native Windows **socat** builds are uncommon; WSL + usbipd is the usual path.

## Linux

```bash
export ISR_CONSOLE_DEVICE=/dev/ttyUSB0
export ISR_CONSOLE_PORT=3322
./scripts/socat-console-bridge.sh
```

## Enable / passwords

Console may not ask for a **username**. Use inventory **`ansible_user: ""`** and set **`ansible_become_method: enable`** with **`ansible_become_password`** (vault) if the device requires **enable**.

After the ISR role aligns **`iac`**, **`kbolsen`**, and **SSH**, switch back to **`inventory/network.yml`** + **SSH** for normal runs.

## Playbook

```bash
cd ansible
ansible-playbook -i inventory/isr-console-bridge.yml playbooks/network-isr.yml \
  -e ansible_become_password='YOUR_ENABLE_SECRET'
```

Do **not** set `ansible_connection: network_cli` for this path unless your Ansible version documents **telnet** support under that plugin — the example inventory uses **`ansible.netcommon.telnet`**.
