package plugin

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"recon/internal/utils"
	"recon/internal/utils/helper"
	"strings"
	"sync"
	"time"

	"recon/internal/plugins"
	schedulerCore "recon/internal/scheduler"

	"github.com/gin-gonic/gin"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"recon/internal/constants"
	"recon/internal/database/mongodb"
	"recon/internal/logger"
	"recon/internal/models"
	"recon/internal/options"
	"recon/internal/repositories/plugin"
	"recon/internal/services/node"
)

type Service interface {
	List(ctx *gin.Context, req *models.PluginListRequest) (*models.PluginListResponse, error)
	ListByModule(ctx *gin.Context, module string) ([]models.Plugin, error)
	GetPluginByHash(ctx context.Context, hash string) (*models.Plugin, error)
	Detail(ctx *gin.Context, req *models.PluginDetailRequest) (*models.Plugin, error)
	Save(ctx *gin.Context, req *models.PluginSaveRequest) error
	Delete(ctx *gin.Context, req *models.PluginDeleteRequest) error
	GetLogs(ctx *gin.Context, req *models.PluginLogRequest) (string, error)
	CleanLogs(ctx *gin.Context, req *models.PluginLogRequest) error
	CleanAllLogs(ctx *gin.Context) error
	Import(ctx *gin.Context, filePath string, key string) error
	Reinstall(ctx *gin.Context, req *models.PluginReinstallRequest) error
	Recheck(ctx *gin.Context, req *models.PluginRecheckRequest) error
	Uninstall(ctx *gin.Context, req *models.PluginUninstallRequest) error
	CheckKey(ctx *gin.Context, req *models.PluginKeyCheckRequest) error
	SearchRemotePlugins(ctx *gin.Context) (map[string]interface{}, error)
	ImportByData(ctx *gin.Context, req *models.PluginImportByDataRequest) error
	UpdateStatus(ctx *gin.Context, req *models.PluginStatusRequest) error
	Run(ctx *gin.Context, req *models.PluginRunRequest) error
	RunTaskEnd(task models.Task) error
}

type service struct {
	repo        plugin.Repository
	nodeService node.Service
}

func NewService() Service {
	return &service{
		repo:        plugin.NewRepository(),
		nodeService: node.NewService(),
	}
}

func (s *service) List(ctx *gin.Context, req *models.PluginListRequest) (*models.PluginListResponse, error) {
	query := bson.M{}
	if req.Type != "all" {
		if req.Type != "server" {
			query["type"] = bson.M{"$ne": "server"}
		} else {
			query["type"] = req.Type
		}
	}
	if req.Search != "" {
		query["name"] = bson.M{"$regex": req.Search, "$options": "i"}
	}

	opts := mongoOptions.Find().
		SetSkip(int64((req.PageIndex - 1) * req.PageSize)).
		SetLimit(int64(req.PageSize))

	plugins, err := s.repo.FindWithPagination(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find plugins: %w", err)
	}

	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to count plugins: %w", err)
	}

	return &models.PluginListResponse{
		List:  plugins,
		Total: total,
	}, nil
}

func (s *service) ListByModule(ctx *gin.Context, module string) ([]models.Plugin, error) {
	result, err := s.repo.FindByModule(ctx, module)
	if result == nil {
		return []models.Plugin{}, err
	}
	return result, err
}

func (s *service) GetPluginByHash(ctx context.Context, hash string) (*models.Plugin, error) {
	return s.repo.FindByHash(ctx, hash)
}

func (s *service) Detail(ctx *gin.Context, req *models.PluginDetailRequest) (*models.Plugin, error) {
	id, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %w", err)
	}

	return s.repo.FindByID(ctx, id)
}

