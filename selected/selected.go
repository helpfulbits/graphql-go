package selected

import (
	"context"
	"errors"
)

const ContextKey = "Resolver_Selected_Fields"

type SelectedFields func() []SelectedField

type SelectedField struct {
	Name     string
	Args     map[string]interface{}
	Selected []SelectedField
}

func GetFieldsFromContext(ctx context.Context) (fields []SelectedField, err error) {
	fieldsFunc, ok := ctx.Value(ContextKey).(SelectedFields)
	if ok == false {
		err = errors.New("could not get graphql fields from context")
		return
	}
	fields = fieldsFunc()
	return
}
