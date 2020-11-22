package util

import (
	"bufio"
	"context"
	"encoding/binary"
	"github.com/gogo/protobuf/proto"
)

func SizeDelimitedReader(ctx context.Context, r *bufio.Reader, msg proto.Message) error {
	msgSizeBuf := make([]byte, 8)
	_, err := r.Read(msgSizeBuf)
	if err != nil {
		return err
	}
	msgSize := binary.LittleEndian.Uint64(msgSizeBuf)
	msgBuf := make([]byte, msgSize)
	_, err = r.Read(msgBuf)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(msgBuf, msg)
	if err != nil {
		return err
	}
	return nil

}
func SizeDelimtedWriter(ctx context.Context, w *bufio.Writer, msg proto.Message) error {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgSize := len(msgBytes)
	msgBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(msgBuf, uint64(msgSize))
	msgBuf = append(msgBuf, msgBytes...)
	_, err = w.Write(msgBuf)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	return nil

}
