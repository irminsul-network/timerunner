package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	config           map[string]string
	executablesPath  string
	descriptionsPath string
)

func ensureConfig() error {
	configFile, err := os.ReadFile("conf.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return err
	}

	return nil
}

func ensureDirectories() error {
	// holds executable list or folders containing executables of the same name
	executablesPath = path.Join(config["dataDir"], "executables")
	stat, err := os.Stat(executablesPath)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("expected %s to be a directory", executablesPath)
	}

	// contains descriptions of how to run the executable, exe id , exe name , frequency., start time (if omitted, run immediately)
	descriptionsPath = path.Join(config["dataDir"], "descriptions.json")
	stat, err = os.Stat(descriptionsPath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("expected %s to be a file", descriptionsPath)
	}

	return nil
}

func ensureAll() error {

	log.Println("ensuring conf.json is present...")
	if err := ensureConfig(); err != nil {
		return fmt.Errorf("failed to ensure config file err: %s", err)
	}

	log.Println("ensuring required  data directory structure is present")
	if err := ensureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure required directory structure, err: %s", err)
	}
	return nil
}

func main() {
	log.Println("timerunner booting up...")
	err := ensureAll()
	if err != nil {
		log.Fatalf("failed to ensure data dir's file structure is as expected. err: %s", err)
	}

	runs := loadRunInfo()

	//wg := sync.WaitGroup{}
	//wg.Add(1)
	scheduler := NewScheduler(runs)
	scheduler.ScheduleAllRuns()

	//wg.Wait()

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		runInfo, err := AddExecutable(r.Body)
		if err != nil {
			log.Println(err)
		}
		scheduler.scheduleRun(*runInfo)

		w.WriteHeader(http.StatusCreated)
	})

	http.ListenAndServe(":3004", nil)

	//executable, _ := os.ReadFile("demo.zip")
	//exeReader := bytes.NewReader(executable)

}

// so first let users upload binaries, give them an id. yo get 201 http.
// application will only have one primary config files (future projects form now on should follow this pattern)
// that configures all other parameters.
