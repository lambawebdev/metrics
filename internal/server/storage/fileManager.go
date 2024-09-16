package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/lambawebdev/metrics/internal/server/config"
)

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file   *os.File
	reader *bufio.Reader
}

func NewConsumer(path string) (*Consumer, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (c *Consumer) ReadEvent() (map[string]interface{}, error) {
	data, err := c.reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}

	if len(data) == 0 {
		data = []byte("{}")
	}

	event := make(map[string]interface{})
	err = json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func GetAllMetrics() (map[string]interface{}, error) {
	err := CreateDir()
	if err != nil {
		return nil, err
	}
	p, err := NewConsumer(config.GetFileStoragePath() + "/metrics.json")

	if err != nil {
		return nil, err
	}

	return p.ReadEvent()
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

func WriteToFile(s *MemStorage) error {
	p, err := NewProducer(config.GetFileStoragePath() + "/metrics.json")

	if err != nil {
		return err
	}

	defer p.Close()

	m := s.GetAll()

	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if _, err = p.writer.Write(data); err != nil {
		return err
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	return p.writer.Flush()
}

func StartToWrite(s *MemStorage, interval uint64) {
	err := CreateDir()

	if err != nil {
		fmt.Println(err)
	}

	storeTicker := time.NewTicker(time.Duration(interval) * time.Second)
	for range storeTicker.C {
		WriteToFile(s)
	}
}

func CreateDir() error {
	err := os.MkdirAll(config.GetFileStoragePath(), 0777)
	if err != nil {
		return err
	}
	return nil
}
