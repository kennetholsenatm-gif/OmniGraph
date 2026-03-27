# Phase 4: Data Fabric — NiFi and Kafka to Postgres (pgvector)

This document defines how **NiFi** and **Kafka** connect to the shared **Postgres (pgvector)** database in the messaging stack to log pipeline events. All services use the same `msg_backbone_net`; Postgres is the single sink for pipeline event logging.

## Overview

| Component    | Role |
|-------------|------|
| **devsecops-postgres** | pgvector/pg16; hostname `postgres`, database/user from `POSTGRES_DB` / `POSTGRES_USER`. Holds `pipeline.pipeline_events` for event log. |
| **NiFi**     | Writes pipeline events (e.g. flow completion, data lineage) via DBCP connection and PutDatabaseRecord or ExecuteSQL. |
| **Kafka**    | Events are sunk to Postgres via **Kafka Connect JDBC Sink** connector (topic → `pipeline.pipeline_events`). |

## Postgres

- **Host:** `postgres` (Docker service name on `msg_backbone_net`).
- **Port:** `5432`.
- **Database:** `POSTGRES_DB` (default `qminiwasm_db`).
- **User / Password:** `POSTGRES_USER` / `POSTGRES_PASSWORD` (from compose env; same as in [docker-compose.messaging.yml](../docker-compose/docker-compose.messaging.yml)).
- **Schema/table:** `pipeline.pipeline_events` (see [init-pipeline-events.sql](../docker-compose/init-scripts/init-pipeline-events.sql)).

### Table: `pipeline.pipeline_events`

| Column      | Type         | Description |
|------------|--------------|-------------|
| `id`       | BIGSERIAL    | Primary key. |
| `event_time` | TIMESTAMPTZ | When the event occurred. |
| `source`   | VARCHAR(64)  | `nifi`, `kafka`, or `n8n`. |
| `topic`    | VARCHAR(255) | Kafka topic or NiFi flow/processor name. |
| `event_type` | VARCHAR(128) | e.g. `flow_file_received`, `vuln_scan_completed`. |
| `payload`  | JSONB        | Event body. |
| `embedding`| vector(384)  | Optional pgvector for similarity search. |
| `created_at` | TIMESTAMPTZ | Insert time. |

Initialize the schema once (if not using initdb):

```bash
docker exec -i devsecops-postgres psql -U "${POSTGRES_USER:-qminiwasm}" -d "${POSTGRES_DB:-qminiwasm_db}" < docker-compose/init-scripts/init-pipeline-events.sql
```

Or mount the script into Postgres so it runs on first start (see [docker-compose.messaging.yml](../docker-compose/docker-compose.messaging.yml) comment for init volume).

---

## NiFi → Postgres

NiFi runs in the same compose as Postgres and connects over `msg_backbone_net`.

### Connection details (in NiFi)

- **JDBC URL:** `jdbc:postgresql://postgres:5432/${POSTGRES_DB:-qminiwasm_db}`  
  (NiFi receives `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB` from env; build URL as `jdbc:postgresql://${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}`.)
- **Driver:** PostgreSQL (ensure NiFi has the PostgreSQL JDBC driver in its lib or use the NiFi PostgreSQL bundle).
- **User / Password:** From env `POSTGRES_USER`, `POSTGRES_PASSWORD` (injected by Ansible or compose).

### NiFi configuration

1. **DBCP Connection Pool (Controller Service)**  
   - Database Connection URL: `jdbc:postgresql://postgres:5432/qminiwasm_db` (or expression using `POSTGRES_*` if NiFi supports env in controller services).  
   - Driver Class Name: `org.postgresql.Driver`.  
   - User / Password: set from your secrets (Vault → env → NiFi).

2. **Logging pipeline events**  
   - Use **PutDatabaseRecord** (with Record Reader and Record Writer) to insert into `pipeline.pipeline_events`, or  
   - Use **ExecuteSQL** with a parameterized INSERT, e.g.:

     ```sql
     INSERT INTO pipeline.pipeline_events (event_time, source, topic, event_type, payload)
     VALUES (?, 'nifi', ?, ?, ?::jsonb);
     ```  
   - Map flow file attributes or content to `topic` (e.g. flow name), `event_type`, and `payload` (JSON string).

