package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"recon/internal/constants"
	"recon/internal/models"
	"recon/internal/services/assets/app"
	"recon/internal/services/assets/asset"
	assetCommon "recon/internal/services/assets/common"
	"recon/internal/services/assets/crawler"
	"recon/internal/services/assets/dirscan"
	"recon/internal/services/assets/ip"
	"recon/internal/services/assets/mp"
	"recon/internal/services/assets/page_monitoring"
	"recon/internal/services/assets/root_domain"
	"recon/internal/services/assets/sensitive"
	"recon/internal/services/assets/subdomain"
	"recon/internal/services/assets/url"
	"recon/internal/services/assets/vulnerability"
	"recon/internal/services/node"
	"recon/internal/services/plugin"
	"recon/internal/services/project"
	taskCommon "recon/internal/services/task/common"
	"recon/internal/services/task/task"
	"recon/internal/services/task/template"
)

type deps struct {
	projectService    project.Service
	taskService       task.Service
	taskCommonService taskCommon.Service
	templateService   template.Service
	assetService      asset.Service
	rootDomainService root_domain.Service
	subdomainService  subdomain.Service
	appService        app.Service
	mpService         mp.Service
	urlService        url.Service
	sensitiveService  sensitive.Service
	dirscanService    dirscan.Service
	crawlerService    crawler.Service
	vulnService       vulnerability.Service
	ipService         ip.Service
	pageMonService    page_monitoring.Service
	commonService     assetCommon.Service
	nodeService       node.Service
	pluginService     plugin.Service
}

var d = &deps{
	projectService:    project.NewService(),
	taskService:       task.NewService(),
	taskCommonService: taskCommon.NewService(),
	templateService:   template.NewService(),
	assetService:      asset.NewService(),
	rootDomainService: root_domain.NewService(),
	subdomainService:  subdomain.NewService(),
	appService:        app.NewService(),
	mpService:         mp.NewService(),
	urlService:        url.NewService(),
	sensitiveService:  sensitive.NewService(),
	dirscanService:    dirscan.NewService(),
	crawlerService:    crawler.NewService(),
	vulnService:       vulnerability.NewService(),
	ipService:         ip.NewService(),
	pageMonService:    page_monitoring.NewService(),
	commonService:     assetCommon.NewService(),
	nodeService:       node.NewService(),
	pluginService:     plugin.NewService(),
}

func registerTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects",
		Description: "List projects grouped by tag",
	}, listProjects)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects_data",
		Description: "List projects with pagination and name search",
	}, listProjectsData)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_project",
		Description: "Get project details by ID",
	}, getProject)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_project",
		Description: "Create a project. Requires name and a valid target; tag, template and node are optional. If the name exists and normalized targets remain, replace only its asset scope.",
	}, createProject)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tasks",
		Description: "List scan tasks with pagination",
	}, listTasks)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task",
		Description: "Get scan task details by ID",
	}, getTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_scan_templates",
		Description: "List scan templates with pagination",
	}, listScanTemplates)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_scan_template",
		Description: "Get scan template details by ID",
	}, getScanTemplate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_plugin_modules",
		Description: "List all scan template modules (pipeline stages). Each module contains plugins.",
	}, listPluginModules)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_plugins",
		Description: "List scanning plugins, optionally filtered by module. Returns hash, name, module and default parameter. Template module fields reference plugin hashes.",
	}, listPlugins)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_scan_template",
		Description: createScanTemplateToolDesc,
	}, createScanTemplate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_scan_task",
		Description: createScanTaskToolDesc,
	}, createScanTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_assets",
		Description: listAssetsToolDesc,
	}, listAssets)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "count_assets",
		Description: countAssetsToolDesc,
	}, countAssets)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_asset_detail",
		Description: "Get asset details. Supported asset_type values: asset, vulnerability.",
	}, getAssetDetail)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_asset_tag",
		Description: "Add an asset tag",
	}, addAssetTag)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_nodes",
		Description: "List scanner nodes; online_only=true returns only online nodes",
	}, listNodes)
}

type listProjectsDataInput struct {
	Search    string `json:"search,omitempty" jsonschema:"Partial project name search"`
	PageIndex int    `json:"pageIndex,omitempty" jsonschema:"Page number, starting at 1; default 1"`
	PageSize  int    `json:"pageSize,omitempty" jsonschema:"Results per page; default 20"`
}

