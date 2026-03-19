/**
 * Single Pane of Glass — Dashboard server.
 * Serves the unified UI, proxies Gitea API for wiki/docs, exposes SSE for event-driven refresh.
 * No secrets in code; GITEA_API_TOKEN, REFRESH_SECRET from env (Varlock).
 */
const http = require("http");
const fs = require("fs");
const path = require("path");
const url = require("url");

const PORT = parseInt(process.env.PORT || "3000", 10);
const GITEA_URL = (process.env.GITEA_URL || "http://gitea:3000").replace(/\/$/, "");
const GITEA_API_TOKEN = process.env.GITEA_API_TOKEN || "";
const GITEA_WIKI_OWNER = process.env.GITEA_WIKI_OWNER || "";
const GITEA_WIKI_REPO = process.env.GITEA_WIKI_REPO || "";
const LMNOTEBOOK_URL = process.env.LMNOTEBOOK_URL || "";
const REFRESH_SECRET = process.env.REFRESH_SECRET || "";

const clients = new Set();

function sendRefreshToAll() {
  clients.forEach((res) => {
    try {
      res.write("data: refresh\n\n");
    } catch (_) {}
  });
}

const server = http.createServer(async (req, res) => {
  const parsed = url.parse(req.url, true);
  const pathname = parsed.pathname;

  // POST /api/refresh — trigger from webhook-listener (internal); optional secret
  if (pathname === "/api/refresh" && req.method === "POST") {
    const auth = req.headers["authorization"];
    const secret = auth && auth.startsWith("Bearer ") ? auth.slice(7) : "";
    if (REFRESH_SECRET && secret !== REFRESH_SECRET) {
      res.writeHead(401, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ error: "Unauthorized" }));
      return;
    }
    sendRefreshToAll();
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ ok: true }));
    return;
  }

  // GET /api/wiki — list wiki pages or get page content from Gitea
  if (pathname === "/api/wiki" || pathname.startsWith("/api/wiki/")) {
    const token = GITEA_API_TOKEN;
    const owner = GITEA_WIKI_OWNER || parsed.query.owner;
    const repo = GITEA_WIKI_REPO || parsed.query.repo;
    if (!owner || !repo) {
      res.writeHead(400, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ error: "GITEA_WIKI_OWNER and GITEA_WIKI_REPO (or query params) required" }));
      return;
    }
    const slug = pathname === "/api/wiki" ? null : pathname.slice("/api/wiki/".length);
    const giteaPath = slug
      ? `/api/v1/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/wiki/page/${encodeURIComponent(slug)}`
      : `/api/v1/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/wiki/pages`;
    const u = url.parse(GITEA_URL);
    const opts = {
      hostname: u.hostname,
      port: u.port || (u.protocol === "https:" ? 443 : 80),
      path: giteaPath,
      method: "GET",
      headers: token ? { Authorization: `token ${token}` } : {},
    };
    const proto = u.protocol === "https:" ? require("https") : require("http");
    const proxyReq = proto.request(opts, (proxyRes) => {
      res.writeHead(proxyRes.statusCode, { "Content-Type": proxyRes.headers["content-type"] || "application/json" });
      proxyRes.pipe(res);
    });
    proxyReq.on("error", (e) => {
      res.writeHead(502, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ error: "Bad Gateway", message: e.message }));
    });
    proxyReq.end();
    return;
  }

  // GET /api/config — safe config for frontend (no secrets)
  if (pathname === "/api/config") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(
      JSON.stringify({
        lmnotebookUrl: LMNOTEBOOK_URL || null,
        wikiOwner: GITEA_WIKI_OWNER || null,
        wikiRepo: GITEA_WIKI_REPO || null,
      })
    );
    return;
  }

  // GET /api/events — SSE for refresh events
  if (pathname === "/api/events") {
    res.writeHead(200, {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      Connection: "keep-alive",
    });
    res.write("data: connected\n\n");
    clients.add(res);
    req.on("close", () => clients.delete(res));
    return;
  }

  // Static: index.html or 404
  const file = pathname === "/" || pathname === "" ? "/index.html" : pathname;
  const filePath = path.join(__dirname, "public", file === "/index.html" ? "index.html" : file.slice(1));
  if (!filePath.startsWith(path.join(__dirname, "public"))) {
    res.writeHead(404);
    res.end();
    return;
  }
  fs.readFile(filePath, (err, data) => {
    if (err) {
      res.writeHead(404);
      res.end();
      return;
    }
    const ext = path.extname(filePath);
    const types = { ".html": "text/html", ".js": "application/javascript", ".css": "text/css", ".ico": "image/x-icon" };
    res.writeHead(200, { "Content-Type": types[ext] || "application/octet-stream" });
    res.end(data);
  });
});

server.listen(PORT, "0.0.0.0", () => {
  console.log("Dashboard listening on", PORT);
});
