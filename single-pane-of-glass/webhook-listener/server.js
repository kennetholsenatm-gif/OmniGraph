/**
 * Webhook listener for SAM / n8n doc-generation events and Gitea-triggered docs sync.
 * On POST /webhook/doc-push: optionally verify HMAC, then trigger dashboard refresh.
 * On POST /webhook/docs-sync: verify Gitea HMAC, then start docs-sync container via Docker Engine API (unix socket).
 * No secrets in code; env from Varlock / compose.
 */
const http = require("http");
const crypto = require("crypto");
const { URL } = require("url");

const PORT = parseInt(process.env.PORT || "4000", 10);
const DASHBOARD_REFRESH_URL = process.env.DASHBOARD_REFRESH_URL || "http://dashboard:3000/api/refresh";
const GATEWAY_REFRESH_SECRET = process.env.GATEWAY_REFRESH_SECRET || "";
const WEBHOOK_HMAC_SECRET = process.env.WEBHOOK_HMAC_SECRET || "";

const DOCS_SYNC_ENABLED = process.env.DOCS_SYNC_ENABLED === "true";
const DOCS_SYNC_CONTAINER_NAME = process.env.DOCS_SYNC_CONTAINER_NAME || "devsecops-docs-sync";
const GITEA_DOCS_WEBHOOK_SECRET = process.env.GITEA_DOCS_WEBHOOK_SECRET || "";
const DOCS_SYNC_REPO_FULL_NAME = (process.env.DOCS_SYNC_REPO_FULL_NAME || "").trim();
const DOCKER_SOCK = process.env.DOCKER_SOCK || "/var/run/docker.sock";

function verifyHmac(body, signature) {
  if (!WEBHOOK_HMAC_SECRET) return true;
  if (!signature) return false;
  const alg = signature.startsWith("sha256=") ? "sha256" : "sha256";
  const expected = crypto.createHmac(alg, WEBHOOK_HMAC_SECRET).update(body).digest("hex");
  const received = signature.replace(/^sha256=/, "").trim();
  return crypto.timingSafeEqual(Buffer.from(expected, "hex"), Buffer.from(received, "hex"));
}

/** Gitea sends X-Gitea-Signature (hex HMAC-SHA256 of raw body). Some setups use X-Hub-Signature-256 (sha256=<hex>). */
function verifyGiteaDocsSignature(body, req) {
  if (!GITEA_DOCS_WEBHOOK_SECRET) return false;
  const giteaSig = req.headers["x-gitea-signature"] || req.headers["x-gitea-signature-256"];
  const hubSig = req.headers["x-hub-signature-256"];
  const headerVal = giteaSig || hubSig;
  if (!headerVal) return false;
  let hex = String(headerVal).trim();
  if (hex.toLowerCase().startsWith("sha256=")) {
    hex = hex.slice(7).trim();
  }
  const expected = crypto.createHmac("sha256", GITEA_DOCS_WEBHOOK_SECRET).update(body).digest("hex");
  try {
    const a = Buffer.from(expected, "hex");
    const b = Buffer.from(hex, "hex");
    if (a.length !== b.length || a.length === 0) return false;
    return crypto.timingSafeEqual(a, b);
  } catch {
    return false;
  }
}

function repoMatchesFilter(bodyStr) {
  if (!DOCS_SYNC_REPO_FULL_NAME) return true;
  try {
    const payload = JSON.parse(bodyStr);
    const full = payload.repository && payload.repository.full_name;
    if (!full) return false;
    return String(full).toLowerCase() === DOCS_SYNC_REPO_FULL_NAME.toLowerCase();
  } catch {
    return false;
  }
}

