package service

import (
	"net"
	"time"
	"context"
	"io"
	"bytes"
	"encoding/binary"
)

type Conn struct {
	sid        string
	name       string
	rawConn    net.Conn
	sendCh     chan []byte
	done       chan error
	hbTimer    *time.Timer
	messageCh  chan *Message
	hbInterval time.Duration
	hbTimeout  time.Duration
}

func NewConn(c net.Conn, hbInterval time.Duration, hbTimeout time.Duration) *Conn{
	conn := &Conn{
		rawConn:    c,
		sendCh:     make(chan []byte, 100),
		done:       make(chan error),
		messageCh:  make(chan *Message, 100),
		hbInterval: hbInterval,
		hbTimeout:  hbTimeout,
	}

	conn.name = c.RemoteAddr().String()
	conn.hbTimer = time.NewTimer(conn.hbInterval)

	if conn.hbInterval == 0 {
		conn.hbTimer.Stop()
	}

	return conn
}

func (c *Conn) Close () {
	c.hbTimer.Stop()
	c.rawConn.Close()
}

func (c *Conn) SendMessage(msg *Message) error {
	pkg, err := Encode(msg)
	if err != nil {
		return err
	}
	c.sendCh <- pkg
	return nil
}

func (c *Conn) read (ctx context.Context) {
	for {
		select {

		case <-ctx.Done():
			return
		default:
			// Set time out
			if c.hbInterval > 0 {
				err := c.rawConn.SetReadDeadline(time.Now().Add(c.hbTimeout))
				if err != nil {
					c.done <- err
					continue
				}
			}
			// Get data size
			buf := make([]byte, 4)
			_, err := io.ReadFull(c.rawConn, buf)
			if err != nil {
				c.done <- err
				continue
			}
			var dataSize int32
			bufReader := bytes.NewReader(buf)
			err = binary.Read(bufReader, binary.LittleEndian, &dataSize)
			if err != nil{
				c.done <- err
				continue
			}
			// Read data from conn stream
			dataBuf := make([]byte, dataSize)
			_, err = io.ReadFull(c.rawConn, dataBuf)
			if err != nil {
				c.done <- err
				continue
			}
			// Decode message
			msg, err := Decode(dataBuf)
			if err != nil {
				c.done <- err
				continue
			}

			c.messageCh <- msg

		}
	}
}

func (c *Conn) write (ctx context.Context) {
	hbData := make([]byte, 0)
	for {
		select {
		case <- ctx.Done():
			return
		case pkt := <- c.sendCh:
			if pkt == nil {
				continue
			}
			_, err := c.rawConn.Write(pkt)
			if err != nil {
				c.done <- err
			}
		case <- c.hbTimer.C:
			hbMessage := NewMessage(MsgHeartbeat, hbData)
			c.SendMessage(hbMessage)
			// 设置心跳timer
			if c.hbInterval > 0 {
				c.hbTimer.Reset(c.hbInterval)
			}
		}
	}
}

