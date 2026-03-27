# Publishing `wiki/` to GitHub Wiki

GitHub Wikis are stored in a **second** repository: `https://github.com/kennetholsenatm-gif/OmniGraph.wiki.git`  
The [`wiki/` folder on `main`](https://github.com/kennetholsenatm-gif/OmniGraph/tree/main/wiki) is the source copy next to the code. **CI can push it for you** (see below); you can still sync by hand anytime.

## Automated sync (CI)

On every push to `main` that changes files under **`wiki/`**, the **[Wiki sync workflow](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/.github/workflows/wiki-sync.yml)** mirrors this folder into the GitHub Wiki.

**One-time setup for maintainers**

1. Ensure **Wikis** are enabled (step 1 below) and the Wiki has at least an initial **Home** page so the `.wiki` git repo exists.
2. Create a **fine-grained or classic PAT** with **`Contents` read and write** on this repository (the PAT must be allowed to push the Wiki git remote).
3. In the repo on GitHub: **Settings → Secrets and variables → Actions → New repository secret**  
   - Name: `WIKI_PUSH_TOKEN`  
   - Value: the PAT

If the secret is missing, the workflow **skips** the push with a notice (CI stays green). Forks without the secret behave the same way.

You can also run the workflow manually: **Actions → Wiki sync → Run workflow**.

## 1. Turn Wikis on

1. Open **[Repository settings → General → Features](https://github.com/kennetholsenatm-gif/OmniGraph/settings)**  
2. Enable **Wikis**.  
3. Decide who can edit (same page / **Wiki** section as needed).

Official reference: [About wikis](https://docs.github.com/en/communities/documenting-your-project-with-wikis/about-wikis).

## 2. Clone the wiki repository

Until the first push, the wiki repo may not exist yet; creating the first page in the GitHub UI, or an initial push, initializes it.

```bash
git clone https://github.com/kennetholsenatm-gif/OmniGraph.wiki.git
cd OmniGraph.wiki
```

Use SSH if that is how you authenticate to GitHub.

## 3. Copy markdown from `main`

From a checkout of **OmniGraph** (`main`):

```bash
# example: OmniGraph = main repo, OmniGraph.wiki = wiki clone (sibling dirs)
cp ../OmniGraph/wiki/*.md .
```

Review diffs. Commit and push:

```bash
git add .
git status
git commit -m "docs: sync wiki from main"
git push
```

Repeat whenever you change `wiki/` on `main` and want the Wiki tab updated without waiting for CI (or if CI is not configured).

## 4. Links and `_Sidebar.md`

- **Home** uses full `github.com/.../blob/main/...` links so pages still resolve from the Wiki repo (the wiki clone does not contain `docs/`).
- **`_Sidebar.md`** in this folder becomes the Wiki sidebar after you copy it across.

## 5. Keeping docs canonical

Long-form documentation belongs in **`docs/`** on `main`. Use **`wiki/`** for short navigation, onboarding, and pointers—then sync with the steps above when the GitHub Wiki should match.
