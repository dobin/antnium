package wingman

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/dobin/antnium/pkg/executor"

	log "github.com/sirupsen/logrus"
)

type WingDirectory struct {
	directory string
}

func MakeWingDirectory() WingDirectory {
	wing := WingDirectory{"C:\\temp\\"}
	return wing
}

func (e WingDirectory) Start(destination string) {
	e.loop()
}

func (dl WingDirectory) loop() {
	r, _ := regexp.Compile(".*\\.dwn")
	for {
		files, err := ioutil.ReadDir(dl.directory)
		if err != nil {
			log.Error(err)
			return
		}

		for _, f := range files {
			if r.MatchString(f.Name()) {
				path := dl.directory + f.Name()
				err := dl.handleFile(path)
				if err != nil {
					log.Errorf("When handling file: %s: %s", path, err.Error())
				}

				// Delete file always
				os.Remove(path)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (dl WingDirectory) handleFile(filename string) error {
	executor := executor.MakeExecutor()

	// Handle it
	log.Infof("Reading file: %s", filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	fmt.Println("WingDirectory: HandleFile(): Read: ")

	packet, err := DecodePacket(string(data))
	if err != nil {
		return fmt.Errorf("Wingman: Error: %s", err.Error())
	}

	// Execute
	packet, err = executor.Execute(packet)
	if err != nil {
		log.WithFields(log.Fields{
			"packet": packet,
			"error":  err,
		}).Info("Wingman: Error executing packet")

		// TODO ERR
	}

	// Write answer
	packetEncoded, err := EncodePacket(packet)
	if err != nil {
		log.Error("Wingman: Error: ", err.Error())
	}
	path := dl.directory + "1.pu"
	err = os.WriteFile(path, packetEncoded, 0644)
	if err != nil {
		return err
	}
	log.Infof("Finished writing answer: %s", packetEncoded)

	return nil
}