func (s *service) Save(ctx *gin.Context, req *models.PluginSaveRequest) error {
	if err := checkPluginKey(ctx, req.Key); err != nil {
		return err
	}

	if req.ID == "" {

		plugin := &models.Plugin{
			Name:          req.Name,
			Module:        req.Module,
			Hash:          generatePluginHash(32, req.Type),
			Parameter:     req.Parameter,
			Help:          req.Help,
			Introduction:  req.Introduction,
			Source:        req.Source,
			Version:       req.Version,
			ParameterList: req.ParameterList,
			Status:        true,
			Type:          req.Type,
			IsSystem:      false,
		}
		id, err := s.repo.Create(ctx, plugin)
		if err != nil {
			return fmt.Errorf("failed to create plugin: %w", err)
		}
		if req.Type != "server" {
			go func() {
				msg := models.Message{
					Name:    "all",
					Type:    "install_plugin",
					Content: fmt.Sprintf(`%v`, id),
				}
				err = s.nodeService.RefreshConfig(ctx, msg)
				if err != nil {
					logger.Error("failed to refresh config", zap.Error(err))
				}
			}()
		} else {
			RegisterPlugin(plugin.Source, plugin.Hash)
		}
	} else {

		id, err := primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			return fmt.Errorf("invalid id format: %w", err)
		}

		update := bson.M{
			"name":          req.Name,
			"module":        req.Module,
			"parameter":     req.Parameter,
			"help":          req.Help,
			"introduction":  req.Introduction,
			"parameterList": req.ParameterList,
			"source":        req.Source,
			"version":       req.Version,
		}
		err = s.repo.Update(ctx, id, update)
		if err != nil {
			return err
		}
		if req.Type != "server" {
			go func() {
				msg := models.Message{
					Name:    "all",
					Type:    "install_plugin",
					Content: fmt.Sprintf(`%v`, req.ID),
				}
				err = s.nodeService.RefreshConfig(ctx, msg)
				if err != nil {
					logger.Error("failed to refresh config", zap.Error(err))
				}
			}()
		} else {

			RegisterPlugin(req.Source, req.Hash)
		}
	}

	if req.Type == "server" {
		plugins.RunPluginOnce(req.Hash)
	}
	return nil
}

func (s *service) Delete(ctx *gin.Context, req *models.PluginDeleteRequest) error {
	var hashes []string
	for _, item := range req.Data {
		hashes = append(hashes, item.Hash)
		go func() {
			if item.Module != "" {
				msg := models.Message{
					Name:    "all",
					Type:    "delete_plugin",
					Content: fmt.Sprintf(`%v_%v`, item.Hash, item.Module),
				}
				err := s.nodeService.RefreshConfig(ctx, msg)
				if err != nil {
					logger.Error("failed to refresh config", zap.Error(err))
				}
			} else {

				plugins.GlobalPluginManager.DeletePlugin(item.Hash)

				err := schedulerCore.GetGlobalScheduler().RemoveJob(item.Hash)
				if err != nil {
					logger.Error(fmt.Sprintf("failed to remove job: %v", err))
				}
			}
		}()
	}
	return s.repo.DeleteByHash(ctx, hashes)
}

func (s *service) GetLogs(ctx *gin.Context, req *models.PluginLogRequest) (string, error) {
	logKey := ""
	if req.Type == "server" {
		logKey = fmt.Sprintf("logs:server_plugins:%v", req.Hash)
	} else {
		logKey = fmt.Sprintf("logs:plugins:%v:%v", req.Module, req.Hash)
	}
	logs, err := s.repo.GetLogs(ctx, logKey)
	if err != nil {
		return "", err
	}
	return logs, nil
}

