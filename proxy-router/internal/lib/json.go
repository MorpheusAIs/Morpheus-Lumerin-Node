package lib

import "encoding/json"

// NormalizeJson returns normalized json message, without spaces and newlines
func NormalizeJson(msg []byte) ([]byte, error) {
	var a interface{}
	err := json.Unmarshal(msg, &a)
	if err != nil {
		return nil, err
	}
	return json.Marshal(a)
}

func MustUnmarshallString(value json.RawMessage) string {
	var res string
	err := json.Unmarshal(value, &res)
	if err != nil {
		panic(err)
	}
	return res
}
