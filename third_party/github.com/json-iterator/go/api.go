package jsoniter

import "encoding/json"

type API interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type apiImpl struct{}

var ConfigCompatibleWithStandardLibrary API = apiImpl{}

func (apiImpl) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (apiImpl) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
