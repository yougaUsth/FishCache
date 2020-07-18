package service

import (
	"bytes"
	"encoding/binary"
	"hash/adler32"
	"errors"
)

type Message struct {
	msgID    int32
	msgSize  int32
	data     []byte
	erasureCode uint32
}


func NewMessage(msgID int32, data []byte) *Message {
	msg := &Message{
		msgSize: int32(len(data)) + 4 + 4,
		msgID:   msgID,
		data:    data,
	}
	msg.erasureCode = msg.CalculateErasureCode()
	return msg
}


func (m *Message) GetData() []byte {
	return m.data
}

	func (m *Message) CalculateErasureCode() uint32 {
	if m == nil {
		return 0
	}
	data := new(bytes.Buffer)

	err := binary.Write(data, binary.LittleEndian, m.msgID)
	if err != nil {
		return 0
	}
	err = binary.Write(data, binary.LittleEndian, m.data)
	if err != nil {
		return 0
	}

	return adler32.Checksum(data.Bytes())
}

// Validate erasure code
func (m *Message) Validate() bool {
	if m == nil{
		return false
	}
	return m.erasureCode == m.CalculateErasureCode()
}


// Encode message to []byte
func Encode(msg *Message) ([]byte, error) {

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, msg.msgSize)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, msg.msgID)
	if err != nil{
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, msg.data)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, msg.erasureCode)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil

}

func Decode(data []byte) (*Message, error) {
	bufReader := bytes.NewReader(data)
	dataSize := len(data)

	var msgID int32
	err := binary.Read(bufReader, binary.LittleEndian, &msgID)
	if err != nil {
		return nil, err
	}
	// 读取数据
	dataBufLength := dataSize - 4 - 4
	dataBuf := make([]byte, dataBufLength)
	err = binary.Read(bufReader, binary.LittleEndian, &dataBuf)
	if err != nil {
		return nil, err
	}
	// check erasure code
	var checksum uint32
	err = binary.Read(bufReader, binary.LittleEndian, &checksum)
	if err != nil {
		return nil, err
	}

	message := &Message{}
	message.msgSize = int32(dataSize)
	message.msgID = msgID
	message.data = dataBuf
	message.erasureCode = checksum

	if ! message.Validate() {
		return nil, errors.New("erasure code validate error")
	}

	return message, nil

}




