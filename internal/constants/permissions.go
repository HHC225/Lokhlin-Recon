package constants

const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
	RoleGuest      = "guest"
)

const (
	PrefixUser    = "user"
	PrefixRole    = "role"
	PrefixAsset   = "asset"
	PrefixDomain  = "domain"
	PrefixApp     = "app"
	PrefixTask    = "task"
	PrefixProject = "project"
	PrefixSystem  = "system"
)

const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionList   = "list"
	ActionExport = "export"
	ActionImport = "import"
)

const (
	PermissionUserCreate = PrefixUser + ":" + ActionCreate
	PermissionUserRead   = PrefixUser + ":" + ActionRead
	PermissionUserUpdate = PrefixUser + ":" + ActionUpdate
	PermissionUserDelete = PrefixUser + ":" + ActionDelete
	PermissionUserList   = PrefixUser + ":" + ActionList

	PermissionRoleCreate = PrefixRole + ":" + ActionCreate
	PermissionRoleRead   = PrefixRole + ":" + ActionRead
	PermissionRoleUpdate = PrefixRole + ":" + ActionUpdate
	PermissionRoleDelete = PrefixRole + ":" + ActionDelete
	PermissionRoleList   = PrefixRole + ":" + ActionList

	PermissionAssetCreate = PrefixAsset + ":" + ActionCreate
	PermissionAssetRead   = PrefixAsset + ":" + ActionRead
	PermissionAssetUpdate = PrefixAsset + ":" + ActionUpdate
	PermissionAssetDelete = PrefixAsset + ":" + ActionDelete
	PermissionAssetList   = PrefixAsset + ":" + ActionList
	PermissionAssetExport = PrefixAsset + ":" + ActionExport
	PermissionAssetImport = PrefixAsset + ":" + ActionImport

	PermissionDomainCreate = PrefixDomain + ":" + ActionCreate
	PermissionDomainRead   = PrefixDomain + ":" + ActionRead
	PermissionDomainUpdate = PrefixDomain + ":" + ActionUpdate
	PermissionDomainDelete = PrefixDomain + ":" + ActionDelete
	PermissionDomainList   = PrefixDomain + ":" + ActionList
	PermissionDomainExport = PrefixDomain + ":" + ActionExport
	PermissionDomainImport = PrefixDomain + ":" + ActionImport

	PermissionAppCreate = PrefixApp + ":" + ActionCreate
	PermissionAppRead   = PrefixApp + ":" + ActionRead
	PermissionAppUpdate = PrefixApp + ":" + ActionUpdate
	PermissionAppDelete = PrefixApp + ":" + ActionDelete
	PermissionAppList   = PrefixApp + ":" + ActionList

	PermissionTaskCreate = PrefixTask + ":" + ActionCreate
	PermissionTaskRead   = PrefixTask + ":" + ActionRead
	PermissionTaskUpdate = PrefixTask + ":" + ActionUpdate
	PermissionTaskDelete = PrefixTask + ":" + ActionDelete
	PermissionTaskList   = PrefixTask + ":" + ActionList
	PermissionTaskExport = PrefixTask + ":" + ActionExport

	PermissionProjectCreate = PrefixProject + ":" + ActionCreate
	PermissionProjectRead   = PrefixProject + ":" + ActionRead
	PermissionProjectUpdate = PrefixProject + ":" + ActionUpdate
	PermissionProjectDelete = PrefixProject + ":" + ActionDelete
	PermissionProjectList   = PrefixProject + ":" + ActionList

	PermissionSystemConfig = PrefixSystem + ":config"
	PermissionSystemLog    = PrefixSystem + ":log"
	PermissionSystemBackup = PrefixSystem + ":backup"
)

var RolePermissions = map[string][]string{
	RoleSuperAdmin: {

		PermissionUserCreate,
		PermissionUserRead,
		PermissionUserUpdate,
		PermissionUserDelete,
		PermissionUserList,

		PermissionRoleCreate,
		PermissionRoleRead,
		PermissionRoleUpdate,
		PermissionRoleDelete,
		PermissionRoleList,

		PermissionAssetCreate,
		PermissionAssetRead,
		PermissionAssetUpdate,
		PermissionAssetDelete,
		PermissionAssetList,
		PermissionAssetExport,
		PermissionAssetImport,

		PermissionDomainCreate,
		PermissionDomainRead,
		PermissionDomainUpdate,
		PermissionDomainDelete,
		PermissionDomainList,
		PermissionDomainExport,
		PermissionDomainImport,

		PermissionAppCreate,
		PermissionAppRead,
		PermissionAppUpdate,
		PermissionAppDelete,
		PermissionAppList,

		PermissionTaskCreate,
		PermissionTaskRead,
		PermissionTaskUpdate,
		PermissionTaskDelete,
		PermissionTaskList,
		PermissionTaskExport,

		PermissionProjectCreate,
		PermissionProjectRead,
		PermissionProjectUpdate,
		PermissionProjectDelete,
		PermissionProjectList,

		PermissionSystemConfig,
		PermissionSystemLog,
		PermissionSystemBackup,
	},
	RoleAdmin: {

		PermissionUserRead,
		PermissionUserList,

		PermissionAssetRead,
		PermissionAssetList,
		PermissionAssetExport,

		PermissionDomainRead,
		PermissionDomainList,
		PermissionDomainExport,

		PermissionAppRead,
		PermissionAppList,

		PermissionTaskCreate,
		PermissionTaskRead,
		PermissionTaskList,
		PermissionTaskExport,

		PermissionProjectCreate,
		PermissionProjectRead,
		PermissionProjectUpdate,
		PermissionProjectList,
	},
	RoleUser: {

		PermissionAssetRead,
		PermissionAssetList,

		PermissionDomainRead,
		PermissionDomainList,

		PermissionAppRead,
		PermissionAppList,

		PermissionTaskRead,
		PermissionTaskList,

		PermissionProjectRead,
		PermissionProjectList,
	},
	RoleGuest: {

		PermissionAssetRead,
		PermissionDomainRead,
		PermissionAppRead,
		PermissionTaskRead,
		PermissionProjectRead,
	},
}
