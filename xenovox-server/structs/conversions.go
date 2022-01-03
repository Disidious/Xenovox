package structs

func (m *Message) Convert() ClientMessage {
	return ClientMessage{Message: m.Message, SenderId: m.SenderId, ReceiverId: m.ReceiverId}
}
