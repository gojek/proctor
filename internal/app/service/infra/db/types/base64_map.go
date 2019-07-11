package types

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
)

type Base64Map map[string]string

// Value implements the driver.Valuer interface, convert map into json and encode it as Base64
func (g Base64Map) Value() (driver.Value, error) {
	jsonByte, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.EncodeToString(jsonByte), nil
}

func (g *Base64Map) Scan(src interface{}) error {
	var source string
	switch src.(type) {
	case string:
		source = src.(string)
	default:
		return errors.New("incompatible type for Base64Map")
	}

	jsonByte, err := base64.StdEncoding.DecodeString(source)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonByte, g)
	if err != nil {
		return err
	}
	return nil
}
