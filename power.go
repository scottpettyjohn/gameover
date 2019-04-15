package gameover

import (
	"github.com/stianeikeland/go-rpio"
	"log"
	"runtime"
)

type (
	Plug interface {
		Open() error
		Close() error
		On()
		Off()
	}

	mockPlug struct {
	}

	rpiPlug struct {
		pin rpio.Pin
	}
)

func PowerFactory() Plug {
	if runtime.GOOS == "linux" {
		return &rpiPlug{}
	} else {
		return &mockPlug{}
	}
}

// mockPlug method implementation
func (p mockPlug) Open() error {
	return nil
}

func (p mockPlug) Close() error {
	return nil
}

func (p mockPlug) On() {
	log.Print("click...on")
}

func (p mockPlug) Off() {
	log.Print("click...off")
}

// rpiPlug implementation
func (p *rpiPlug) Open() error {
	if err := rpio.Open(); err != nil {
		return err
	}
	p.pin = rpio.Pin(21)
	p.pin.Output()
	return nil
}

func (p *rpiPlug) Close() error {
	return rpio.Close()
}

func (p *rpiPlug) On() {
	p.pin.High()
}

func (p *rpiPlug) Off() {
	p.pin.Low()
}