func listProjects(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	c := ginContext(ctx)
	result, err := d.projectService.GetProjectsByTag(c)
	if err != nil {
		return errorResult("Failed to list projects", err)
	}
	return jsonToolResult(ginH{"list": result})
}

func listProjectsData(ctx context.Context, _ *mcp.CallToolRequest, input listProjectsDataInput) (*mcp.CallToolResult, any, error) {
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	c := ginContext(ctx)
	result, err := d.projectService.GetProjectsData(c, input.Search, input.PageIndex, input.PageSize)
	if err != nil {
		return errorResult("Failed to retrieve project data", err)
	}
	return jsonToolResult(result)
}

type getProjectInput struct {
	ID string `json:"id" jsonschema:"Project MongoDB ObjectID"`
}

func getProject(ctx context.Context, _ *mcp.CallToolRequest, input getProjectInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id is required", nil)
	}
	c := ginContext(ctx)
	result, err := d.projectService.GetProjectContent(c, input.ID)
	if err != nil {
		return errorResult("Failed to retrieve project details", err)
	}
	if result == nil {
		return errorResult("Project not found", nil)
	}
	return jsonToolResult(result)
}

type createProjectInput struct {
	Name           string   `json:"name" jsonschema:"Project name, required"`
	Tag            string   `json:"tag,omitempty" jsonschema:"Project tag for grouping"`
	Target         string   `json:"target" jsonschema:"Scan targets, required. Newline or comma-separated domains, IP addresses or URLs."`
	Template       string   `json:"template,omitempty" jsonschema:"Associated scan template ID"`
	Node           []string `json:"node,omitempty" jsonschema:"Scanner node names"`
	AllNode        bool     `json:"allNode,omitempty" jsonschema:"Use all nodes"`
	Ignore         string   `json:"ignore,omitempty" jsonschema:"Ignored targets, in the same format as target"`
	Duplicates     string   `json:"duplicates,omitempty" jsonschema:"Deduplication policy"`
	ScheduledTasks bool     `json:"scheduledTasks,omitempty" jsonschema:"Enable scheduled scans"`
	Hour           int      `json:"hour,omitempty" jsonschema:"Scan interval in hours; used only when scheduledTasks is enabled"`
}

func createProject(ctx context.Context, _ *mcp.CallToolRequest, input createProjectInput) (*mcp.CallToolResult, any, error) {
	if input.Name == "" || strings.TrimSpace(input.Target) == "" {
		return errorResult("name and target are required", nil)
	}
	p := &models.Project{
		Name:           input.Name,
		Tag:            input.Tag,
		Target:         input.Target,
		Template:       input.Template,
		Node:           input.Node,
		AllNode:        input.AllNode,
		Ignore:         input.Ignore,
		Duplicates:     input.Duplicates,
		ScheduledTasks: input.ScheduledTasks,
		Hour:           input.Hour,
		Tp:             "project",
	}
	c := ginContext(ctx)
	id, replaced, err := d.projectService.ApplyProjectScope(c, p)
	if err != nil {
		return errorResult("Failed to create project", err)
	}
	if replaced {
		return jsonToolResult(ginH{
			"success":  true,
			"message":  "Project scope replaced",
			"id":       id,
			"replaced": true,
		})
	}
	return jsonToolResult(ginH{
		"success":  true,
		"message":  "Project created",
		"id":       id,
		"replaced": false,
	})
}

type listTasksInput struct {
	Search    string `json:"search,omitempty" jsonschema:"Partial task name search"`
	PageIndex int    `json:"pageIndex,omitempty" jsonschema:"Page number, starting at 1; default 1"`
	PageSize  int    `json:"pageSize,omitempty" jsonschema:"Results per page; default 20"`
}

func listTasks(ctx context.Context, _ *mcp.CallToolRequest, input listTasksInput) (*mcp.CallToolResult, any, error) {
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	c := ginContext(ctx)
	tasks, total, err := d.taskService.List(c, input.Search, input.PageIndex, input.PageSize)
	if err != nil {
		return errorResult("Failed to list tasks", err)
	}
	return jsonToolResult(ginH{"list": tasks, "total": total})
}

type getTaskInput struct {
	ID string `json:"id" jsonschema:"Task MongoDB ObjectID"`
}

