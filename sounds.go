package gameover

import (
	"fmt"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
	"io"
	"log"
	"os"
)

const (
	OneMin          = 60
	FiveMin         = OneMin * 5
	OneMinSoundLoc  = "resources/1minute.mp3"
	FiveMinSoundLoc = "resources/5minutes.mp3"
)

type Foley struct {
	soundPlayer *oto.Player
}

func (f *Foley) playSound(fileLocation string, newContext bool) error {
	file, err := os.Open(fileLocation)
	if err != nil {
		return err
	}
	defer file.Close()

	d, err := mp3.NewDecoder(file)
	if err != nil {
		return err
	}
	if newContext {
		f.soundPlayer, err = oto.NewPlayer(d.SampleRate(), 2, 2, 8192)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Length: %d[bytes]\n", d.Length())

	if _, err := io.Copy(f.soundPlayer, d); err != nil {
		return err
	}
	return nil
}

func (f *Foley) OnNotify(event Event) {
	switch event.Type {
	case TimeRemaining:
		f.checkForImportantTimeRemainingEvents(event.Data)
	}
}

func (f *Foley) checkForImportantTimeRemainingEvents(secondsLeft int64) {
	if secondsLeft == FiveMin {
		go func() {
			log.Println("play 5 min sound")
			f.playSound(FiveMinSoundLoc, true)
		}()
	} else if secondsLeft == OneMin {
		go func() {
			log.Println("play 1 min sound")
			f.playSound(OneMinSoundLoc, false)
		}()
	}
}
