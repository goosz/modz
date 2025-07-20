package modz

type Binder interface {
	getData(DataKey) (any, error)
	putData(DataKey, any) error
}
