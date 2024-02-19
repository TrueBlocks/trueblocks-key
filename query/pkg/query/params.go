package query

type RpcParams[T rpcRequestParams] []T

func (r RpcParams[T]) Validate() error {
	if len(r) != 1 {
		return ErrWrongNumOfParameters
	}
	return nil
}

func (r RpcParams[T]) Get() *T {
	return &r[0]
}
