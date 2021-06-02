package meshobjects

type Cache map[ObjectKey]string

type ObjectKey struct {
	Mesh string
	Kind string
	Key  string
}
