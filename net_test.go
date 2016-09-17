package main

import (
	"bytes"
	"net/textproto"

	"golang.org/x/net/context"

	. "testing"
)

type mockedReadWriteCloser struct {
	Input  *bytes.Buffer
	Closed bool
	output bytes.Buffer
}

func (m *mockedReadWriteCloser) Close() error {
	m.Closed = true
	return nil
}

func (m *mockedReadWriteCloser) Read(b []byte) (int, error) {
	return m.Input.Read(b)
}

func (m *mockedReadWriteCloser) Write(b []byte) (int, error) {
	return m.output.Write(b)
}

type inputOutputTest struct {
	mrwc *mockedReadWriteCloser
}

func testInput(input string) inputOutputTest {
	m := mockedReadWriteCloser{
		bytes.NewBufferString(input),
		false,
		bytes.Buffer{},
	}
	return inputOutputTest{&m}
}

func (iot inputOutputTest) ExpectingOutput(t *T, expected string) {
	ctx, cancel := context.WithCancel(context.Background())
	ids := generateIds(ctx)
	srv := newServer(ids)

	ch := connectionHandler{
		srv,
		ctx,
		cancel,
		textproto.NewConn(iot.mrwc),
	}
	ch.Handle()

	if !iot.mrwc.Closed {
		t.Error("Connection was not closed.")
	}

	if output := iot.mrwc.output.String(); output != expected {
		t.Errorf("Unexpected output. Output: %s Expected: %s", output, expected)
	}

	select {
	case <-ctx.Done():
	default:
		t.Error("Context wasn't done.")
	}
}

func TestPut(t *T) {
	testInput("put 0 0 10 5\r\nhello\r\n").ExpectingOutput(t, "INSERTED 1\r\n")
}

func TestUnknownCommand(t *T) {
	testInput("this is a test\r\n").ExpectingOutput(t, "UNKNOWN_COMMAND\r\n")
}
