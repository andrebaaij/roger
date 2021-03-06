package roger

import (
	"errors"
	"strconv"

	"github.com/senseyeio/roger/constants"
	"github.com/senseyeio/roger/sexp"
)

// Packet is the interface satisfied by objects returned from a R command.
// It contains either the resulting object or an error.
type Packet interface {

	// GetResultObject will parse the packet's contents, returning a go interface{}.
	// If the Packet contains an error this will be returned.
	// If conversion fails, an error will be returned.
	GetResultObject() (interface{}, error)

	// IsError returns a boolean defining whether the Packet contains an error.
	IsError() bool
	// IsOk returns a boolean defining whether Packet was success
	IsOk() bool
}

type packet struct {
	cmd     int
	content []byte
	err     error
}

func newPacket(cmd int, content []byte) Packet {
	return &packet{
		cmd:     cmd,
		content: content,
	}
}

func newErrorPacket(err error) Packet {
	return &packet{
		err: err,
	}
}

func (p *packet) IsError() bool {
	return p.err != nil || p.cmd&15 == 2
}

func (p *packet) IsOk() bool {
	return !p.IsError()
}

func (p *packet) getStatusCode() int {
	return p.cmd >> 24 & 127
}

func (p *packet) getError() error {
	errDescMap := map[int]string{
		2:   "Invalid expression",
		3:   "Parse error",
		127: "Unknown variable/method"}

	if p.err != nil {
		return p.err
	}

	if errDesc, found := errDescMap[p.getStatusCode()]; found {
		return errors.New("Command error with status: " + errDesc)
	}

	return errors.New("Command error with status code: " + strconv.Itoa(p.getStatusCode()))
}

func (p *packet) GetResultObject() (interface{}, error) {
	if p.IsError() {
		return nil, p.getError()
	}
	if len(p.content) == 0 {
		return nil, errors.New("Command failed for an unknown reason")
	}
	isSexp := p.content[0] == byte(constants.DtSexp)
	isLarge := p.content[0] == byte(constants.DtSexp|constants.DtLarge)
	if !isSexp && !isLarge {
		return nil, errors.New("Expected DT_SEXP or DT_LARGE response")
	}
	offset := constants.SmHeaderSize
	if isLarge {
		offset = constants.LgHeaderSize
	}
	return sexp.Parse(p.content[offset:len(p.content)], 0)
}