func (s *service) CleanLogs(ctx *gin.Context, req *models.PluginLogRequest) error {
	if req.Hash == "" {
		return fmt.Errorf("module and hash are required")
	}
	var logKey string
	if req.Type == "server" {
		logKey = fmt.Sprintf("logs:server_plugins:%v", req.Hash)
	} else {
		logKey = fmt.Sprintf("logs:plugins:%v:%v", req.Module, req.Hash)
	}
	err := s.repo.CleanLogs(ctx, logKey)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) CleanAllLogs(ctx *gin.Context) error {
	err := s.repo.CleanAllLogs(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Import(ctx *gin.Context, filePath string, reqKey string) error {
	if err := checkPluginKey(ctx, reqKey); err != nil {
		return err
	}

	zipData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Failed to read ZIP file: %w", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("Failed to open ZIP file: %w", err)
	}

	var pluginInfo models.PluginInfo
	var pluginSource string

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		baseName := strings.ToLower(filepath.Base(file.Name))

		f, err := file.Open()
		if err != nil {
			return fmt.Errorf("Failed to open archive entry: %w", err)
		}
		content, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return fmt.Errorf("Failed to read archive entry: %w", err)
		}

		switch baseName {
		case "info.json":
			if err := json.Unmarshal(content, &pluginInfo); err != nil {
				return fmt.Errorf("Failed to parse info.json: %w", err)
			}
		case "plugin.go":
			pluginSource = string(content)
		}
	}

	if pluginInfo.Name == "" {
		return fmt.Errorf("info.json is missing name")
	}
	if pluginInfo.Module == "" && pluginInfo.Type != "server" {
		return fmt.Errorf("info.json is missing module")
	}

	if pluginInfo.Hash == "" {
		pluginInfo.Hash = generatePluginHash(32, pluginInfo.Type)
	}
	var PLUGINS = make(map[string]bool)
	for _, p := range constants.Plugins {
		PLUGINS[p.Hash] = true
	}

	if PLUGINS[pluginInfo.Hash] {
		return fmt.Errorf("Plugin already exists")
	}
	var pluginsModules = make(map[string]bool)
	for _, module := range constants.PLUGINSMODULES {
		pluginsModules[module] = true
	}
	if pluginInfo.Type != "server" {
		if !pluginsModules[pluginInfo.Module] {
			return fmt.Errorf("Invalid module: %s", pluginInfo.Module)
		}
	}

	pluginInfo.Source = pluginSource
	pluginInfo.IsSystem = false

	p := &models.Plugin{
		Name:          pluginInfo.Name,
		Module:        pluginInfo.Module,
		Hash:          pluginInfo.Hash,
		Parameter:     pluginInfo.Parameter,
		Help:          pluginInfo.Help,
		Introduction:  pluginInfo.Introduction,
		ParameterList: pluginInfo.ParameterList,
		Source:        pluginInfo.Source,
		Version:       pluginInfo.Version,
		Type:          pluginInfo.Type,
		IsSystem:      false,
	}
	id, err := s.repo.Create(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}
	go func() {
		if pluginInfo.Type != "server" {
			msg := models.Message{
				Name:    "all",
				Type:    "install_plugin",
				Content: fmt.Sprintf(`%v`, id),
			}
			err = s.nodeService.RefreshConfig(ctx, msg)
			if err != nil {
				logger.Error("failed to refresh config", zap.Error(err))
			}
		}
	}()
	if pluginInfo.Type == "server" {
		RegisterPlugin(pluginInfo.Source, pluginInfo.Hash)

		plugins.RunPluginOnce(pluginInfo.Hash)
	}

	return nil
}

func (s *service) Reinstall(ctx *gin.Context, req *models.PluginReinstallRequest) error {
	if req.Node == "" {
		req.Node = "all"
	}
	if req.Module == "" {

		plg, flag := plugins.GlobalPluginManager.GetPlugin(req.Hash)
		if !flag {
			logger.Error(fmt.Sprintf("Plugin %s not found", req.Hash))
			return fmt.Errorf("Plugin %s not found", req.Hash)
		}
		plg.Install()
	} else {
		msg := models.Message{
			Name:    req.Node,
			Type:    "re_install_plugin",
			Content: fmt.Sprintf(`%v_%v`, req.Hash, req.Module),
		}
		err := s.nodeService.RefreshConfig(ctx, msg)
		if err != nil {
			logger.Error("failed to refresh config", zap.Error(err))
		}
	}
	return nil
}

func (s *service) Recheck(ctx *gin.Context, req *models.PluginRecheckRequest) error {
	if req.Node == "" {
		req.Node = "all"
	}
	msg := models.Message{
		Name:    req.Node,
		Type:    "re_check_plugin",
		Content: fmt.Sprintf(`%v_%v`, req.Hash, req.Module),
	}
	err := s.nodeService.RefreshConfig(ctx, msg)
	if err != nil {
		logger.Error("failed to refresh config", zap.Error(err))
	}
	return nil
}

