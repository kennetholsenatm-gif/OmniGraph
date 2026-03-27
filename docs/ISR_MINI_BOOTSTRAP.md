# Mini IOS‑XE bootstrap (two ISRs → Ansible `iac` + human `kbolsen`)

Paste from **enable / configure terminal** on each ISR (console or existing SSH). Adjust **hostnames**, **passwords**, and **management addressing** to your lab.

**Goal:** SSH works with **`iac` + password** (Ansible `network_cli`). **`kbolsen`** exists with **no password** and **only** the SSH public key below (matches `ansible/roles/cisco_isr_platform/files/ssh/kbolsen.pub`).

> **Security:** `secret 0 …` keeps the password in **clear text** in `show run` — fine for lab; rotate and use Vault/Ansible for real configs.

---

## ISR primary (example)

```text
hostname isr01
ip domain name edge.lab

! Host key for SSH (idempotent if key already exists — skip errors or use "yes" when prompted)
crypto key generate rsa general-keys modulus 2048
ip ssh version 2
! Defaults on modern IOS‑XE already allow publickey + password for local users.

! Ansible / automation login (password + optional key later from the role)
username iac privilege 15 secret 0 CHANGEME_IAC_PASSWORD

! Human admin: key-only (nopassword)
username kbolsen privilege 15 nopassword
ip ssh pubkey-chain
 username kbolsen
  key-string
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBu+jzq0cmgsAQ+JgiwEAaNQKSPWuNPlxN/3JjI6gTTv
 exit
 exit

line vty 0 15
 login local
 transport input ssh
 exec-timeout 10 0
```

---

## ISR secondary (example)

Same block as above; only change **`hostname`** (and passwords if you want them distinct):

```text
hostname isr02
ip domain name edge.lab

crypto key generate rsa general-keys modulus 2048
ip ssh version 2

username iac privilege 15 secret 0 CHANGEME_IAC_PASSWORD

username kbolsen privilege 15 nopassword
ip ssh pubkey-chain
 username kbolsen
  key-string
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBu+jzq0cmgsAQ+JgiwEAaNQKSPWuNPlxN/3JjI6gTTv
 exit
 exit

line vty 0 15
 login local
 transport input ssh
 exec-timeout 10 0
```

---

## After paste

1. Ensure at least one **L3 interface** has the IP you use in **`ansible_host`** (inventory).
2. From your control node: `ssh iac@<ansible_host>` — password should match **`CHANGEME_IAC_PASSWORD`**.
3. From your laptop: `ssh -i ~/.ssh/kbolsen_ed25519 kbolsen@<ansible_host>` (or whatever private key matches the pubkey above).
4. Run Ansible with **`ansible_user: iac`** and **`ansible_password`** (or `--ask-pass`), then the **`cisco_isr_platform`** role to align banner, logging, interfaces, etc.

If **`crypto key generate`** complains the key exists, use `show crypto key mypubkey rsa` to confirm SSH is already set up and skip that line.

If **`key-string`** is rejected on your IOS‑XE train, finish **`kbolsen`** with Ansible instead: the role uses `cisco.ios.ios_user` to install the same key.
