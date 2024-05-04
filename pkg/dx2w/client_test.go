package dx2w

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestClient_ReadAll(t *testing.T) {
	// Mock modbus client and responses

	client := New(
		TCPDevice{
			Url: "tcp://192.168.50.60:502",
			Id:  200,
		},
		time.Minute,
	)

	jsonBytes := func(data interface{}) []byte {
		json, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return []byte(err.Error())
		}
		return json
	}

	result := client.ReadAll()
	fmt.Println(string(jsonBytes(result)))
}
