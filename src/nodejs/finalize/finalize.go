package finalize

import "github.com/cloudfoundry/libbuildpack"

type Finalizer struct {
	Stager *libbuildpack.Stager
}

func Run(f *Finalizer) error {
	if err := f.TipVendorDependencies(); err != nil {
		f.Stager.Log.Error(err.Error())
		return err
	}

	return nil
}

func (f *Finalizer) TipVendorDependencies() error {
	return nil
}
