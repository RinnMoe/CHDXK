package setting

import "sort"

type Registry struct {
	defs map[string]Definition
}

func NewRegistry(defs []Definition) *Registry {
	items := make(map[string]Definition, len(defs))
	for _, def := range defs {
		if def.Key == "" {
			continue
		}
		items[def.Key] = def
	}
	return &Registry{defs: items}
}

func (r *Registry) Get(key string) (Definition, bool) {
	if r == nil {
		return Definition{}, false
	}
	def, ok := r.defs[key]
	return def, ok
}

func (r *Registry) List(includePrivate bool) []Definition {
	if r == nil {
		return nil
	}
	defs := make([]Definition, 0, len(r.defs))
	for _, def := range r.defs {
		if !includePrivate && !def.Public {
			continue
		}
		defs = append(defs, def)
	}
	sort.Slice(defs, func(i, j int) bool {
		if defs[i].Group != defs[j].Group {
			return defs[i].Group < defs[j].Group
		}
		return defs[i].Key < defs[j].Key
	})
	return defs
}
