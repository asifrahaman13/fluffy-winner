package helper

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ConvertToSlice(a primitive.A) []string {
	result := make([]string, len(a))
	for i, v := range a {
		result[i] = fmt.Sprint(v)
	}
	return result
}