3. **Environment variables for NiFi** (set in [docker-compose.messaging.yml](../docker-compose/docker-compose.messaging.yml)):

   - `POSTGRES_HOST=postgres`  
   - `POSTGRES_PORT=5432`  
   - `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` (same as Postgres service).

NiFi does not expose env to DBCP directly; use the same values when configuring the controller service (or a parameterized context so operators can plug in from Vault).

---

## Kafka → Postgres

Kafka does not connect to Postgres directly. Use **Kafka Connect** with the **JDBC Sink** connector to consume from a topic and write rows into `pipeline.pipeline_events`.

### Kafka Connect (optional service)

The repo defines an optional **kafka-connect** service in the messaging compose (or in a snippet) that:

- Uses the Confluent Kafka Connect image with the JDBC plugin.
- Connects to **Kafka** at `kafka:9092` and **Postgres** at `postgres:5432`.
- Runs a **JDBC Sink** connector that reads from a topic (e.g. `devsecops.pipeline.events`) and inserts into `pipeline.pipeline_events`.

### JDBC Sink connector (example)

Connector config is in [docker-compose/kafka-connect/jdbc-sink-pipeline-events.json](../docker-compose/kafka-connect/jdbc-sink-pipeline-events.json). Register it by POSTing to the Kafka Connect REST API (e.g. `http://kafka-connect:8083/connectors`). Replace `connection.password` with the actual `POSTGRES_PASSWORD` before sending; Kafka Connect does not expand env vars in connector config by default.

```json
{
  "name": "jdbc-sink-pipeline-events",
  "config": {
    "connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
    "tasks.max": "1",
    "topics": "devsecops.pipeline.events",
    "connection.url": "jdbc:postgresql://postgres:5432/qminiwasm_db",
    "connection.user": "qminiwasm",
    "connection.password": "${POSTGRES_PASSWORD}",
    "auto.create": "false",
    "insert.mode": "insert",
    "pk.mode": "none",
    "table.name.format": "pipeline.pipeline_events",
    "fields.whitelist": "event_time,source,topic,event_type,payload,created_at",
    "key.converter": "org.apache.kafka.connect.storage.StringConverter",
    "value.converter": "org.apache.kafka.connect.json.JsonConverter",
    "value.converter.schemas.enable": "false"
  }
}
```

Kafka message value should be a JSON object with fields aligned to the table (e.g. `event_time`, `source`, `topic`, `event_type`, `payload`). `created_at` can be defaulted in the DB or sent in the value. The connector maps value fields to columns by name.

### Alternative: consumer application

Instead of Kafka Connect, a small service or n8n workflow can consume from the topic and INSERT into `pipeline.pipeline_events` via Postgres (e.g. HTTP Request to a sidecar or Execute SQL). The schema and table remain the same; only the writer changes.

---

## Security and Varlock

- **Credentials:** Do not store `POSTGRES_PASSWORD` (or Kafka Connect DB password) in workflow JSON or in repo. Inject from Vault (or Ansible Vault) into the compose/env; NiFi and Kafka Connect use the same `POSTGRES_*` values as the Postgres service where appropriate.
- **Network:** NiFi, Kafka, Kafka Connect, and Postgres are all on `msg_backbone_net`; no external exposure of Postgres required for this fabric.

---

## Artifacts

| Item | Location |
|------|----------|
| Postgres schema (pipeline_events + pgvector) | [docker-compose/init-scripts/init-pipeline-events.sql](../docker-compose/init-scripts/init-pipeline-events.sql) |
| Messaging compose (Postgres, NiFi, Kafka) | [docker-compose/docker-compose.messaging.yml](../docker-compose/docker-compose.messaging.yml) |
| Kafka Connect JDBC Sink example | [docker-compose/kafka-connect/jdbc-sink-pipeline-events.json](../docker-compose/kafka-connect/jdbc-sink-pipeline-events.json) |
| Env schema (POSTGRES_* for NiFi/Connect) | [devsecops.env.schema](../devsecops.env.schema) (Phase 4 block) |
