package mcp

const listAssetsToolDesc = `Query Recon assets.

Request parameters match the asset POST APIs (for example /api/assets/asset) and models.SearchRequest:
- pageIndex/pageSize: pagination, starting at 1
- search: custom search expression, not SQL
- filter: exact filters combined with search
- sort: supported only for selected types
- sid: sensitive rule name, only for SensitiveResult

asset_type maps to the backend Index and helper.GetSearchQuery builds the MongoDB query.
Supported types: asset, RootDomain, subdomain, app, mp, UrlScan, SensitiveResult,
DirScanResult, crawler, vulnerability, PageMonitoring, IPAsset, SubdomainTakerResult.

Search syntax:
- field=value: case-insensitive regex match
- field=="value": exact match; quote values containing spaces
- field!="value": exclude matching values
- expr1 && expr2: AND; expr1 || expr2: OR; (expr): grouping
- Common fields: tag -> tags, task -> taskName (task name), rootDomain
- project is not a search field. Use filter.project with project ObjectIDs from list_projects or list_projects_data.

Search fields by asset type (arrows indicate stored field names):
- asset: domain->host, ip, port, service, app->technologies, title, statuscode, icon->faviconmmh3, banner->metadata, type, body, header->rawheaders
- RootDomain: domain, icp, company
- subdomain: domain->host, ip, type, value
- app: name, icp, company, category, description, url, apk
- mp: name, icp, company, category, description, url
- UrlScan: url->output, input, source, resultId, type->outputtype. No statuscode search; use filter.status for HTTP status.
- SensitiveResult: url, sname->sid, body, info->match, md5
- DirScanResult: url, statuscode->status, redirect->msg, length
- vulnerability: url, vulname, matched, request, response, level
- crawler: url, method, body, resultId
- PageMonitoring: url, hash, diff, response
- IPAsset: ip, domain->ports.server.domain, port->ports.port, service->ports.server.service, webServer->ports.server.webServer, app->ports.server.technologies
- SubdomainTakerResult: domain->input, value, type->cname, response

Examples: domain=example && port==443; port==443 && service=nginx; title="admin login" || body=admin.

Exact filter is a JSON object with arrays of strings. Values for one key are OR; different keys are AND.
Only keys defined in filterKeyCache are applied; unknown keys are ignored.
Global filter mappings: project->project, port->port, service->service, app->technologies,
icon->faviconmmh3, statuscode->statuscode, status->status, level->level, type->type,
color->color, tags->tags, task->taskName, sname->sid.

Filters by asset type (other keys may not apply to that collection):
- asset: project, port, service, app, icon, statuscode, type, task, tags
- RootDomain: project, tags
- subdomain: project, type, task, tags
- app/mp: project, tags
- UrlScan: status (HTTP status, not statuscode), tags
- SensitiveResult: status (1 pending, 2 in progress, 3 ignored, 4 suspected, 5 confirmed, 6 resolved), color, sname, tags
- DirScanResult: status (HTTP status), tags
- crawler: project, task, tags
- vulnerability: project, level (critical/high/medium/low/info/unknown), status (1-6), task, tags
- PageMonitoring: tags
- IPAsset: project, port, service, app, using aggregation of nested ports fields
- SubdomainTakerResult: tags

Project filter values are ObjectIDs, never project names. Task values are names from list_tasks, never task IDs.
Examples: {"project":["<project-object-id>"],"port":["443"]}; {"level":["high"]}.
search and filter can be combined: search=domain=example && port==443, filter={"project":["<project-object-id>"]}.
Do not use project=... in search; combining that with && can fail.
DirScanResult uses statuscode in search but status in filter. UrlScan has no statuscode search.
Sensitive rule names use sname in search or filter.sname.

Sort: UrlScan/DirScanResult accept {"length":"ascending"}; all other values, including descending and -1, mean descending.
Other types have a fixed time or _id order and generally ignore sort.
sid expands match details for a named sensitive rule, only for SensitiveResult.`

const createScanTemplateToolDesc = `Create a scan template composed of pipeline modules and plugins.
Module fields contain plugin hashes, not plugin IDs or names. Plugins execute in array order.
1. Use list_plugin_modules to discover module names.
2. Use list_plugins, optionally filtered by module, to obtain plugin hashes and default parameters.
3. Supply modules as {"SubdomainScan":["<plugin-hash>"],"PortScan":["<plugin-hash>"]}.
4. Default plugin parameters are filled automatically; parameters overrides them.

name is required. parameters has the shape module -> plugin hash -> parameter string.
vullist is an optional list of nuclei POC template IDs, used only with nuclei in VulnerabilityScan.
template_json is an optional complete ScanTemplate JSON and takes precedence over modules.
The returned template ID can be passed to create_scan_task.template.
Modules: TargetHandler, SubdomainScan, SubdomainSecurity, PortScanPreparation, PortScan,
PortFingerprint, AssetMapping, AssetHandle, URLScan, WebCrawler, URLSecurity, DirScan,
VulnerabilityScan, PassiveScan.`

const createScanTaskToolDesc = `Create a scan task. name and node are required.
Use a scan template ObjectID from list_scan_templates as template to execute scanning.

Target source controls target resolution:
- general: use target with newline or comma-separated domains, IP addresses or URLs
- project: read targets from project; provide project ObjectIDs, without target
- asset, RootDomain, subdomain, UrlScan: search the corresponding asset collection;
  provide search and optionally project, filter and targetNumber
- assetSource, RootDomainSource, subdomainSource, UrlScanSource: create from an asset page
  - targetTp=search: select using search, filter, project and targetNumber
  - targetTp=select: select using targetIds (asset ObjectIDs)

Search uses the list_assets DSL. Resume from collected subdomains with targetSource=subdomain
and search=task=="<subdomain-task-name>".
filter uses exact values, and filter.project contains project ObjectIDs.
targetNumber limits search results; 0 means unlimited. targetIds lists selected asset ObjectIDs.
allNode includes all online nodes. ignore excludes targets; duplicates selects a deduplication policy.
bindProject assigns result ownership. scheduledTasks and cycleType/hour/minute/day/week configure scheduling.

For a complete root-domain scan, first use general with only SubdomainScan/SubdomainSecurity.
After completion, use subdomain with search=task=="<first-task-name>" for subsequent stages.
This distributes each discovered subdomain as an independent target across scanner nodes.`

const countAssetsToolDesc = `Count matching assets without pagination using /api/assets/common/total.
asset_type, search and filter have the same values and syntax as list_assets.
Returns total. Example project subdomains:
{"asset_type":"subdomain","filter":{"project":["<project-object-id>"]}}
Example assets from a task: {"asset_type":"asset","search":"task==\"<task-name>\""}.`
