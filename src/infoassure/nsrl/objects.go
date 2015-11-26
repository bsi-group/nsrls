package main

// Holds the various objects/structs that are used in the system that don't warrant their own individual file

import (
	"github.com/voxelbrain/goptions"
)

// ##### Structs #############################################################

// Structure to store options
type Options struct {
	Mode 			string			`goptions:"-m, --mode, obligatory, description='Mode e.g. f (file) or s (server)'"`
	ConfigFile   	string      	`goptions:"-c, --config, description='Config file path'"`
	DataFile  		string        	`goptions:"-d, --data, obligatory, description='Data file path'"`
	InputFile  		string        	`goptions:"-i, --input, description='Input file path'"`
	OutputFile  	string        	`goptions:"-o, --output, description='Output file path'"`
	RemoveQuotes 	bool           	`goptions:"-r, --removequotes, description='Remove quotes if CSV data is quoted'"`
	CsvField     	int           	`goptions:"-s, --csvfield, description='CSV field to use'"`
	CsvDelimiter 	string       	`goptions:"-l, --csvdelimiter, description='CSV delimiter'"`
	Format 			string       	`goptions:"-f, --format, description='Format for output e.g. i (identified), u (unidentified) or a (all)'"`
	Help         	goptions.Help	`goptions:"-h, --help, description='Show this help'"`
}

// Stores the YAML config file data
type Config struct {
	ApiIp			string	`yaml:"api_ip"`
	ApiPort			int16	`yaml:"api_port"`
	ShowRequests	bool	`yaml:"show_requests"`
}

// Struct to marshal the "data" to a JSON string for the API
type JsonResult struct {
	Hash 	string		`db:"hash"`
	Exists	bool		`db:"exists"`
}