func (s *service) Uninstall(ctx *gin.Context, req *models.PluginUninstallRequest) error {

	if req.Node == "" {
		req.Node = "all"
	}
	msg := models.Message{
		Name:    req.Node,
		Type:    "uninstall_plugin",
		Content: fmt.Sprintf(`%v_%v`, req.Hash, req.Module),
	}
	err := s.nodeService.RefreshConfig(ctx, msg)
	if err != nil {
		logger.Error("failed to refresh config", zap.Error(err))
	}
	return nil
}

func (s *service) CheckKey(ctx *gin.Context, req *models.PluginKeyCheckRequest) error {
	return checkPluginKey(ctx, req.Key)
}

func checkPluginKey(ctx *gin.Context, supplied string) error {
	if ctx.GetBool(constants.GatewayAuthenticatedContextKey) && ctx.GetString("role") == "super_admin" {
		return nil
	}
	key, err := ioutil.ReadFile("PLUGINKEY")
	if err != nil {
		return fmt.Errorf("failed to read plugin key: %w", err)
	}
	if supplied != strings.TrimSpace(string(key)) {
		return fmt.Errorf("invalid plugin key")
	}
	return nil
}

type RemotePluginResponse struct {
	Timestamp int64            `json:"timestamp"`
	Status    string           `json:"status"`
	Message   string           `json:"message"`
	Data      RemotePluginData `json:"data"`
}

type RemotePluginData struct {
	Total int64              `json:"total"`
	Data  []RemotePluginItem `json:"data"`
}

type RemotePluginItem struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Module       string `json:"module"`
	PriceStatus  int    `json:"priceStatus"`
	Price        *int   `json:"price"`
	Hash         string `json:"hash"`
	Introduction string `json:"introduction"`
	Version      string `json:"version"`
	CreateTime   string `json:"createTime"`
	Username     string `json:"username"`
	IsInstalled  bool   `json:"isInstalled"`
	NeedUpdate   bool   `json:"needUpdate"`
}

func (s *service) SearchRemotePlugins(ctx *gin.Context) (map[string]interface{}, error) {
	var installedPlugins []models.Plugin
	var remotePlugins RemotePluginResponse
	var installedErr, remoteErr error
	var wg sync.WaitGroup
	catalogURL := strings.TrimSpace(os.Getenv("RECON_PLUGIN_CATALOG_URL"))

	wg.Add(2)

	go func() {
		defer wg.Done()
		query := bson.M{}
		opts := mongoOptions.Find()
		installedPlugins, installedErr = s.repo.FindWithPagination(ctx, query, opts)
	}()

	go func() {
		defer wg.Done()
		if catalogURL == "" {
			return
		}
		client := &fasthttp.Client{
			ReadTimeout: 30 * time.Second,
		}

		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI(catalogURL)
		req.Header.SetMethod("POST")
		req.Header.SetContentType("application/json")

		requestBody := map[string]interface{}{
			"name":        "",
			"priceStatus": 2,
			"module":      "All",
			"tags":        "",
			"page":        1,
			"size":        200,
		}
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			remoteErr = fmt.Errorf("failed to marshal request body: %w", err)
			return
		}
		req.SetBody(bodyBytes)

		err = client.Do(req, resp)
		if err != nil {
			remoteErr = fmt.Errorf("failed to request remote api: %w", err)
			return
		}

		if resp.StatusCode() != http.StatusOK {
			remoteErr = fmt.Errorf("remote api returned status code: %d", resp.StatusCode())
			return
		}

		if err := json.Unmarshal(resp.Body(), &remotePlugins); err != nil {
			remoteErr = fmt.Errorf("failed to unmarshal response: %w", err)
			return
		}
	}()

	wg.Wait()

	if installedErr != nil {
		return nil, fmt.Errorf("failed to get installed plugins: %w", installedErr)
	}
	if remoteErr != nil {
		return nil, fmt.Errorf("failed to get remote plugins: %w", remoteErr)
	}
	if catalogURL == "" {
		remotePlugins.Data.Total = int64(len(installedPlugins))
		for index, plugin := range installedPlugins {
			remotePlugins.Data.Data = append(remotePlugins.Data.Data, RemotePluginItem{
				ID: index + 1, Name: plugin.Name, Module: plugin.Module, Hash: plugin.Hash,
				Introduction: plugin.Introduction, Version: plugin.Version, PriceStatus: 2,
			})
		}
	}

	installedHashMap := make(map[string]models.Plugin)
	for _, plugin := range installedPlugins {
		installedHashMap[plugin.Hash] = plugin
	}

	resultPlugins := make([]RemotePluginItem, 0, len(remotePlugins.Data.Data))
	for _, remotePlugin := range remotePlugins.Data.Data {
		installedPlugin, isInstalled := installedHashMap[remotePlugin.Hash]

		remotePlugin.IsInstalled = isInstalled

		if isInstalled {

			remoteVersion := strings.TrimSpace(remotePlugin.Version)
			installedVersion := strings.TrimSpace(installedPlugin.Version)

			if remoteVersion != "" && remoteVersion != installedVersion {
				remotePlugin.NeedUpdate = true
			} else {
				remotePlugin.NeedUpdate = false
			}
		} else {
			remotePlugin.NeedUpdate = false
		}

		resultPlugins = append(resultPlugins, remotePlugin)
	}

	return map[string]interface{}{
		"total": remotePlugins.Data.Total,
		"data":  resultPlugins,
	}, nil
}

