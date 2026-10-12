---
name: recon-mcp
description: Use the Recon MCP service to manage collection projects, scan tasks, templates, assets and scanner nodes.
---

# Recon MCP

Use this skill with a running Recon service. The console deployment connects to its private `/mcp` endpoint automatically using a signed gateway assertion. Gateway secrets and assertions belong on the server and must never be copied to browser settings, URLs or logs.

A standalone MCP client can use an API key created in the Recon account settings. The existing key format starts with `ssk_`:

```json
{
  "mcpServers": {
    "recon": {
      "url": "http://<host>:8082/mcp",
      "headers": { "X-API-Key": "ssk_<your-key>" }
    }
  }
}
```

Use the service's reachable endpoint if a reverse proxy changes the host or port. Restart or reload the MCP connection and verify that tools such as `list_projects` and `list_assets` appear. REST APIs continue to accept standalone login JWTs through `Authorization: Bearer <token>`.

## Tools

- Projects: `list_projects` groups by tag; `list_projects_data` returns paginated records; `get_project` reads details; `create_project` creates a project or replaces the scope of an existing name.
- Tasks: `list_tasks`, `get_task`, `create_scan_task`.
- Templates and plugins: `list_scan_templates`, `get_scan_template`, `list_plugin_modules`, `list_plugins`, `create_scan_template`.
- Assets: `list_assets`, `count_assets`, `get_asset_detail`, `add_asset_tag`.
- Nodes: `list_nodes`, optionally with `online_only: true`.

Read the tool schema and description before choosing parameters. They describe the exact stored-field mappings and supported filters.

## Project and scan workflow

Use `list_projects` or `list_projects_data` to obtain project ObjectIDs, then query assets with `filter.project`.

```json
{"asset_type":"subdomain","filter":{"project":["<project-object-id>"]},"pageIndex":1,"pageSize":20}
```

Before creating a scan, use `list_nodes` to choose online node names and `list_scan_templates` to select a template ObjectID. Task names must be unique. For direct targets, use `targetSource: "general"` and newline or comma-separated targets:

```json
{"name":"example-collection","node":["<node-name>"],"template":"<template-object-id>","targetSource":"general","target":"example.test"}
```

For project targets, use `targetSource: "project"` with project ObjectIDs. `ignore` excludes targets and `duplicates` selects the deduplication policy. `bindProject` assigns result ownership. Scheduling uses `scheduledTasks`, `cycleType`, `hour`, `minute`, `day` and `week`.

For complete root-domain collection, use two stages. First run only `SubdomainScan` and `SubdomainSecurity` against root domains. Wait for completion using `get_task`. Then create a task with `targetSource: "subdomain"` and `search: "task==\"<first-task-name>\""` for port, asset and vulnerability stages. Each discovered subdomain becomes an independent target, improving distribution across nodes.

Asset-derived targets support `asset`, `RootDomain`, `subdomain` and `UrlScan`. The corresponding `*Source` variants support `targetTp: "search"` with search/filter/project/targetNumber or `targetTp: "select"` with targetIds. A targetNumber of 0 means unlimited.

## Create a template

1. Use `list_plugin_modules` to obtain valid module names.
2. Use `list_plugins`, optionally filtered by module, to obtain plugin hashes and default parameters.
3. Pass `modules` as a map of module names to arrays of plugin hashes. Plugins execute in array order.
4. Use `parameters` to override defaults with module -> plugin hash -> parameter string.

Module fields use plugin hashes, not names or database IDs. Optional `vullist` contains nuclei POC template IDs. Advanced `template_json` takes precedence over modules. The returned template ID can be passed directly to `create_scan_task.template`.

Pipeline modules: TargetHandler, SubdomainScan, SubdomainSecurity, PortScanPreparation, PortScan, PortFingerprint, AssetMapping, AssetHandle, URLScan, WebCrawler, URLSecurity, DirScan, VulnerabilityScan and PassiveScan.

## Asset query contract

Supported asset types: asset, RootDomain, subdomain, app, mp, UrlScan, SensitiveResult, DirScanResult, crawler, vulnerability, PageMonitoring, IPAsset and SubdomainTakerResult. Common aliases include web, root_domain, miniprogram, url, sensitive, directory, vuln, page_monitoring, ip and subdomain_taker.

`list_assets` defaults to pageIndex 1 and pageSize 20. `count_assets` accepts the same asset_type/search/filter and returns total without pagination. search and filter can be combined.

Search is a custom DSL:

- `field=value`: case-insensitive regex matching.
- `field=="value"`: exact equality; quote values with spaces.
- `field!="value"`: exclusion.
- `&&`, `||`, and parentheses combine expressions.
- Common fields: tag, task (task name), rootDomain.

Examples: `domain==www.example.test && port==443`, `ip==192.0.2.1`, `task=="collection-task"`, `level==high`.

Project selection requires `filter.project` with ObjectIDs, never project names. `project` is not supported in search and can fail when combined with `&&`. Task search/filter values are task names from `list_tasks`, never task IDs.

Filter is an object containing arrays of strings. Values under one key are OR; separate keys are AND. Unknown keys are ignored. Supported keys depend on asset type; inspect the list_assets description. Common keys include project, task, port, service, app, icon, statuscode, status, level, type, color, tags and sname.

Prefer a known project filter when the asset type supports it to narrow the query. Exact equality is usually preferable to broad regex queries; index use also depends on the query and available database indexes.

DirScanResult searches HTTP status through `statuscode`, but filters through `status`. UrlScan has no statuscode search and filters HTTP status through `filter.status`. Sensitive rule names use `sname` in search or filter; `sid` expands sensitive match details.

Only UrlScan and DirScanResult support length sorting: `{"length":"ascending"}` means ascending; every other value means descending. Other asset types use the server's fixed time or ID ordering.

Sensitive and vulnerability workflow status values remain: 1 pending, 2 in progress, 3 ignored, 4 suspected, 5 confirmed, 6 resolved. Vulnerability levels: critical, high, medium, low, info and unknown.

Asset detail lookup supports `asset` using its ObjectID and `vulnerability` using its hash. `add_asset_tag` uses an asset ObjectID.

## Connection and deployment checks

If no tools appear, check the endpoint, authentication and service health. A 401 indicates a missing or invalid API key or gateway assertion. Use `list_nodes` before a scan; an available template and an online node are required to execute it. If queries are empty, verify the asset type and filter values before broadening the search.

The temporary display label can be changed with `RECON_DISPLAY_NAME`. Optional `RECON_VERSION_URL` and `RECON_PLUGIN_CATALOG_URL` enable a configured version feed and plugin catalog. Without those settings, the service reports local versions and lists its installed plugins. `RECON_STATE_DIR` selects the persistent configuration and artifact directory; standalone defaults use the executable directory.
