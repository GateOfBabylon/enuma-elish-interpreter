package validate

import "regexp"

func IsValidName(name string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)
	return re.MatchString(name)
}

var supportedHttpMethodNames = []string{"POST", "GET", "DELETE"}

func IsSupportedHttpMethod(httpMethod string) bool {
	for _, name := range supportedHttpMethodNames {
		if httpMethod == name {
			return true
		}
	}
	return false
}