func getTask(ctx context.Context, _ *mcp.CallToolRequest, input getTaskInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id is required", nil)
	}
	c := ginContext(ctx)
	result, err := d.taskService.GetTaskDetail(c, input.ID)
	if err != nil {
		return errorResult("Failed to retrieve task details", err)
	}
	if result == nil {
		return errorResult("Task not found", nil)
	}
	return jsonToolResult(result)
}

type listScanTemplatesInput struct {
	Query     string `json:"query,omitempty" jsonschema:"Partial template name search"`
	PageIndex int    `json:"pageIndex,omitempty" jsonschema:"Page number, starting at 1; default 1"`
	PageSize  int    `json:"pageSize,omitempty" jsonschema:"Results per page; default 20"`
}

func listScanTemplates(ctx context.Context, _ *mcp.CallToolRequest, input listScanTemplatesInput) (*mcp.CallToolResult, any, error) {
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	result, err := d.templateService.List(ctx, input.PageIndex, input.PageSize, input.Query)
	if err != nil {
		return errorResult("Failed to list templates", err)
	}
	return jsonToolResult(result)
}

type getScanTemplateInput struct {
	ID string `json:"id" jsonschema:"Scan template MongoDB ObjectID"`
}

func getScanTemplate(ctx context.Context, _ *mcp.CallToolRequest, input getScanTemplateInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id is required", nil)
	}
	c := ginContext(ctx)
	result, err := d.templateService.Detail(c, input.ID)
	if err != nil {
		return errorResult("Failed to retrieve template details", err)
	}
	return jsonToolResult(result)
}

type createScanTemplateInput struct {
	Name         string                       `json:"name" jsonschema:"Template name, required"`
	Modules      map[string][]string          `json:"modules,omitempty" jsonschema:"Module-to-plugin-hash mapping. Keys are module names from list_plugin_modules; values are hashes from list_plugins, executed in array order. Example: {\"SubdomainScan\":[\"d60ba73c...\"]}"`
	Parameters   map[string]map[string]string `json:"parameters,omitempty" jsonschema:"Optional parameter overrides: module name -> plugin hash -> parameter string. Omit to use plugin defaults."`
	VulList      []string                     `json:"vullist,omitempty" jsonschema:"Optional nuclei POC template IDs; used only for nuclei in VulnerabilityScan"`
	TemplateJSON string                       `json:"template_json,omitempty" jsonschema:"Optional complete ScanTemplate JSON. Takes precedence over modules for advanced configuration."`
}

func createScanTemplate(ctx context.Context, _ *mcp.CallToolRequest, input createScanTemplateInput) (*mcp.CallToolResult, any, error) {
	if input.Name == "" && input.TemplateJSON == "" {
		return errorResult("Provide name or template_json", nil)
	}

	var tmpl *models.ScanTemplate

	switch {
	case input.TemplateJSON != "":
		tmpl = &models.ScanTemplate{}
		if err := json.Unmarshal([]byte(input.TemplateJSON), tmpl); err != nil {
			return errorResult("Invalid template_json format", err)
		}
		if tmpl.Name == "" {
			tmpl.Name = input.Name
		}
	case len(input.Modules) > 0:
		built, err := buildTemplateFromModules(ctx, input)
		if err != nil {
			return errorResult(err.Error(), nil)
		}
		tmpl = built
	default:
		tmpl = &models.ScanTemplate{Name: input.Name}
	}

	if tmpl.Name == "" {
		return errorResult("Template name is required", nil)
	}

	id, err := d.templateService.Save(ctx, "", tmpl)
	if err != nil {
		return errorResult("Failed to create template", err)
	}
	return jsonToolResult(ginH{"success": true, "id": id, "message": "Template created. Use its ID as the template parameter of create_scan_task."})
}

