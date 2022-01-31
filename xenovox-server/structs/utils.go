package structs

import (
	"reflect"
)

func (m *Message) Convert() ClientDM {
	return ClientDM{Message: m.Message, SenderId: m.SenderId, ReceiverId: m.ReceiverId}
}

func (m *GroupMessage) Convert() ClientGM {
	return ClientGM{Message: m.Message, SenderId: m.SenderId, GroupId: m.GroupId}
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
