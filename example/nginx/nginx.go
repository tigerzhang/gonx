package main

import (
	gonx "../.."
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"strconv"
)

var conf string
var format string
var logFile string

func init() {
	flag.StringVar(&conf, "conf", "dummy", "Nginx config file (e.g. /etc/nginx/nginx.conf)")
	flag.StringVar(&format, "format", "main", "Nginx log_format name")
	flag.StringVar(&logFile, "log", "dummy", "Log file name to read. Read from STDIN if file name is '-'")
}

func main() {
	flag.Parse()

	// Read given file or from STDIN
	var logReader io.Reader
	var err error
	if logFile == "dummy" {
		logReader = strings.NewReader(`89.234.89.123 [08/Nov/2013:13:39:18 +0000] "GET /api/foo/bar HTTP/1.1"`)
	} else if logFile == "-" {
		logReader = os.Stdin
	} else {
		file, err := os.Open(logFile)
		if err != nil {
			panic(err)
		}
		logReader = file
		defer file.Close()
	}

	// Use nginx config file to extract format by the name
	var nginxConfig io.Reader
	if conf == "dummy" {
		nginxConfig = strings.NewReader(`
            http {
                log_format   main  '$remote_addr [$time_local] "$request"';
            }
        `)
	} else {
		nginxConfigFile, err := os.Open(conf)
		if err != nil {
			panic(err)
		}
		nginxConfig = nginxConfigFile
		defer nginxConfigFile.Close()
	}

	// Read from STDIN and use log_format to parse log records
	reader, err := gonx.NewNginxReader(logReader, nginxConfig, format)
	if err != nil {
		panic(err)
	}
	var bytes_sent_per_seconds map[string]int
	bytes_sent_per_seconds = make(map[string]int)
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		// Process the record... e.g.
		// fmt.Printf("Parsed entry: %+v\n", rec)
		timestamp, err := rec.Field("time_local")
		bytes_sent, err := rec.Field("body_bytes_sent")
		// fmt.Printf("timestamp %v bytes sent %v\n", timestamp, bytes_sent)
		i, _ := bytes_sent_per_seconds[timestamp] 
		bytes, err := strconv.Atoi(bytes_sent)
		i += bytes
		bytes_sent_per_seconds[timestamp] = i
	}

	for key, value := range bytes_sent_per_seconds {
		fmt.Println(key, value)
	}
}
