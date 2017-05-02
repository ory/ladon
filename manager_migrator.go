package ladon

type ManagerMigrator interface {
	Create(policy Policy) (err error)
	Migrate() (err error)
	GetManager() Manager
}