func buildTemplateFromModules(ctx context.Context, input createScanTemplateInput) (*models.ScanTemplate, error) {
	validModule := make(map[string]bool, len(constants.PLUGINSMODULES))
	for _, m := range constants.PLUGINSMODULES {
		validModule[m] = true
	}

	templateMap := map[string]any{
		"name": input.Name,
	}
	paramMap := map[string]map[string]string{}

	for module, hashes := range input.Modules {
		if !validModule[module] {
			return nil, fmt.Errorf("Invalid module: %s (use list_plugin_modules to find valid modules)", module)
		}
		if len(hashes) == 0 {
			continue
		}
		templateMap[module] = hashes

		modParams := map[string]string{}
		for _, hash := range hashes {

			if input.Parameters != nil {
				if mp, ok := input.Parameters[module]; ok {
					if v, ok := mp[hash]; ok {
						modParams[hash] = v
						continue
					}
				}
			}

			plg, err := d.pluginService.GetPluginByHash(ctx, hash)
			if err != nil || plg == nil {
				return nil, fmt.Errorf("Plugin hash not found: %s (module %s)", hash, module)
			}
			if plg.Module != module {
				return nil, fmt.Errorf("Plugin %s (hash=%s) belongs to module %s and cannot be added to module %s", plg.Name, hash, plg.Module, module)
			}
			modParams[hash] = plg.Parameter
		}
		if len(modParams) > 0 {
			paramMap[module] = modParams
		}
	}

	if len(paramMap) > 0 {
		templateMap["Parameters"] = paramMap
	}
	if len(input.VulList) > 0 {
		templateMap["vullist"] = input.VulList
	}

	data, err := json.Marshal(templateMap)
	if err != nil {
		return nil, fmt.Errorf("Failed to assemble template: %w", err)
	}
	tmpl := &models.ScanTemplate{}
	if err := json.Unmarshal(data, tmpl); err != nil {
		return nil, fmt.Errorf("Failed to assemble template: %w", err)
	}
	return tmpl, nil
}

type createScanTaskInput struct {
	Name           string              `json:"name" jsonschema:"Task name, required and unique"`
	Target         string              `json:"target,omitempty" jsonschema:"Newline or comma-separated targets; required when targetSource is general"`
	Node           []string            `json:"node" jsonschema:"Scanner node names, required"`
	Template       string              `json:"template,omitempty" jsonschema:"Scan template ObjectID; required to execute a scan"`
	AllNode        bool                `json:"allNode,omitempty" jsonschema:"Automatically include all online nodes"`
	Ignore         string              `json:"ignore,omitempty" jsonschema:"Ignored targets, in the same format as target"`
	Duplicates     string              `json:"duplicates,omitempty" jsonschema:"Deduplication policy, such as None"`
	Project        []string            `json:"project,omitempty" jsonschema:"Associated project ObjectIDs; required when targetSource is project"`
	TargetSource   string              `json:"targetSource,omitempty" jsonschema:"Target source: general, project, asset, RootDomain, subdomain, UrlScan, or the corresponding Source variant. Default general; see tool description."`
	TargetTp       string              `json:"targetTp,omitempty" jsonschema:"Selection mode for Source variants: search or select"`
	Search         string              `json:"search,omitempty" jsonschema:"Asset search expression, with the same syntax as list_assets; see tool description"`
	Filter         map[string][]string `json:"filter,omitempty" jsonschema:"Exact filters combined with search. filter.project uses project ObjectIDs."`
	TargetNumber   int                 `json:"targetNumber,omitempty" jsonschema:"Target limit in search mode; 0 means unlimited"`
	TargetIds      []string            `json:"targetIds,omitempty" jsonschema:"Selected asset ObjectIDs in select mode"`
	BindProject    string              `json:"bindProject,omitempty" jsonschema:"Project ObjectID that owns the scan results"`
	ScheduledTasks bool                `json:"scheduledTasks,omitempty" jsonschema:"Create as a scheduled task"`
	Hour           int                 `json:"hour,omitempty" jsonschema:"Scheduled task interval in hours"`
	Minute         int                 `json:"minute,omitempty" jsonschema:"Scheduled task interval in minutes"`
	Day            int                 `json:"day,omitempty" jsonschema:"Scheduled task interval in days"`
	Week           int                 `json:"week,omitempty" jsonschema:"Weekday for weekly tasks"`
	CycleType      string              `json:"cycleType,omitempty" jsonschema:"Cycle type, such as nhours, daily or weekly"`
}

