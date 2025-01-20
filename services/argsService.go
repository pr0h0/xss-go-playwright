package services

import (
	"errors"
	"os"
	"strings"

	"github.com/akamensky/argparse"
)

// ArgKey defines the enum type for argument keys
type ArgKey string

// Predefined keys for the argument map
const (
	keyMethod   ArgKey = "method"
	keyProtocol ArgKey = "protocol"
	keyVerbose  ArgKey = "verbose"
	keyUrls     ArgKey = "urls"
	keyThreads  ArgKey = "threads"
	keyOutput   ArgKey = "output"
	keyReport   ArgKey = "report"
	keyHeader   ArgKey = "header"
	keyContinue ArgKey = "continue"
	keyPayload  ArgKey = "payload"
	keyCommand  ArgKey = "command"
	keyDelay    ArgKey = "delay"
	keyTimeout  ArgKey = "timeout"
)

// ArgKeys is a "named enum" collection for reference
var ArgKeys = struct {
	Method   ArgKey
	Protocol ArgKey
	Verbose  ArgKey
	Urls     ArgKey
	Threads  ArgKey
	Output   ArgKey
	Report   ArgKey
	Header   ArgKey
	Continue ArgKey
	Payload  ArgKey
	Command  ArgKey
	Delay    ArgKey
	Timeout  ArgKey
}{
	Method:   keyMethod,
	Protocol: keyProtocol,
	Verbose:  keyVerbose,
	Urls:     keyUrls,
	Threads:  keyThreads,
	Output:   keyOutput,
	Report:   keyReport,
	Header:   keyHeader,
	Continue: keyContinue,
	Payload:  keyPayload,
	Command:  keyCommand,
	Delay:    keyDelay,
	Timeout:  keyTimeout,
}

// ArgsService holds the map of arguments and provides methods to interact with them
type ArgsService struct {
	argsMap map[ArgKey]interface{}
}

// NewArgsService initializes the ArgsService and parses the command-line arguments.
var argServiceInstance *ArgsService

func GetArgsService() (*ArgsService, error) {
	if argServiceInstance == nil {
		argServiceInstance = &ArgsService{}
		err := argServiceInstance.Init()
		if err != nil {
			return nil, err
		}
	}
	return argServiceInstance, nil
}

// It inits args and returns a pointer to the ArgsService and an error if any.
func (as *ArgsService) Init() error {
	parser := argparse.NewParser("xss-scanner", "A scanning tool for xss vulnerabilities")

	// Define positional arguments
	command := parser.SelectorPositional([]string{"scan", "report"}, &argparse.Options{Help: "Command to execute: 'scan' or 'report'", Required: true})

	// Define flags for "scan" command
	method := parser.String("m", "method", &argparse.Options{Help: "HTTP method to use (GET, POST, PUT, etc.)", Default: "GET"})
	protocol := parser.String("", "protocol", &argparse.Options{Help: "HTTP protocol to use (http/https)", Default: "https"})

	urls := parser.String("u", "urls", &argparse.Options{Help: "URLs to scan for GET method, request for other methods", Default: "urls.txt"})
	payloads := parser.String("p", "payloads", &argparse.Options{Help: "Payloads to use", Default: "payloads.txt"})

	threads := parser.Int("T", "threads", &argparse.Options{Help: "Number of threads to run scans in parallel", Default: 10})
	continueFrom := parser.Int("c", "continue", &argparse.Options{Help: "Offset to continue scan from", Default: 0})

	verbose := parser.String("v", "verbose", &argparse.Options{Help: "Log level (ALL, LOG, INFO, WARN, ERROR)", Default: "ALL"})
	headers := parser.List("H", "header", &argparse.Options{Help: "Headers to append to GET request, multiple allowed", Default: []string{}})

	timeout := parser.Int("t", "timeout", &argparse.Options{Help: "Timeout for requests in ms", Default: 5000})
	delay := parser.Int("d", "delay", &argparse.Options{Help: "Delay between requests", Default: 0})

	output := parser.String("o", "output", &argparse.Options{Help: "Output file"})
	report := parser.String("r", "report", &argparse.Options{Help: "Report file"})

	// Parse the arguments
	err := parser.Parse(os.Args)
	if err != nil {
		return err
	}

	// Initialize the argsMap with parsed values
	argsMap := make(map[ArgKey]interface{})
	argsMap[ArgKeys.Command] = *command

	// Common for both commands
	argsMap[ArgKeys.Report] = *report
	argsMap[ArgKeys.Output] = *output

	if *command == "scan" {
		// unique for scan command
		argsMap[ArgKeys.Urls] = *urls
		argsMap[ArgKeys.Payload] = *payloads
		argsMap[ArgKeys.Verbose] = *verbose
		argsMap[ArgKeys.Threads] = *threads
		argsMap[ArgKeys.Protocol] = *protocol
		argsMap[ArgKeys.Method] = strings.ToUpper(*method)
		argsMap[ArgKeys.Header] = *headers
		argsMap[ArgKeys.Continue] = *continueFrom
		argsMap[ArgKeys.Delay] = *delay
		argsMap[ArgKeys.Timeout] = *timeout
	} else if *command == "report" {
		// unique for report command
		argsMap[ArgKeys.Report] = *report
	}

	as.argsMap = argsMap
	return nil
}

// Set sets a value in the argsMap
func (s *ArgsService) Set(key ArgKey, value interface{}) {
	s.argsMap[key] = value
}

// Get retrieves a value from the argsMap
func (s *ArgsService) Get(key ArgKey) (interface{}, error) {
	value, exists := s.argsMap[key]
	if !exists {
		return nil, errors.New("key not found in argsMap")
	}
	return value, nil
}

// Has checks if a key exists in the argsMap
func (s *ArgsService) Has(key ArgKey) bool {
	_, exists := s.argsMap[key]
	return exists
}

func (s *ArgsService) GetAll() map[ArgKey]interface{} {
	return s.argsMap
}
