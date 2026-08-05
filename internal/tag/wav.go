package tag

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/bogem/id3v2/v2"
)

type wavTagger struct {
	path string
}

type wavChunk struct {
	id      string
	payload []byte
}

type infoField struct {
	id    string
	value string
}

func (t *wavTagger) write(m Metadata) error {
	id3Chunk, err := buildID3Chunk(m)
	if err != nil {
		return err
	}

	var chunks []wavChunk
	if info := buildInfoChunk(m); info != nil {
		chunks = append(chunks, wavChunk{id: "LIST", payload: info})
	}
	if id3Chunk != nil {
		chunks = append(chunks, wavChunk{id: "id3 ", payload: id3Chunk})
	}

	return rewriteWAV(t.path, chunks)
}

func buildID3Chunk(m Metadata) ([]byte, error) {
	tag := id3v2.NewEmptyTag()
	applyID3Frames(tag, m)
	if !tag.HasFrames() {
		return nil, nil
	}

	var buf bytes.Buffer
	if _, err := tag.WriteTo(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func buildInfoChunk(m Metadata) []byte {
	fields := []infoField{
		{"INAM", m.Title},
		{"IART", m.Artists},
		{"IGNR", m.Genre},
		{"ITRK", m.TrackNumber},
	}

	if m.Album != nil {
		date := m.Album.ReleaseDate
		if parts := strings.Split(date, "-"); len(parts) == 3 {
			date = parts[0]
		}

		fields = append(fields,
			infoField{"IPRD", m.Album.Title},
			infoField{"ICRD", date},
			infoField{"ICMT", m.Album.ProducerLine},
			infoField{"ICOP", m.Album.Copyright},
		)
	}

	var buf bytes.Buffer
	buf.WriteString("INFO")

	for _, field := range fields {
		if field.value == "" {
			continue
		}
		writeChunk(&buf, field.id, append([]byte(field.value), 0))
	}

	if buf.Len() == 4 {
		return nil
	}

	return buf.Bytes()
}

func writeChunk(w io.Writer, id string, payload []byte) {
	header := make([]byte, 0, 8)
	header = append(header, id...)
	header = binary.LittleEndian.AppendUint32(header, uint32(len(payload)))

	w.Write(header)
	w.Write(payload)
	if len(payload)%2 != 0 {
		w.Write([]byte{0})
	}
}

func rewriteWAV(path string, chunks []wavChunk) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	header := make([]byte, 12)
	if _, err := io.ReadFull(src, header); err != nil {
		return err
	}
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return errors.New("not a wav file")
	}

	tmpPath := path + ".tmp"
	dst, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	done := false
	defer func() {
		if !done {
			dst.Close()
			os.Remove(tmpPath)
		}
	}()

	if _, err := dst.Write(header); err != nil {
		return err
	}

	size, err := copyChunks(dst, src)
	if err != nil {
		return err
	}

	for _, chunk := range chunks {
		var buf bytes.Buffer
		writeChunk(&buf, chunk.id, chunk.payload)
		if _, err := dst.Write(buf.Bytes()); err != nil {
			return err
		}
		size += int64(buf.Len())
	}

	riffSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(riffSize, uint32(size))
	if _, err := dst.WriteAt(riffSize, 4); err != nil {
		return err
	}

	if err := dst.Sync(); err != nil {
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	done = true

	return nil
}

func copyChunks(dst io.Writer, src io.Reader) (int64, error) {
	size := int64(4)
	head := make([]byte, 8)

	for {
		if _, err := io.ReadFull(src, head); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return size, nil
			}
			return size, err
		}

		id := string(head[0:4])
		payloadSize := int64(binary.LittleEndian.Uint32(head[4:8]))

		if id == "id3 " || id == "ID3 " {
			if err := skipPayload(src, payloadSize); err != nil {
				return size, err
			}
			continue
		}

		if id == "LIST" {
			payload := make([]byte, payloadSize)
			if _, err := io.ReadFull(src, payload); err != nil {
				return size, err
			}
			if err := skipPad(src, payloadSize); err != nil {
				return size, err
			}
			if bytes.HasPrefix(payload, []byte("INFO")) {
				continue
			}

			var buf bytes.Buffer
			writeChunk(&buf, id, payload)
			if _, err := dst.Write(buf.Bytes()); err != nil {
				return size, err
			}
			size += int64(buf.Len())

			continue
		}

		if _, err := dst.Write(head); err != nil {
			return size, err
		}
		if _, err := io.CopyN(dst, src, payloadSize); err != nil {
			return size, err
		}
		size += 8 + payloadSize

		if payloadSize%2 != 0 {
			if _, err := dst.Write([]byte{0}); err != nil {
				return size, err
			}
			size++
			if err := skipPad(src, payloadSize); err != nil {
				return size, err
			}
		}
	}
}

func skipPayload(src io.Reader, payloadSize int64) error {
	if _, err := io.CopyN(io.Discard, src, payloadSize); err != nil {
		return err
	}
	return skipPad(src, payloadSize)
}

func skipPad(src io.Reader, payloadSize int64) error {
	if payloadSize%2 == 0 {
		return nil
	}
	if _, err := io.CopyN(io.Discard, src, 1); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
