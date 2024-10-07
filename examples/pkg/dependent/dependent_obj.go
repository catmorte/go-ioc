//go:generate go-ioc
package dependent

import (
	"github.com/catmorte/go-ioc/examples/pkg/independent"
	. "github.com/catmorte/go-ioc/pkg/context/prototype"
)

type DependentObj struct {
	Bean[*DependentObj]
	IndependentObj1 *independent.IndependentObj1 `bean:"independentScope1"`
	IndependentObj2 independent.IndependentObj2  `bean:"independentScope2"`
}
