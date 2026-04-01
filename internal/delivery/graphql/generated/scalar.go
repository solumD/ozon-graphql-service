// Этот файл написан вручную.
// Он находится в пакете generated, потому что сгенерированный gqlgen-код
// использует методы executionContext, определённые здесь, для кастомного scalar Time.
package generated

import (
	"fmt"
	"io"
	"time"

	gqlgengraphql "github.com/99designs/gqlgen/graphql"
)

func MarshalTime(value time.Time) gqlgengraphql.Marshaler {
	return gqlgengraphql.MarshalTime(value)
}

func UnmarshalTime(value any) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time format: %w", err)
		}
		return parsed, nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time value type %T", value)
	}
}

func (ec *executionContext) unmarshalInputTime(_ any, value any) (time.Time, error) {
	return UnmarshalTime(value)
}

func (ec *executionContext) _Time(_ any, _ any, value *time.Time) gqlgengraphql.Marshaler {
	if value == nil {
		return gqlgengraphql.Null
	}

	return gqlgengraphql.WriterFunc(func(w io.Writer) {
		MarshalTime(*value).MarshalGQL(w)
	})
}