func (s *service) ImportByData(ctx *gin.Context, req *models.PluginImportByDataRequest) error {
	if err := checkPluginKey(ctx, req.Key); err != nil {
		return err
	}

	var pluginInfo models.PluginInfo
	if err := json.Unmarshal([]byte(req.JSON), &pluginInfo); err != nil {
		return fmt.Errorf("Failed to parse JSON string: %w", err)
	}

	if pluginInfo.Name == "" {
		return fmt.Errorf("JSON is missing name")
	}
	if pluginInfo.Type != "server" && pluginInfo.Module == "" {
		return fmt.Errorf("JSON is missing Module")
	}

	if pluginInfo.Hash == "" {
		pluginInfo.Hash = generatePluginHash(32, pluginInfo.Type)
	}

	if pluginInfo.Type != "server" {
		var pluginsModules = make(map[string]bool)
		for _, module := range constants.PLUGINSMODULES {
			pluginsModules[module] = true
		}
		if !pluginsModules[pluginInfo.Module] {
			return fmt.Errorf("Invalid module: %s", pluginInfo.Module)
		}
	}

	var PLUGINS = make(map[string]bool)
	for _, p := range constants.Plugins {
		PLUGINS[p.Hash] = true
	}

	if PLUGINS[pluginInfo.Hash] {
		req.IsSystem = true
	}

	p := &models.Plugin{
		Name:          pluginInfo.Name,
		Module:        pluginInfo.Module,
		Hash:          pluginInfo.Hash,
		Parameter:     pluginInfo.Parameter,
		Help:          pluginInfo.Help,
		Introduction:  pluginInfo.Introduction,
		ParameterList: pluginInfo.ParameterList,
		Source:        req.Source,
		Version:       pluginInfo.Version,
		IsSystem:      req.IsSystem,
		Type:          pluginInfo.Type,
	}

	pluginID, err := s.repo.UpsertByHash(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to upsert plugin: %w", err)
	}

	if !req.IsSystem && pluginInfo.Type != "server" {
		go func() {
			msg := models.Message{
				Name:    "all",
				Type:    "install_plugin",
				Content: fmt.Sprintf(`%v`, pluginID),
			}
			err = s.nodeService.RefreshConfig(ctx, msg)
			if err != nil {
				logger.Error("failed to refresh config", zap.Error(err))
			}
		}()
	}
	if pluginInfo.Type == "server" {
		RegisterPlugin(req.Source, pluginInfo.Hash)
		plugins.RunPluginOnce(pluginInfo.Hash)
	}

	return nil
}