func createScanTask(ctx context.Context, _ *mcp.CallToolRequest, input createScanTaskInput) (*mcp.CallToolResult, any, error) {
	if input.Name == "" || len(input.Node) == 0 {
		return errorResult("name and node are required", nil)
	}
	c := ginContext(ctx)
	exists, err := d.taskService.CheckTaskNameExists(c, input.Name)
	if err != nil {
		return errorResult("Failed to check task name", err)
	}
	if exists {
		return errorResult("Task name already exists", nil)
	}

	targetSource := input.TargetSource
	if targetSource == "" {
		targetSource = "general"
	}

	taskModel := &models.Task{
		Name:           input.Name,
		Target:         input.Target,
		Node:           input.Node,
		Template:       input.Template,
		AllNode:        input.AllNode,
		Ignore:         input.Ignore,
		Duplicates:     input.Duplicates,
		Project:        input.Project,
		TargetSource:   targetSource,
		TargetTp:       input.TargetTp,
		Search:         input.Search,
		TargetNumber:   input.TargetNumber,
		TargetIds:      input.TargetIds,
		BindProject:    input.BindProject,
		ScheduledTasks: input.ScheduledTasks,
		Hour:           input.Hour,
		Minute:         input.Minute,
		Day:            input.Day,
		Week:           input.Week,
		CycleType:      input.CycleType,
	}
	if len(input.Filter) > 0 {
		filter := make(map[string][]interface{}, len(input.Filter))
		for k, vals := range input.Filter {
			items := make([]interface{}, len(vals))
			for i, v := range vals {
				items[i] = v
			}
			filter[k] = items
		}
		taskModel.Filter = filter
	}
	taskID, err := d.taskCommonService.Insert(ctx, taskModel)
	if err != nil {
		return errorResult("Failed to create scan task", err)
	}
	return jsonToolResult(ginH{"success": true, "id": taskID})
}

type listAssetsInput struct {
	AssetType        string              `json:"asset_type" jsonschema:"Asset type, required: asset, RootDomain, subdomain, app, mp, UrlScan, SensitiveResult, DirScanResult, crawler, vulnerability, PageMonitoring, IPAsset or SubdomainTakerResult"`
	PageIndex        int                 `json:"pageIndex,omitempty" jsonschema:"Page number, starting at 1; default 1"`
	PageSize         int                 `json:"pageSize,omitempty" jsonschema:"Results per page; default 20"`
	SearchExpression string              `json:"search,omitempty" jsonschema:"Search expression (custom DSL, not SQL); see tool description"`
	Filter           map[string][]string `json:"filter,omitempty" jsonschema:"Exact filters combined with search. filter.project uses project ObjectIDs from list_projects, not project names; filter.task uses task names. See tool description."`
	Sort             map[string]string   `json:"sort,omitempty" jsonschema:"Sort: only UrlScan and DirScanResult support length; ascending means ascending, other values mean descending"`
	Sid              string              `json:"sid,omitempty" jsonschema:"Sensitive rule name; used only for SensitiveResult"`
}

type countAssetsInput struct {
	AssetType        string              `json:"asset_type" jsonschema:"Asset type, required; same values as list_assets"`
	SearchExpression string              `json:"search,omitempty" jsonschema:"Search expression; same syntax as list_assets"`
	Filter           map[string][]string `json:"filter,omitempty" jsonschema:"Exact filters; same syntax as list_assets. filter.project uses project ObjectIDs."`
	Sid              string              `json:"sid,omitempty" jsonschema:"Sensitive rule name; used only for SensitiveResult"`
}

func countAssets(ctx context.Context, _ *mcp.CallToolRequest, input countAssetsInput) (*mcp.CallToolResult, any, error) {
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}

	query := models.SearchRequest{
		Index:            index,
		SearchExpression: input.SearchExpression,
		Sid:              input.Sid,
	}
	if len(input.Filter) > 0 {
		filter := make(map[string][]interface{}, len(input.Filter))
		for k, vals := range input.Filter {
			items := make([]interface{}, len(vals))
			for i, v := range vals {
				items[i] = v
			}
			filter[k] = items
		}
		query.Filter = filter
	}

	total, err := d.commonService.TotalData(ctx, &query)
	if err != nil {
		return errorResult("Failed to count assets", err)
	}
	return jsonToolResult(ginH{"total": total})
}

func listAssets(ctx context.Context, _ *mcp.CallToolRequest, input listAssetsInput) (*mcp.CallToolResult, any, error) {
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}
	if input.PageIndex <= 0 {
		input.PageIndex = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	query := models.SearchRequest{
		PageIndex:        input.PageIndex,
		PageSize:         input.PageSize,
		Index:            index,
		SearchExpression: input.SearchExpression,
		Sort:             input.Sort,
		Sid:              input.Sid,
	}
	if len(input.Filter) > 0 {
		filter := make(map[string][]interface{}, len(input.Filter))
		for k, vals := range input.Filter {
			items := make([]interface{}, len(vals))
			for i, v := range vals {
				items[i] = v
			}
			filter[k] = items
		}
		query.Filter = filter
	}

	c := ginContext(ctx)
	data, err := queryAssets(ctx, c, index, query)
	if err != nil {
		return errorResult("Failed to query assets", err)
	}
	return jsonToolResult(data)
}

