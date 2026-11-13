package videobuffer

import (
	"bytes"
	"errors"
	"fmt"
	"os"
)

const (
	tsPacketSize = 188
	tsSyncByte   = 0x47
)

type VideoFrame struct {
	Index   int
	PTS     uint64
	DTS     uint64
	Payload []byte
}

func ExtractFramesFromTS(filePath string) ([][]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TS file: %w", err)
	}
	frames, err := ParseTSVideoFrames(data)
	if err != nil {
		return nil, err
	}
	if len(frames) == 0 {
		return nil, errors.New("no video frames found in TS file")
	}
	rawFrames := make([][]byte, len(frames))
	for i, f := range frames {
		rawFrames[i] = f.Payload
	}
	return rawFrames, nil
}

func ParseTSVideoFrames(data []byte) ([]VideoFrame, error) {
	if len(data) < tsPacketSize {
		return nil, errors.New("data too short for MPEG-TS")
	}
	var videoPID uint16 = 0xFFFF
	var currentPES []byte
	var frames []VideoFrame
	for offset := 0; offset+tsPacketSize <= len(data); offset += tsPacketSize {
		pkt := data[offset : offset+tsPacketSize]
		if pkt[0] != tsSyncByte {
			for offset < len(data) && data[offset] != tsSyncByte {
				offset++
			}
			if offset+tsPacketSize > len(data) {
				break
			}
			pkt = data[offset : offset+tsPacketSize]
		}
		pusi := (pkt[1] & 0x40) != 0
		pid := (uint16(pkt[1]&0x1F) << 8) | uint16(pkt[2])
		afc := (pkt[3] >> 4) & 0x03
		payloadOffset := 4
		if afc == 2 {
			continue
		} else if afc == 3 {
			adaptLen := int(pkt[4])
			payloadOffset = 5 + adaptLen
			if payloadOffset >= tsPacketSize {
				continue
			}
		}
		payload := pkt[payloadOffset:]
		if pusi && len(payload) >= 4 {
			if payload[0] == 0x00 && payload[1] == 0x00 && payload[2] == 0x01 {
				streamID := payload[3]
				if streamID >= 0xE0 && streamID <= 0xEF {
					if videoPID == 0xFFFF {
						videoPID = pid
					}
				}
			}
		}
		if pid == videoPID {
			if pusi {
				if len(currentPES) > 0 {
					framePayload := extractESFromPES(currentPES)
					if len(framePayload) > 0 {
						frames = append(frames, VideoFrame{
							Index:   len(frames),
							Payload: framePayload,
						})
					}
					currentPES = nil
				}
			}
			currentPES = append(currentPES, payload...)
		}
	}
	if len(currentPES) > 0 {
		framePayload := extractESFromPES(currentPES)
		if len(framePayload) > 0 {
			frames = append(frames, VideoFrame{
				Index:   len(frames),
				Payload: framePayload,
			})
		}
	}
	return frames, nil
}

func extractESFromPES(pes []byte) []byte {
	if len(pes) < 9 {
		return pes
	}
	if pes[0] == 0x00 && pes[1] == 0x00 && pes[2] == 0x01 {
		headerLen := int(pes[8])
		payloadStart := 9 + headerLen
		if payloadStart < len(pes) {
			return pes[payloadStart:]
		}
	}
	return pes
}

func SplitNALUnits(es []byte) [][]byte {
	var nals [][]byte
	startCode3 := []byte{0x00, 0x00, 0x01}
	start := 0
	for start < len(es) {
		idx := bytes.Index(es[start:], startCode3)
		if idx == -1 {
			if start < len(es) {
				nals = append(nals, es[start:])
			}
			break
		}
		absIdx := start + idx
		if absIdx > start {
			end := absIdx
			if end > 0 && es[end-1] == 0x00 {
				end--
			}
			nals = append(nals, es[start:end])
		}
		start = absIdx + 3
		if absIdx > 0 && es[absIdx-1] == 0x00 {
			// был 4-байтовый
		}
	}
	return nals
}
