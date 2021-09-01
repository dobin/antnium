// +build linux freebsd netbsd openbsd

package arch

func GetPermissions() (bool, bool, error) {
	isElevated := true
	isAdmin := true
	return isElevated, isAdmin, nil
}