const pluginSalt = "Recon_"
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generatePluginHash(length int, tp string) string {

	rand.Seed(time.Now().UnixNano())

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	randomString := string(b)

	salted := randomString + pluginSalt + tp

	hash := md5.Sum([]byte(salted))

	return hex.EncodeToString(hash[:])
}

func (s *service) UpdateStatus(ctx *gin.Context, req *models.PluginStatusRequest) error {

	update := bson.M{
		"status": req.Status,
	}
	err := s.repo.UpdateByHash(ctx, req.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}
	if req.Status {
		plg, flag := plugins.GlobalPluginManager.GetPlugin(req.ID)
		if !flag {
			logger.Error(fmt.Sprintf("failed to get plugin status: %v", req.ID))
			return fmt.Errorf("failed to get plugin: %v", req.ID)
		}
		if plg.Cycle() != "" {
			if plg.Cycle() == "1" {
				plugins.RunPluginOnce(plg.GetPluginId())
			} else {
				err := AddJob(plg.GetPluginId(), plg.Cycle())
				if err != nil {
					return err
				}
			}
		}
	} else {
		flag := utils.HasContext(req.ID)
		if flag {
			utils.CancelContext(req.ID)
		}
		err := schedulerCore.GetGlobalScheduler().RemoveJob(req.ID)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to remove job: %v", err))
			return err
		}
	}
	return nil
}

func (s *service) Run(ctx *gin.Context, req *models.PluginRunRequest) error {

	plgRunner, flag := plugins.GlobalPluginManager.GetPlugin(req.Hash)
	if !flag {

		plg, err := s.repo.FindByHash(ctx, req.Hash)
		if err != nil {
			return fmt.Errorf("failed to find plugin: %w", err)
		}
		if plg == nil {
			return fmt.Errorf("plugin not found")
		}

		if plg.Type != "server" {
			return fmt.Errorf("only server plugins can be run manually")
		}

		plugin, err := plugins.LoadPlugin(plg.Source, plg.Hash)
		if err != nil {
			return fmt.Errorf("failed to load plugin: %w", err)
		}
		plugins.GlobalPluginManager.RegisterPlugin(req.Hash, plugin)
		err = plugin.Install()
		if err != nil {
			logger.Error(fmt.Sprintf("failed to install plugin: %v", err))
			return fmt.Errorf("failed to install plugin: %w", err)
		}
		plgRunner = plugin.Clone()
	}

	plgRunner.Log("plugin is running manually")
	_, err := mongodb.UpdateOne("plugins", bson.M{"hash": req.Hash}, bson.M{"$set": bson.M{"lastTime": helper.GetNowTimeString()}})
	if err != nil {
		logger.Error("failed to update plugin", zap.Error(err))
	}

	go func() {
		err = plgRunner.Execute(options.NewOption(req.Hash))
		if err != nil {
			logger.Error(fmt.Sprintf("plugin execution error: %v", err))
			plgRunner.Log(fmt.Sprintf("plugin execution error: %v", err), "e")
		}
		plgRunner.Log("plugin execution completed")
	}()

	return nil
}

func (s *service) RunTaskEnd(task models.Task) error {
	query := bson.M{"type": "server", "status": true}
	logger.Info(fmt.Sprintf("executing task end plugin: %v", task.Name))
	opts := mongoOptions.Find()

	plgs, err := s.repo.FindWithPagination(context.Background(), query, opts)
	if err != nil {
		return fmt.Errorf("failed to find plugins: %w", err)
	}
	for _, plgTmp := range plgs {
		plg, flag := plugins.GlobalPluginManager.GetPlugin(plgTmp.Hash)
		if flag {
			go func() {
				err := plg.TaskEnd(task, plgTmp.Hash)
				if err != nil {
					logger.Error(fmt.Sprintf("plugin taskend execution error: %v", err))
					return
				}
			}()
		} else {
			logger.Error(fmt.Sprintf("plugin not found: %v", plgTmp.Hash))
		}
	}
	return nil
}
