package command

func (u User) HasPermission(p Permission) bool {
	return level(u.Permission) >= level(p)
}

func level(p Permission) int {
	if p == nil {
		return 0
	}
	return p.Level()
}
