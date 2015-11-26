package main

import (
	"log"
	"bufio"
	"fmt"
	"os"
	"strings"
	"bytes"
	"github.com/cznic/b"
	"github.com/op/go-logging"
	"github.com/voxelbrain/goptions"
	"github.com/gin-gonic/gin"
	util "github.com/woanware/goutil"
	"gopkg.in/yaml.v2"
	"sync"
)

// ##### Variables ###########################################################

var (
	logger 	*logging.Logger
	config  *Config
	bTree 	*b.Tree
	opt		*Options
)

// ##### Constants  ###########################################################

const APP_TITLE string = "NSRL Server"
const APP_NAME string = "nsrls"
const APP_VERSION string = "1.00"

// Application modes
const (
	MODE_FILE 	= "f"
	MODE_SERVER	= "s"
)

// Formats for output
const (
	FORMAT_ALL 			= "a"
	FORMAT_IDENTIFIED 	= "i"
	FORMAT_UNIDENTIFIED = "u"
)

// ##### Methods  #############################################################

// Application entry point
func main() {

	fmt.Printf("\n%s (%s) %s\n\n", APP_TITLE, APP_NAME, APP_VERSION)

	// Setup the logging infrastructure
	initialiseLogging()

	// Setup some default values
	opt = new(Options)
	opt.ConfigFile = "./" + APP_NAME + ".config"
	opt.CsvField = -1
	opt.CsvDelimiter = ","
	opt.RemoveQuotes = false
	opt.Format = FORMAT_ALL // Default to all output

	goptions.ParseAndFail(opt)

	// Increment the CSV field to make it easier for the user, since our arrays are 0 based
	if opt.CsvField > -1 {
		opt.CsvField -= 1
	}

	// Validate the mode value
	switch (opt.Mode) {
	case MODE_SERVER:
		// Load the applications configuration such as ports and IP
		config = loadConfig(opt.ConfigFile, true)
	case MODE_FILE:
		if len(opt.InputFile) == 0 {
			logger.Fatal("Input file path must be supplied when in file mode")
		}

		if len(opt.OutputFile) == 0 {
			logger.Fatal("Output file path must be supplied when in file mode")
		}

		// Load the applications configuration such as ports and IP
		config = loadConfig(opt.ConfigFile, false)
	default:
		logger.Fatal("Invalid mode value (m): %v", opt.Mode)
	}

	// Validate the format value
	switch (opt.Format) {
	case FORMAT_UNIDENTIFIED:
	case FORMAT_IDENTIFIED:
	case FORMAT_ALL:
	default:
		logger.Fatal("Invalid format value (f): %v", opt.Format)
	}

	// Lets make sure that the users input file actually exists
	if _, err := os.Stat(opt.DataFile); os.IsNotExist(err) {
		logger.Fatal("Data file does not exist")
	}

	processDataFile(opt.DataFile)

	// Start the web API interface if the user wants it running
	if opt.Mode == "s" {
		logger.Info("HTTP API server running: " + config.ApiIp + ":" + fmt.Sprintf("%d", config.ApiPort))
		go func() {
			var r *gin.Engine
			if config.ShowRequests == true {
				r = gin.Default()
			} else {
				gin.SetMode(gin.ReleaseMode)
				r = gin.New()

				r.Use(gin.Recovery())
			}

			r.GET("/single/:hash/", lookupSingleHash)
			r.POST("/bulk", lookupMultipleHashes)
			r.Run(config.ApiIp + ":" + fmt.Sprintf("%d", config.ApiPort))
		}()

		var wg sync.WaitGroup
		wg.Add(1)
		wg.Wait()
	} else {
		processInputFile()
	}
}

// Import the import data file into the BTree
func processDataFile(inputFile string) {

	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatal("Error opening the data file: %v", err)
	}
	defer file.Close()

	// String array used when importing from a CSV file
	var parts []string

	logger.Info("Starting import")

	// Initialise the BTree structure
	bTree = b.TreeNew(cmp)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if opt.CsvField > -1 {
			parts = strings.Split(scanner.Text(), opt.CsvDelimiter)

			if opt.CsvField > len(parts) {
				logger.Error("CSV field index is greater than the length of the split parts: %v", scanner.Text())
				continue
			}

			if opt.RemoveQuotes == true {
				bTree.Set(strings.ToUpper(parts[opt.CsvField][1:len(parts[opt.CsvField])-1]), nil)
			} else {
				bTree.Set(strings.ToUpper(parts[opt.CsvField]), nil)
			}
		} else {
			if opt.RemoveQuotes == true {
				bTree.Set(strings.ToUpper(scanner.Text()[1:len(scanner.Text())-1]), nil)
			} else {
				bTree.Set(strings.ToUpper(scanner.Text()), nil)
			}
		}
	}

	logger.Info("Import complete")

	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}
}

