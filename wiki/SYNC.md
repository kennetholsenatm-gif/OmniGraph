# Publishing `wiki/` to GitHub Wiki

GitHub Wikis are stored in a **second** repository: `https://github.com/kennetholsenatm-gif/OmniGraph.wiki.git`  
The [`wiki/` folder on `main`](https://github.com/kennetholsenatm-gif/OmniGraph/tree/main/wiki) is the copy we keep next to the code. Nothing syncs automatically; you push updates when you want the Wiki tab to change.

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

Repeat whenever you change `wiki/` on `main` and want the Wiki tab updated.

## 4. Links and `_Sidebar.md`

- **Home** uses full `github.com/.../blob/main/...` links so pages still resolve from the Wiki repo (the wiki clone does not contain `docs/`).
- **`_Sidebar.md`** in this folder becomes the Wiki sidebar after you copy it across.

## 5. Keeping docs canonical

Long-form documentation belongs in **`docs/`** on `main`. Use **`wiki/`** for short navigation, onboarding, and pointers—then sync with the steps above when the GitHub Wiki should match.
