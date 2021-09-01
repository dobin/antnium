// +build darwin

package arch

func GetPermissions() (bool, bool, error) {
	isElevated := true
	isAdmin := true
	return isElevated, isAdmin, nil
}
