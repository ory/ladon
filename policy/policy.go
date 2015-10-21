package policy

// Policy represent a policy model.
type Policy interface {
	// GetID returns the policies id.
	GetID() string

	// GetDescription returns the policies description.
	GetDescription() string

	// GetSubjects returns the policies subjects.
	GetSubjects() []string

	// HasAccess returns true if the policy effect is allow, otherwise false.
	HasAccess() bool

	// GetEffect returns the policies effect which might be 'allow' or 'deny'.
	GetEffect() string

	// GetResources returns the policies resources.
	GetResources() []string

	// GetPermissions returns the policies permissions.
	GetPermissions() []string
}

type DefaultPolicy struct {
	ID          string
	Description string
	Subjects    []string
	Effect      string
	Resources   []string
	Permissions []string
}

func (p *DefaultPolicy) GetID() string {
	return p.ID
}

func (p *DefaultPolicy) GetDescription() string {
	return p.Description
}

func (p *DefaultPolicy) GetSubjects() []string {
	return p.Subjects
}

func (p *DefaultPolicy) HasAccess() bool {
	return p.Effect == AllowAccess
}

func (p *DefaultPolicy) GetEffect() string {
	return p.Effect
}

func (p *DefaultPolicy) GetResources() []string {
	return p.Resources
}

func (p *DefaultPolicy) GetPermissions() []string {
	return p.Permissions
}
