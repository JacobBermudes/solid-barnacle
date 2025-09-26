package vpncoder

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"io"
	"strings"
)

type VpnConfig struct {
	Config string
	Key    string
}

func qCompress(data []byte, level int) ([]byte, error) {
	var b bytes.Buffer
	// Write uncompressed size as 4-byte big-endian
	if err := binary.Write(&b, binary.BigEndian, uint32(len(data))); err != nil {
		return nil, err
	}
	// Create zlib writer with specified compression level
	w, err := zlib.NewWriterLevel(&b, level)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(data)
	if err != nil {
		w.Close()
		return nil, err
	}
	w.Close()
	return b.Bytes(), nil
}

func qUncompress(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, nil
	}
	// Read uncompressed size
	r := bytes.NewReader(data[:4])
	var uncompressedSize uint32
	if err := binary.Read(r, binary.BigEndian, &uncompressedSize); err != nil {
		return nil, err
	}
	// Decompress the data
	compressedData := data[4:]
	zr, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, nil
	}
	defer zr.Close()

	uncompressedData, err := io.ReadAll(zr)
	if err != nil {
		return nil, nil
	}
	if len(uncompressedData) != int(uncompressedSize) {
		return nil, nil
	}
	return uncompressedData, nil
}

func base64urlEncode(data []byte) string {
	encoded := base64.URLEncoding.EncodeToString(data)
	return strings.TrimRight(encoded, "=")
}

func base64urlDecode(data string) ([]byte, error) {
	// Add padding if needed
	padding := len(data) % 4
	if padding > 0 {
		data += strings.Repeat("=", 4-padding)
	}
	return base64.URLEncoding.DecodeString(data)
}

func (v VpnConfig) Encode() (string, error) {
	dataBytes := []byte(v.Config)
	compressed, err := qCompress(dataBytes, 8)
	if err != nil {
		return "", err
	}
	encoded := base64urlEncode(compressed)
	return "vpn://" + encoded, nil
}

func (v VpnConfig) Decode() (string, error) {
	data := strings.TrimPrefix(v.Key, "vpn://")
	compressed, err := base64urlDecode(data)
	if err != nil {
		return "", err
	}
	uncompressed, err := qUncompress(compressed)
	if err != nil {
		return "", err
	}
	if uncompressed == nil {
		return string(compressed), nil
	}
	return string(uncompressed), nil
}