type getAssetDetailInput struct {
	AssetType string `json:"asset_type" jsonschema:"Asset type for details: asset or vulnerability"`
	ID        string `json:"id" jsonschema:"Asset MongoDB ObjectID; use hash for vulnerability"`
}

func getAssetDetail(ctx context.Context, _ *mcp.CallToolRequest, input getAssetDetailInput) (*mcp.CallToolResult, any, error) {
	if input.ID == "" {
		return errorResult("id is required", nil)
	}
	c := ginContext(ctx)
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}

	switch index {
	case "asset":
		result, err := d.assetService.GetAssetByID(c, input.ID)
		if err != nil {
			return errorResult("Failed to retrieve asset details", err)
		}
		if result == nil {
			return errorResult("Asset not found", nil)
		}
		return jsonToolResult(result)
	case "vulnerability":
		result, err := d.vulnService.GetVulnerabilityDetailByHash(c, input.ID)
		if err != nil {
			return errorResult("Failed to retrieve vulnerability details", err)
		}
		return jsonToolResult(result)
	default:
		return errorResult("Asset detail lookup supports only asset or vulnerability", nil)
	}
}

type addAssetTagInput struct {
	AssetType string `json:"asset_type" jsonschema:"Asset type; same values as list_assets"`
	ID        string `json:"id" jsonschema:"Asset MongoDB ObjectID"`
	Tag       string `json:"tag" jsonschema:"Tag name to add"`
}

func addAssetTag(ctx context.Context, _ *mcp.CallToolRequest, input addAssetTagInput) (*mcp.CallToolResult, any, error) {
	index, err := normalizeAssetIndex(input.AssetType)
	if err != nil {
		return errorResult(err.Error(), nil)
	}
	if input.ID == "" || input.Tag == "" {
		return errorResult("id and tag are required", nil)
	}
	c := ginContext(ctx)
	req := &models.TagRequest{Type: index, ID: input.ID, Tag: input.Tag}
	if err := d.commonService.AddTag(c, req); err != nil {
		return errorResult("Failed to add tag", err)
	}
	return jsonToolResult(ginH{"success": true})
}

type listNodesInput struct {
	OnlineOnly bool `json:"online_only,omitempty" jsonschema:"true returns only online nodes; default false returns all nodes"`
}

func listNodes(ctx context.Context, _ *mcp.CallToolRequest, input listNodesInput) (*mcp.CallToolResult, any, error) {
	result, err := d.nodeService.GetNodeData(ctx, input.OnlineOnly)
	if err != nil {
		return errorResult("Failed to list nodes", err)
	}
	return jsonToolResult(ginH{"list": result})
}

func listPluginModules(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	return jsonToolResult(ginH{"modules": constants.PLUGINSMODULES})
}

type listPluginsInput struct {
	Module string `json:"module,omitempty" jsonschema:"Filter by module name from list_plugin_modules; leave empty for all scanning plugins"`
	Search string `json:"search,omitempty" jsonschema:"Partial plugin name search; used only when module is not specified"`
}

type pluginBrief struct {
	Hash      string `json:"hash"`
	Name      string `json:"name"`
	Module    string `json:"module"`
	Parameter string `json:"parameter"`
	Type      string `json:"type"`
	Status    bool   `json:"status"`
}

func toPluginBriefs(plugins []models.Plugin) []pluginBrief {
	briefs := make([]pluginBrief, 0, len(plugins))
	for _, p := range plugins {

		if p.Type == "server" {
			continue
		}
		briefs = append(briefs, pluginBrief{
			Hash:      p.Hash,
			Name:      p.Name,
			Module:    p.Module,
			Parameter: p.Parameter,
			Type:      p.Type,
			Status:    p.Status,
		})
	}
	return briefs
}

