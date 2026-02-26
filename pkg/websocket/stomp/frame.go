// Package stomp 实现 STOMP 协议 over WebSocket
package stomp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// STOMP 命令常量
const (
	CmdConnect     = "CONNECT"
	CmdConnected   = "CONNECTED"
	CmdSubscribe   = "SUBSCRIBE"
	CmdUnsubscribe = "UNSUBSCRIBE"
	CmdSend        = "SEND"
	CmdMessage     = "MESSAGE"
	CmdReceipt     = "RECEIPT"
	CmdError       = "ERROR"
	CmdDisconnect  = "DISCONNECT"
	CmdAck         = "ACK"
	CmdNack        = "NACK"
)

// 常用头字段
const (
	HdrDestination  = "destination"
	HdrContentType  = "content-type"
	HdrSubscription = "subscription"
	HdrMessageId    = "message-id"
	HdrId           = "id"
	HdrAck          = "ack"
	HdrReceipt      = "receipt"
	HdrReceiptId    = "receipt-id"
	HdrVersion      = "version"
	HdrHeartBeat    = "heart-beat"
	HdrAcceptVersion = "accept-version"
	HdrHost         = "host"
	HdrLogin        = "login"
	HdrPasscode     = "passcode"
)

// Frame STOMP 帧
type Frame struct {
	Command string            // 命令
	Headers map[string]string // 头信息
	Body    []byte            // 消息体
}

// NewFrame 创建新帧
func NewFrame(command string) *Frame {
	return &Frame{
		Command: command,
		Headers: make(map[string]string),
		Body:    nil,
	}
}

// AddHeader 添加头信息
func (f *Frame) AddHeader(key, value string) *Frame {
	if f.Headers == nil {
		f.Headers = make(map[string]string)
	}
	f.Headers[key] = value
	return f
}

// GetHeader 获取头信息
func (f *Frame) GetHeader(key string) string {
	if f.Headers == nil {
		return ""
	}
	return f.Headers[key]
}

// SetBody 设置消息体
func (f *Frame) SetBody(body []byte) *Frame {
	f.Body = body
	return f
}

// SetBodyString 设置字符串消息体
func (f *Frame) SetBodyString(body string) *Frame {
	f.Body = []byte(body)
	return f
}

// Marshal 将帧序列化为字节
func (f *Frame) Marshal() ([]byte, error) {
	var buf bytes.Buffer

	// 写入命令
	buf.WriteString(f.Command)
	buf.WriteByte('\n')

	// 写入头信息
	for key, value := range f.Headers {
		// 转义头信息中的特殊字符
		buf.WriteString(escapeHeader(key))
		buf.WriteByte(':')
		buf.WriteString(escapeHeader(value))
		buf.WriteByte('\n')
	}

	// 空行分隔头和消息体
	buf.WriteByte('\n')

	// 写入消息体
	if f.Body != nil {
		buf.Write(f.Body)
	}

	// 终止符
	buf.WriteByte(0)

	return buf.Bytes(), nil
}

// Unmarshal 从字节解析帧
func Unmarshal(data []byte) (*Frame, error) {
	if len(data) == 0 {
		return nil, errors.New("empty frame data")
	}

	reader := bufio.NewReader(bytes.NewReader(data))
	frame := &Frame{
		Headers: make(map[string]string),
	}

	// 读取命令行
	commandLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read command: %w", err)
	}
	frame.Command = strings.TrimSpace(commandLine)

	// 读取头信息
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header: %w", err)
		}

		line = strings.TrimSuffix(line, "\n")

		// 空行表示头信息结束
		if line == "" {
			break
		}

		// 解析键值对
		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue
		}

		key := unescapeHeader(line[:colonIndex])
		value := unescapeHeader(line[colonIndex+1:])
		frame.Headers[key] = value
	}

	// 读取消息体（直到 NULL 字符）
	var body bytes.Buffer
	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read body: %w", err)
		}

		// NULL 字符表示帧结束
		if b == 0 {
			break
		}
		body.WriteByte(b)
	}

	if body.Len() > 0 {
		frame.Body = body.Bytes()
	}

	return frame, nil
}

// escapeHeader 转义头信息中的特殊字符
func escapeHeader(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, ":", "\\c")
	return value
}

// unescapeHeader 反转义头信息
func unescapeHeader(value string) string {
	value = strings.ReplaceAll(value, "\\c", ":")
	value = strings.ReplaceAll(value, "\\r", "\r")
	value = strings.ReplaceAll(value, "\\n", "\n")
	value = strings.ReplaceAll(value, "\\\\", "\\")
	return value
}

// NewConnectedFrame 创建 CONNECTED 帧
func NewConnectedFrame() *Frame {
	frame := NewFrame(CmdConnected)
	frame.AddHeader(HdrVersion, "1.2")
	frame.AddHeader(HdrHeartBeat, "0,0")
	return frame
}

// NewMessageFrame 创建 MESSAGE 帧
func NewMessageFrame(destination, subscriptionId, messageId string, body []byte) *Frame {
	frame := NewFrame(CmdMessage)
	frame.AddHeader(HdrDestination, destination)
	frame.AddHeader(HdrSubscription, subscriptionId)
	frame.AddHeader(HdrMessageId, messageId)
	frame.AddHeader(HdrContentType, "application/json")
	frame.Body = body
	return frame
}

// NewErrorFrame 创建 ERROR 帧
func NewErrorFrame(message string) *Frame {
	frame := NewFrame(CmdError)
	frame.AddHeader("message", message)
	frame.SetBodyString(message)
	return frame
}

// NewReceiptFrame 创建 RECEIPT 帧
func NewReceiptFrame(receiptId string) *Frame {
	frame := NewFrame(CmdReceipt)
	frame.AddHeader(HdrReceiptId, receiptId)
	return frame
}

// ParseHeartBeat 解析心跳配置
func ParseHeartBeat(value string) (sendInterval, recvInterval int, err error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid heart-beat format")
	}

	sendInterval, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid send interval: %w", err)
	}

	recvInterval, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid receive interval: %w", err)
	}

	return sendInterval, recvInterval, nil
}
