package utils

import "regexp"

func FindExpression(regex string, str string) string {
	re := regexp.MustCompile(regex)
	return re.FindString(str)
}
