package structs

import (
	"reflect"
)

func (m *Message) Convert() ClientMessage {
	return ClientMessage{Message: m.Message, SenderId: m.SenderId, ReceiverId: m.ReceiverId}
}

func GetFieldMap(obj interface{}) (fieldMap map[string]reflect.Value) {
	val := reflect.ValueOf(obj)
	fieldMap = make(map[string]reflect.Value)

	for i := 0; i < val.Elem().Type().NumField(); i++ {
		title := val.Elem().Type().Field(i).Tag.Get("json")
		field := val.Elem().Field(i)
		fieldMap[title] = field
	}

	return
}