func listPlugins(ctx context.Context, _ *mcp.CallToolRequest, input listPluginsInput) (*mcp.CallToolResult, any, error) {
	c := ginContext(ctx)

	if input.Module != "" {
		plugins, err := d.pluginService.ListByModule(c, input.Module)
		if err != nil {
			return errorResult("Failed to list plugins", err)
		}
		return jsonToolResult(ginH{"list": toPluginBriefs(plugins)})
	}

	resp, err := d.pluginService.List(c, &models.PluginListRequest{
		PageIndex: 1,
		PageSize:  500,
		Search:    input.Search,
	})
	if err != nil {
		return errorResult("Failed to list plugins", err)
	}
	return jsonToolResult(ginH{"list": toPluginBriefs(resp.List), "total": resp.Total})
}

func queryAssets(ctx context.Context, c *gin.Context, index string, query models.SearchRequest) (any, error) {
	switch index {
	case "asset":
		list, err := d.assetService.GetAssets(c, query)
		return ginH{"list": list}, err
	case "RootDomain":
		return d.rootDomainService.GetRootDomainData(c, query)
	case "subdomain":
		list, err := d.subdomainService.GetSubdomains(ctx, query)
		return ginH{"list": list}, err
	case "app":
		return d.appService.GetAppData(c, query)
	case "mp":
		return d.mpService.GetMPData(c, query)
	case "UrlScan":
		list, err := d.urlService.GetURLs(ctx, query)
		return ginH{"list": list}, err
	case "SensitiveResult":
		list, err := d.sensitiveService.GetSensitiveInfo(ctx, query)
		return ginH{"list": list}, err
	case "DirScanResult":
		list, err := d.dirscanService.List(ctx, query)
		return ginH{"list": list}, err
	case "crawler":
		list, err := d.crawlerService.GetCrawlers(ctx, query)
		return ginH{"list": list}, err
	case "vulnerability":
		return d.vulnService.GetVulnerabilities(ctx, query)
	case "PageMonitoring":
		return d.pageMonService.GetResult(c, query)
	case "IPAsset":
		return d.ipService.GetIPAssets(c, query)
	case "SubdomainTakerResult":
		return querySubdomainTaker(ctx, query)
	default:
		return nil, fmt.Errorf("Unsupported asset type: %s", index)
	}
}

func querySubdomainTaker(ctx context.Context, query models.SearchRequest) (any, error) {
	takerService := subdomain.NewTakerService()
	c := ginContext(ctx)
	return takerService.GetSubdomainTakerData(c, query)
}

func normalizeAssetIndex(assetType string) (string, error) {
	assetType = strings.TrimSpace(assetType)
	if assetType == "" {
		return "", fmt.Errorf("asset_type is required")
	}
	aliases := map[string]string{
		"asset":                "asset",
		"web":                  "asset",
		"rootdomain":           "RootDomain",
		"root_domain":          "RootDomain",
		"root-domain":          "RootDomain",
		"subdomain":            "subdomain",
		"app":                  "app",
		"mp":                   "mp",
		"miniprogram":          "mp",
		"mini_program":         "mp",
		"url":                  "UrlScan",
		"urlscan":              "UrlScan",
		"sensitive":            "SensitiveResult",
		"sensitiveresult":      "SensitiveResult",
		"sensitive_result":     "SensitiveResult",
		"dirscan":              "DirScanResult",
		"dir_scan":             "DirScanResult",
		"dirscanresult":        "DirScanResult",
		"directory":            "DirScanResult",
		"crawler":              "crawler",
		"vulnerability":        "vulnerability",
		"vuln":                 "vulnerability",
		"pagemonitoring":       "PageMonitoring",
		"page_monitoring":      "PageMonitoring",
		"ip":                   "IPAsset",
		"ipasset":              "IPAsset",
		"subdomaintaker":       "SubdomainTakerResult",
		"subdomain_taker":      "SubdomainTakerResult",
		"subdomaintakerresult": "SubdomainTakerResult",
	}
	key := strings.ToLower(assetType)
	if v, ok := aliases[key]; ok {
		return v, nil
	}
	if _, ok := aliases[strings.ReplaceAll(key, "-", "_")]; ok {
		return aliases[strings.ReplaceAll(key, "-", "_")], nil
	}

	valid := []string{"asset", "RootDomain", "subdomain", "app", "mp", "UrlScan",
		"SensitiveResult", "DirScanResult", "crawler", "vulnerability",
		"PageMonitoring", "IPAsset", "SubdomainTakerResult"}
	for _, v := range valid {
		if v == assetType {
			return v, nil
		}
	}
	return "", fmt.Errorf("Unsupported asset_type: %s", assetType)
}

type ginH map[string]any