// Process the input file
func processInputFile() {

	fileInput, err := os.Open(opt.InputFile)
	if err != nil {
		log.Fatal("Error opening the input file: %v", err)
	}
	defer fileInput.Close()

	fileOutput, err := os.Create(opt.OutputFile)
	if err != nil {
		log.Fatal("Error creating the output file: %v", err)
	}
	defer fileOutput.Close()

	// Output some CSV file headers
	switch (opt.Format) {
	case FORMAT_IDENTIFIED: // Found
		fileOutput.Write([]byte(fmt.Sprintf("%s\n", "Hash")))
	case FORMAT_UNIDENTIFIED: // Not Found
		fileOutput.Write([]byte(fmt.Sprintf("%s\n", "Hash")))
	case FORMAT_ALL: // All
		fileOutput.Write([]byte(fmt.Sprintf("%s,%s\n", "Hash", "Status")))
	}

	logger.Info("Starting processing")

	var ret bool

	scanner := bufio.NewScanner(fileInput)
	for scanner.Scan() {
		// Check if the hash exists in the BTree
		_, ret = bTree.Get(strings.ToUpper(scanner.Text()))

		if ret == true {
			switch (opt.Format) {
			case FORMAT_IDENTIFIED: // Found
				fileOutput.Write([]byte(fmt.Sprintf("%s,%s\n", strings.ToUpper(scanner.Text()), "FOUND")))
			case FORMAT_UNIDENTIFIED: // Not Found
				// Ignore
			case FORMAT_ALL: // All
				fileOutput.Write([]byte(fmt.Sprintf("%s,%s\n", strings.ToUpper(scanner.Text()), "FOUND")))
			}
		} else {
			switch (opt.Format) {
			case FORMAT_IDENTIFIED: // Found
				// Ignore
			case FORMAT_UNIDENTIFIED: // Not Found
				fileOutput.Write([]byte(fmt.Sprintf("%s,%s\n", strings.ToUpper(scanner.Text()), "NOT FOUND")))
			case FORMAT_ALL: // All
				fileOutput.Write([]byte(fmt.Sprintf("%s,%s\n", strings.ToUpper(scanner.Text()), "NOT FOUND")))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}

	logger.Info("Processing complete")
}

// Function used by the BTree library to implement the comparison
func cmp(a, b interface{}) int {
	return bytes.Compare([]byte(a.(string)), []byte(b.(string)))
}

// Loads the applications config file contents (yaml) and marshals to a struct
func loadConfig(configPath string, runServer bool) (*Config) {
	c := new(Config)
	data, err := util.ReadTextFromFile(configPath)
	if err != nil {
		logger.Fatal("Error reading the config file: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		logger.Fatal("Error unmarshalling the config file: %v", err)
	}

	if runServer == true {
		if len(c.ApiIp) == 0 {
			logger.Fatal("API IP not set in config file")
		}

		if c.ApiPort == 0 {
			logger.Fatal("API port not set in config file")
		}
	}

	return c
}

// Sets up the logging infrastructure e.g. Stdout and /var/log
func initialiseLogging() {
	// Setup the actual loggers
	logger = logging.MustGetLogger(APP_NAME)

	// Check that we have a "nca" sub directory in /var/log
	if _, err := os.Stat("/var/log/" + APP_NAME); os.IsNotExist(err) {
		logger.Fatal("The /var/log/%s directory does not exist", APP_NAME)
	}

	// Check that we have permission to write to the /var/log/APP_NAME directory
	f, err := os.Create("/var/log/" + APP_NAME + "/test.txt")
	if err != nil {
		logger.Fatal("Unable to write to /var/log/%s", APP_NAME)
	}

	// Clear up our tests
	os.Remove("/var/log/" + APP_NAME + "/test.txt")
	f.Close()

	// Define the /var/log file
	logFile, err := os.OpenFile("/var/log/" + APP_NAME + "/log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Fatal("Error opening the log file: %v", err)
	}

	// Define the StdOut loggingDatabaser
	backendStdOut := logging.NewLogBackend(os.Stdout, "", 0)
	formatStdOut:= logging.MustStringFormatter(
		"%{color}%{time:2006-01-02T15:04:05.000} %{color:reset} %{message}",)
	formatterStdOut := logging.NewBackendFormatter(backendStdOut, formatStdOut)

	// Define the /var/log logging
	backendFile := logging.NewLogBackend(logFile, "", 0)
	formatFile:= logging.MustStringFormatter(
		"%{time:2006-01-02T15:04:05.000} %{level:.4s} %{message}",)
	formatterFile := logging.NewBackendFormatter(backendFile, formatFile)

	logging.SetBackend(formatterStdOut, formatterFile)
}

