package redis

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var ErrNil = errors.New("redis: nil")

type Client struct {
	address  string
	password string
	database int
	timeout  time.Duration
}

func NewClient(host, port, password string, database int, timeout time.Duration) *Client {
	return &Client{
		address:  net.JoinHostPort(host, port),
		password: password,
		database: database,
		timeout:  timeout,
	}
}

func (c *Client) Do(ctx context.Context, args ...string) (interface{}, error) {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(c.timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)

	reader := bufio.NewReader(conn)
	if c.password != "" {
		if err := writeCommand(conn, "AUTH", c.password); err != nil {
			return nil, err
		}
		if _, err := readReply(reader); err != nil {
			return nil, err
		}
	}
	if c.database > 0 {
		if err := writeCommand(conn, "SELECT", strconv.Itoa(c.database)); err != nil {
			return nil, err
		}
		if _, err := readReply(reader); err != nil {
			return nil, err
		}
	}

	if err := writeCommand(conn, args...); err != nil {
		return nil, err
	}
	return readReply(reader)
}

func writeCommand(conn net.Conn, args ...string) error {
	var builder strings.Builder
	builder.WriteString("*")
	builder.WriteString(strconv.Itoa(len(args)))
	builder.WriteString("\r\n")
	for _, arg := range args {
		builder.WriteString("$")
		builder.WriteString(strconv.Itoa(len(arg)))
		builder.WriteString("\r\n")
		builder.WriteString(arg)
		builder.WriteString("\r\n")
	}
	_, err := conn.Write([]byte(builder.String()))
	return err
}

func readReply(reader *bufio.Reader) (interface{}, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")

	switch prefix {
	case '+':
		return line, nil
	case '-':
		return nil, errors.New(line)
	case ':':
		value, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return nil, err
		}
		return value, nil
	case '$':
		size, err := strconv.Atoi(line)
		if err != nil {
			return nil, err
		}
		if size == -1 {
			return nil, ErrNil
		}
		payload := make([]byte, size+2)
		if _, err := reader.Read(payload); err != nil {
			return nil, err
		}
		return string(payload[:size]), nil
	default:
		return nil, fmt.Errorf("unsupported redis reply prefix %q", prefix)
	}
}
