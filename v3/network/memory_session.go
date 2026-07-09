package network

import (
	"bytes"
	"fmt"
)

type MemorySession struct {
	basicSession
}

func NewMemorySession(inputBuffer, outputBuffer []byte, sessionProp SessionProperties) *MemorySession {
	ret := &MemorySession{
		basicSession: basicSession{
			SessionProperties: sessionProp,
		},
	}
	ret.terminal = ret
	ret.inBuffer = bytes.NewBuffer(inputBuffer)
	ret.outBuffer = bytes.NewBuffer(outputBuffer)
	return ret
}

func (ms *MemorySession) read(length int) ([]byte, error) {
	output := make([]byte, length)
	n, err := ms.inBuffer.Read(output)
	if err != nil {
		return nil, err
	}
	if n != length {
		return nil, fmt.Errorf("the read bytes: %d is lower than requested length: %d", n, length)
	}
	return output, nil
}
func (ms *MemorySession) Write() error {
	return nil
}
func (ms *MemorySession) ResetBuffer() {
	ms.inBuffer.Reset()
	ms.outBuffer.Reset()
}

func (ms *MemorySession) GetProperties() SessionProperties {
	return ms.SessionProperties
}