function dockerContainerStart(containerName, callback) {
  const req = http.request(
    {
      socketPath: DOCKER_SOCK,
      path: `/v1.41/containers/${encodeURIComponent(containerName)}/start`,
      method: "POST",
      headers: { "Content-Length": "0" },
    },
    (res) => {
      const chunks = [];
      res.on("data", (c) => chunks.push(c));
      res.on("end", () => {
        const txt = Buffer.concat(chunks).toString("utf8");
        if (res.statusCode === 204 || res.statusCode === 304) {
          callback(null, { statusCode: res.statusCode });
        } else {
          callback(
            new Error(`Docker API HTTP ${res.statusCode}: ${txt || res.statusMessage || "unknown"}`)
          );
        }
      });
    }
  );
  req.on("error", callback);
  req.end();
}

function handleDocPush(req, res, body) {
  const sig = req.headers["x-hub-signature-256"] || req.headers["x-signature"] || "";
  if (!verifyHmac(body, sig)) {
    res.writeHead(401, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ error: "Invalid signature" }));
    return;
  }

  const opts = new URL(DASHBOARD_REFRESH_URL);
  const requestOpts = {
    hostname: opts.hostname,
    port: opts.port || 80,
    path: opts.pathname + (opts.search || ""),
    method: "POST",
    headers: { "Content-Type": "application/json" },
  };
  if (GATEWAY_REFRESH_SECRET) {
    requestOpts.headers.Authorization = "Bearer " + GATEWAY_REFRESH_SECRET;
  }
  const proto = opts.protocol === "https:" ? require("https") : http;
  const refreshReq = proto.request(requestOpts, (refreshRes) => {
    res.writeHead(refreshRes.statusCode, { "Content-Type": "application/json" });
    if (refreshRes.statusCode === 200) {
      res.end(JSON.stringify({ ok: true, refreshed: true }));
    } else {
      res.end(JSON.stringify({ ok: false, dashboardStatus: refreshRes.statusCode }));
    }
  });
  refreshReq.on("error", (e) => {
    res.writeHead(502, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ error: "Dashboard refresh failed", message: e.message }));
  });
  refreshReq.end();
}

function handleDocsSync(req, res, body) {
  if (!DOCS_SYNC_ENABLED) {
    res.writeHead(503, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ error: "Docs sync disabled", hint: "Set DOCS_SYNC_ENABLED=true" }));
    return;
  }
  if (!GITEA_DOCS_WEBHOOK_SECRET) {
    res.writeHead(503, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ error: "GITEA_DOCS_WEBHOOK_SECRET is not configured" }));
    return;
  }
  if (!verifyGiteaDocsSignature(body, req)) {
    res.writeHead(401, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ error: "Invalid Gitea webhook signature" }));
    return;
  }
  if (!repoMatchesFilter(body)) {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ ok: true, skipped: true, reason: "repository does not match DOCS_SYNC_REPO_FULL_NAME" }));
    return;
  }

  dockerContainerStart(DOCS_SYNC_CONTAINER_NAME, (err, result) => {
    if (err) {
      res.writeHead(502, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ error: "Docker start failed", message: err.message }));
      return;
    }
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(
      JSON.stringify({
        ok: true,
        container: DOCS_SYNC_CONTAINER_NAME,
        dockerStatus: result.statusCode,
      })
    );
  });
}

const server = http.createServer((req, res) => {
  if (req.method === "GET" && req.url === "/") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ service: "gateway-webhook-listener", status: "ok" }));
    return;
  }

  if (req.method === "POST" && req.url === "/webhook/doc-push") {
    const chunks = [];
    req.on("data", (c) => chunks.push(c));
    req.on("end", () => {
      const body = Buffer.concat(chunks).toString("utf8");
      handleDocPush(req, res, body);
    });
    return;
  }

  if (req.method === "POST" && req.url === "/webhook/docs-sync") {
    const chunks = [];
    req.on("data", (c) => chunks.push(c));
    req.on("end", () => {
      const body = Buffer.concat(chunks).toString("utf8");
      handleDocsSync(req, res, body);
    });
    return;
  }

  res.writeHead(404, { "Content-Type": "application/json" });
  res.end(JSON.stringify({ error: "Not Found" }));
});

server.listen(PORT, "0.0.0.0", () => {
  console.log("Webhook listener on", PORT);
});
