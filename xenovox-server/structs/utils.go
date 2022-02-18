package structs

import (
	"reflect"
)

func (m *Message) Convert() ClientDM {
	return ClientDM{Message: m.Message, SenderId: m.SenderId, ReceiverId: m.ReceiverId, Id: m.Id}
}

func (m *GroupMessage) Convert() ClientGM {
	return ClientGM{Message: m.Message, SenderId: m.SenderId, GroupId: m.GroupId, IsSystem: m.IsSystem, Id: m.Id}
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
