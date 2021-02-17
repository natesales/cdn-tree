// Package util includes general utility functions that aren't specific to a certain subsystem
package util

// Includes checks for array inclusion by means of linear search
func Includes(array []string, element string) bool {
	for _, item := range array {
		if item == element {
			return true
		}
	}
	return false
}
