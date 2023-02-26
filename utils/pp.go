package utils

import (
	"encoding/json"
	"fmt"
)

func PP(s any) {
	res, err := PrettyStruct(s)
	if err != nil {
		lg.Panic(err)
	}
	fmt.Println(res)
}

func PrettyStruct(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "ERR", err
	}
	return string(val), nil
}
