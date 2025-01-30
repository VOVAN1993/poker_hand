package persistent

type (
	Where struct {
		ID *string
	}
	WhereOpt func(where *Where)
)

func WithID(id string) WhereOpt {
	return func(w *Where) {
		w.ID = &id
	}
}

func constructsOption(fns ...WhereOpt) Where {
	o := Where{}
	for _, f := range fns {
		f(&o)
	}
	return o
}
