package main

import (
	"fmt"
	"reflect"
)

func mergeLists[T any](l ...[]T) []T {
	result := []T{}
	for _, i := range l {
		if len(i) != 0 {
			result = append(result, i...)
		}
	}
	return result
}

func dedup[T comparable](slice []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range slice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func structToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	val := reflect.ValueOf(obj)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	fmt.Println(typ)

	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		fieldValueKind := val.Field(i).Kind()
		var fieldValue interface{}

		if fieldValueKind == reflect.Struct {
			fieldValue = structToMap(val.Field(i).Interface())
		} else {
			fieldValue = val.Field(i).Interface()
		}

		result[fieldName] = fieldValue
	}

	return result
}
