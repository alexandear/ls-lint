package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

func getFullPath(path string) string {
	return fmt.Sprintf("%s%s%s", root, sep, path)
}

func main() {
	var exitCode = 1
	var writer = os.Stderr
	var flags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var warn = flags.Bool("warn", false, "treat lint errors as warnings; write output to stdout and return exit code 0")

	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	var filesystem = os.DirFS(root)

	var config = &Config{
		RWMutex: new(sync.RWMutex),
	}

	var linter = &Linter{
		Statistic: nil,
		Errors:    make([]*Error, 0),
		RWMutex:   new(sync.RWMutex),
	}

	// open config file
	file, err := os.Open(".ls-lint.yml")

	if err != nil {
		log.Fatal(err)
	}

	// close file
	defer func() {
		err = file.Close()

		if err != nil {
			log.Fatal(err)
		}
	}()

	// read file
	configBytes, err := ioutil.ReadAll(file)

	if err != nil {
		log.Fatal(err)
	}

	// to yaml
	err = yaml.Unmarshal(normalizeConfig(configBytes, byte(runeUnixSep), byte(runeSep)), &config)

	if err != nil {
		log.Fatal(err)
	}

	// runner
	if err := linter.Run(filesystem, config, false, false); err != nil {
		log.Fatal(err)
	}

	// errors
	errors := linter.getErrors()

	// no errors
	if len(errors) == 0 {
		os.Exit(exitCode)
	}

	if *warn {
		writer = os.Stdout
		exitCode = 0
	}

	logger := log.New(writer, "", log.LstdFlags)

	// with errors
	for _, err := range linter.getErrors() {
		var ruleMessages []string

		for _, rule := range err.getRules() {
			ruleMessages = append(ruleMessages, rule.GetErrorMessage())
		}

		logger.Printf("%s failed for rules: %s", err.getPath(), strings.Join(ruleMessages, fmt.Sprintf(" %s ", or)))
	}

	os.Exit(exitCode)
}
