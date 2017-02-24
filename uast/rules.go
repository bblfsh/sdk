package uast

type Rule func(...*Node) error

type Selector func(...*Node) bool

func Rules(rules ...Rule) Rule {
	return func(ns ...*Node) error {
		for _, r := range rules {
			if err := r(ns...); err != nil {
				return err
			}
		}

		return nil
	}
}

func (s Selector) Role(roles ...Role) Rule {
	return func(ns ...*Node) error {
		if s(ns...) {
			n := ns[len(ns)-1]
			n.Roles = append(n.Roles, roles...)
		}

		return nil
	}
}

var OnNoRole Selector = func(ns ...*Node) bool {
	if len(ns) == 0 {
		return false
	}

	n := ns[len(ns) - 1]
	return len(n.Roles) == 0
}

func OnInternalType(path ...string) Selector {
	return func(ns ...*Node) bool {
		if len(path) == 0 {
			return false
		}

		if len(path) > len(ns) {
			return false
		}

		i := len(path) - 1
		j := len(ns) - 1
		for i >= 0 && j >= 0 {
			p := path[i]
			n := ns[j]
			if n.InternalType != p {
				return false
			}

			i--
			j--
		}

		return true
	}
}

func OnInternalRole(path ...string) Selector {
	return func(ns ...*Node) bool {
		if len(path) == 0 {
			return false
		}

		if len(path) > len(ns) {
			return false
		}

		i := len(path) - 1
		j := len(ns) - 1
		for i >= 0 && j >= 0 {
			p := path[i]
			n := ns[j]
			r, ok := n.Properties[InternalRoleKey]
			if !ok || r != p {
				return false
			}

			i--
			j--
		}

		return true
	}
}
