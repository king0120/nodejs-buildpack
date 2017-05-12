package supply

import (
	"github.com/cloudfoundry/libbuildpack"
)

type Supplier struct {
	Stager *libbuildpack.Stager
}

func Run(ss *Supplier) error {
	return nil
}
