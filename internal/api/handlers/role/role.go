package role

//var roleRepo role.Repository
//
//// @Summary Create role
//// @Description Create a role with name, description and permissions
//// @Tags Role management
//// @Accept json
//// @Produce json
//// @Param Authorization header string true "Bearer User token"
//// @Param data body models.RoleCreate true "Role creation information"
//// @Success 201 {object} response.Response{data=RoleResponse, code=201, message=string}
//// @Failure 400 {object} response.Response{code=400, message=string}
//// @Failure 401 {object} response.Response{code=401, message=string}
//// @Failure 500 {object} response.Response{code=500, message=string}
//// @Router /api/v1/roles [post]
//func Create(c *gin.Context) {
//	var roleCreate models.RoleCreate
//	if err := c.ShouldBindJSON(&roleCreate); err != nil {
//		response.BadRequest(c, "api.bad_request", err)
//		return
//	}
//

//	existingRole, err := roleRepo.GetByName(c.Request.Context(), roleCreate.Name)
//	if err == nil && existingRole != nil {
//		response.BadRequest(c, "api.role.name.exists", nil)
//		return
//	}
//

//	role := &models.Role{
//		Name:        roleCreate.Name,
//		Description: roleCreate.Description,
//		Permissions: roleCreate.Permissions,
//		Status:      "active",
//	}
//
//	if err := roleRepo.Create(c.Request.Context(), role); err != nil {
//		logger.Error("Failed to create role", zap.Error(err))
//		response.InternalServerError(c, "api.role.create.failed", err)
//		return
//	}
//
//	roleResp := RoleResponse{
//		ID:          role.ID.Hex(),
//		Name:        role.Name,
//		Description: role.Description,
//		Permissions: role.Permissions,
//	}
//	response.Created(c, roleResp, "api.role.create.success")
//}
//
//// @Summary List roles
//// @Description List roles with pagination
//// @Tags Role management
//// @Accept json
//// @Produce json
//// @Param Authorization header string true "Bearer User token"
//// @Param page query int false "Page number" default(1)
//// @Param size query int false "Results per page" default(10)
//// @Success 200 {object} response.Response{data=response.PageResponse{list=[]RoleResponse}, code=200, message=string}
//// @Failure 400 {object} response.Response{code=400, message=string}
//// @Failure 401 {object} response.Response{code=401, message=string}
//// @Failure 500 {object} response.Response{code=500, message=string}
//// @Router /api/v1/roles [get]
//func List(c *gin.Context) {
//	page := c.DefaultQuery("page", "1")
//	size := c.DefaultQuery("size", "10")
//
//	pageInt := 1
//	sizeInt := 10
//	if page != "" {
//		if _, err := fmt.Sscanf(page, "%d", &pageInt); err != nil {
//			response.BadRequest(c, "api.bad_request.page", err)
//			return
//		}
//	}
//	if size != "" {
//		if _, err := fmt.Sscanf(size, "%d", &sizeInt); err != nil {
//			response.BadRequest(c, "api.bad_request.size", err)
//			return
//		}
//	}
//
//	roles, total, err := roleRepo.List(c.Request.Context(), int64(pageInt), int64(sizeInt))
//	if err != nil {
//		logger.Error("Failed to list roles", zap.Error(err))
//		response.InternalServerError(c, "api.role.list.failed", err)
//		return
//	}
//
//	pageData := response.PageResponse{
//		Total:    total,
//		List:     roles,
//		Page:     pageInt,
//		PageSize: sizeInt,
//	}
//	response.Success(c, pageData, "api.role.list.success")
//}
//
//// @Summary Get role details
//// @Description Get role details by ID
//// @Tags Role management
//// @Accept json
//// @Produce json
//// @Param Authorization header string true "Bearer User token"
//// @Param id path string true "Role ID"
//// @Success 200 {object} response.Response{data=RoleResponse, code=200, message=string}
//// @Failure 400 {object} response.Response{code=400, message=string}
//// @Failure 401 {object} response.Response{code=401, message=string}
//// @Failure 404 {object} response.Response{code=404, message=string}
//// @Failure 500 {object} response.Response{code=500, message=string}
//// @Router /api/v1/roles/{id} [get]
//func Get(c *gin.Context) {
//	id := c.Param("id")
//	objectID, err := primitive.ObjectIDFromHex(id)
//	if err != nil {
//		response.BadRequest(c, "api.bad_request.id", err)
//		return
//	}
//
//	role, err := roleRepo.GetByID(c.Request.Context(), objectID)
//	if err != nil {
//		response.NotFound(c, "api.role.not_found", err)
//		return
//	}
//
//	roleResp := RoleResponse{
//		ID:          role.ID.Hex(),
//		Name:        role.Name,
//		Description: role.Description,
//		Permissions: role.Permissions,
//	}
//	response.Success(c, roleResp, "api.role.get.success")
//}
//
//// @Summary Update role
//// @Description Update role information
//// @Tags Role management
//// @Accept json
//// @Produce json
//// @Param Authorization header string true "Bearer User token"
//// @Param id path string true "Role ID"
//// @Param data body models.RoleUpdate true "Role update information"
//// @Success 200 {object} response.Response{code=200, message=string}
//// @Failure 400 {object} response.Response{code=400, message=string}
//// @Failure 401 {object} response.Response{code=401, message=string}
//// @Failure 404 {object} response.Response{code=404, message=string}
//// @Failure 500 {object} response.Response{code=500, message=string}
//// @Router /api/v1/roles/{id} [put]
//func Update(c *gin.Context) {
//	id := c.Param("id")
//	objectID, err := primitive.ObjectIDFromHex(id)
//	if err != nil {
//		response.BadRequest(c, "api.bad_request.id", err)
//		return
//	}
//
//	var roleUpdate models.RoleUpdate
//	if err := c.ShouldBindJSON(&roleUpdate); err != nil {
//		response.BadRequest(c, "api.bad_request", err)
//		return
//	}
//
//	if err := roleRepo.Update(c.Request.Context(), objectID, &roleUpdate); err != nil {
//		logger.Error("Failed to update role", zap.Error(err))
//		response.InternalServerError(c, "api.role.update.failed", err)
//		return
//	}
//
//	response.Success(c, nil, "api.role.update.success")
//}
//
//// @Summary Delete role
//// @Description Delete the specified role
//// @Tags Role management
//// @Accept json
//// @Produce json
//// @Param Authorization header string true "Bearer User token"
//// @Param id path string true "Role ID"
//// @Success 200 {object} response.Response{code=200, message=string}
//// @Failure 400 {object} response.Response{code=400, message=string}
//// @Failure 401 {object} response.Response{code=401, message=string}
//// @Failure 404 {object} response.Response{code=404, message=string}
//// @Failure 500 {object} response.Response{code=500, message=string}
//// @Router /api/v1/roles/{id} [delete]
//func Delete(c *gin.Context) {
//	id := c.Param("id")
//	objectID, err := primitive.ObjectIDFromHex(id)
//	if err != nil {
//		response.BadRequest(c, "api.bad_request.id", err)
//		return
//	}
//
//	if err := roleRepo.Delete(c.Request.Context(), objectID); err != nil {
//		logger.Error("Failed to delete role", zap.Error(err))
//		response.InternalServerError(c, "api.role.delete.failed", err)
//		return
//	}
//
//	response.Success(c, nil, "api.role.delete.success")
//}
//

//type RoleResponse struct {
//	ID          string   `json:"id" example:"5f7b5e2d8f7b5e2d8f7b5e2d"`
//	Name        string   `json:"name" example:"admin"`

//	Permissions []string `json:"permissions" example:"['user:create','user:update']"`
//}
