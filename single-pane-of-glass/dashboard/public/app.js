(function () {
  const wikiSidebar = document.getElementById("wiki-sidebar");
  const wikiContent = document.getElementById("wiki-content");
  const notebookFrame = document.getElementById("notebook-frame");
  const notebookPlaceholder = document.getElementById("notebook-placeholder");

  let config = { wikiOwner: null, wikiRepo: null, lmnotebookUrl: null };

  async function fetchConfig() {
    const r = await fetch("/api/config");
    config = await r.json();
  }

  async function fetchWikiPages() {
    if (!config.wikiOwner || !config.wikiRepo) return null;
    const q = new URLSearchParams({ owner: config.wikiOwner, repo: config.wikiRepo });
    const r = await fetch("/api/wiki?" + q);
    if (!r.ok) return null;
    const data = await r.json();
    return Array.isArray(data) ? data : (data.data || data.pages || []);
  }

  async function fetchWikiPage(slug) {
    if (!config.wikiOwner || !config.wikiRepo) return null;
    const r = await fetch("/api/wiki/" + encodeURIComponent(slug) + "?owner=" + encodeURIComponent(config.wikiOwner) + "&repo=" + encodeURIComponent(config.wikiRepo));
    if (!r.ok) return null;
    const data = await r.json();
    return (typeof data === "string" ? data : (data.content || data.body || data.data || "")) || "";
  }

  function renderMarkdown(text) {
    if (!text) return "";
    const h = document.createElement("div");
    h.innerHTML = text
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
    const raw = h.textContent || text;
    return raw
      .replace(/^### (.+)$/gm, "<h3>$1</h3>")
      .replace(/^## (.+)$/gm, "<h2>$1</h2>")
      .replace(/^# (.+)$/gm, "<h1>$1</h1>")
      .replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
      .replace(/\*(.+?)\*/g, "<em>$1</em>")
      .replace(/`([^`]+)`/g, "<code>$1</code>")
      .replace(/```[\s\S]*?```/g, (m) => "<pre>" + m.replace(/^```\w*\n?|```$/g, "").replace(/</g, "&lt;").replace(/>/g, "&gt;") + "</pre>")
      .replace(/\n\n/g, "</p><p>")
      .replace(/\n/g, "<br/>");
  }

  function showWikiList() {
    wikiSidebar.innerHTML = "";
    wikiContent.innerHTML = "<p>Loading…</p>";
    fetchWikiPages().then(function (pages) {
      if (!pages || !Array.isArray(pages)) {
        wikiContent.innerHTML = "<p>No wiki configured or API error. Set GITEA_WIKI_OWNER and GITEA_WIKI_REPO.</p>";
        return;
      }
      wikiSidebar.innerHTML = pages
        .map(function (p) {
          const slug = p.slug || p.title || "";
          return '<a href="#' + slug + '" data-slug="' + slug + '">' + (p.title || slug) + "</a>";
        })
        .join("");
      wikiSidebar.querySelectorAll("a").forEach(function (a) {
        a.addEventListener("click", function (e) {
          e.preventDefault();
          const slug = a.getAttribute("data-slug");
          wikiSidebar.querySelectorAll("a").forEach(function (x) { x.classList.remove("active"); });
          a.classList.add("active");
          wikiContent.innerHTML = "<p>Loading…</p>";
          fetchWikiPage(slug).then(function (content) {
            wikiContent.innerHTML = "<p>" + renderMarkdown(content) + "</p>";
          });
        });
      });
      if (pages.length) {
        const first = wikiSidebar.querySelector("a");
        if (first) first.click();
      } else {
        wikiContent.innerHTML = "<p>No wiki pages yet.</p>";
      }
    });
  }

  function setupNotebook() {
    if (config.lmnotebookUrl) {
      notebookFrame.src = config.lmnotebookUrl;
      notebookPlaceholder.style.display = "none";
    } else {
      notebookPlaceholder.style.display = "block";
    }
  }

  function setupSSE() {
    const es = new EventSource("/api/events");
    es.onmessage = function (e) {
      if (e.data === "refresh") showWikiList();
    };
    es.onerror = function () { setTimeout(setupSSE, 5000); };
  }

  (async function init() {
    await fetchConfig();
    showWikiList();
    setupNotebook();
    setupSSE();
  })();
})();
