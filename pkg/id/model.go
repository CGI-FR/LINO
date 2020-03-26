package id

// Table involved in an pullion plan.
type Table interface {
	Name() string
	String() string
}

// TableList involved in an pullion plan.
type TableList interface {
	Len() uint
	Table(idx uint) Table
	Contains(string) bool
	String() string
}

// Relation involved in an pullion plan.
type Relation interface {
	Name() string
	Parent() Table
	Child() Table
	String() string
}

// RelationList involved in an pullion plan.
type RelationList interface {
	Len() uint
	Relation(idx uint) Relation
	Contains(string) bool
	String() string
}

// IngressRelation describe how a relation will be accessed.
type IngressRelation interface {
	Relation
	LookUpChild() bool
	LookUpParent() bool
}

// IngressRelationList involved in an pullion plan.
type IngressRelationList interface {
	Len() uint
	Relation(idx uint) IngressRelation
	Contains(string) bool
	String() string
}

// IngressDescriptor from which the pullion plan will be computed.
type IngressDescriptor interface {
	StartTable() Table
	Relations() IngressRelationList
	String() string
}

// A Cycle in the pullion plan.
type Cycle interface {
	IngressRelationList
}

// A CycleList in the pullion plan.
type CycleList interface {
	Len() uint
	Cycle(idx uint) Cycle
	String() string
}

// An Step gives required information to pull data.
type Step interface {
	Index() uint
	Entry() Table
	Following() IngressRelation
	Relations() IngressRelationList
	Tables() TableList
	Cycles() CycleList
	PreviousStep() uint
	String() string
}

// PullionPlan is the computed plan that lists all steps required to pull data.
type PullionPlan interface {
	Len() uint
	Step(idx uint) Step
	Relations() IngressRelationList
	Tables() TableList
	String() string
}

// Error is the error type returned by the domain
type Error struct {
	Description string
}

func (e *Error) Error() string {
	return e.Description
}
