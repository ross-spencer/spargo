package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ross-spencer/spargo/pkg/spargo"
)

// SHEBANG provides some way of recognizing a .sparql file compatible with
// spargo, aka. our .sparql magic number.
var SHEBANG = []string{"#!spargo", "#!/usr/bin/spargo"}

// ENDPOINT must be specified in a .sparql file so that a query can be sent to
// the appropriate SPARQL endpoint.
const ENDPOINT string = "ENDPOINT"

var (
	vers     bool
	query    string
	endpoint string
)

func init() {
	flag.StringVar(&endpoint, "endpoint", "", "endpoint to query")
	flag.StringVar(&query, "query", "", "sparql query to run")
	flag.BoolVar(&vers, "version", false, "Return version")
}

// Check for spargo shebang.
func matchShebang(needle string, slice []string) bool {
	for _, item := range slice {
		if item == needle {
			return true
		}
	}
	return false
}

// Extract the query from the .sparql input.
func extractQuery(sparqlFile string) (string, string, error) {
	var shebang, url, queryString string
	var err error
	for _, line := range strings.Split(sparqlFile, "\n") {

		if line == "" {
			// Pass.
		} else if matchShebang(line, SHEBANG) {
			shebang = line
		} else if strings.Contains(strings.ToUpper(line), ENDPOINT) {
			_url := strings.SplitN(line, "=", 2)
			if len(_url) < 2 {
				err = fmt.Errorf("incorrect endpoint formatting: %s", line)
			}
			// TODO: validate the URL.
			url = strings.TrimSpace(_url[1])
		} else {
			queryString = queryString + line + "\n"
		}
	}
	if shebang == "" {
		err = fmt.Errorf("shebang '%s' is empty or incorrect", shebang)
	}
	return url, queryString, err
}

// TODO: Use a better pattern to parse the input of a SPARQL file...
func runQuery(sparqlFile string) {
	url, queryString, err := extractQuery(sparqlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Connecting to: %s\n\n", url)
	fmt.Fprintf(os.Stderr, "Query: %s\n\n", queryString)

	sparqlMe := spargo.SPARQLClient{}
	sparqlMe.ClientInit(url, queryString)
	res := sparqlMe.SPARQLGo()

	fmt.Println(res.Human)
}

func isPipeInput() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if (info.Mode() & os.ModeNamedPipe) != 0 {
		return true
	}
	return false
}

// interpreterInput tests for a file as the second argument in a call to
// spargo.
//
// TODO: there may be another pattern here using Open, but we are also
//       anticipating other arguments to the program at different times, so...
//
func interpreterInput() (bool, string) {
	if len(os.Args) == 2 {
		sparql := os.Args[1]
		if _, err := os.Stat(sparql); err == nil {
			return true, sparql
		} else if os.IsNotExist(err) {
			// Does not exist.
		} else {
			// Another error.
		}
	}
	return false, ""
}

func handlePipedInput() string {
	reader := bufio.NewReader(os.Stdin)
	var output []rune
	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}
	return string(output)
}

func handleInterpreterInput(sparql string) string {
	data, err := ioutil.ReadFile(sparql)
	if err != nil {
		return ""
	}
	return string(data)
}

func main() {
	// Parse our input and let spargo generate a response.
	if isPipeInput() {
		queryString := handlePipedInput()
		runQuery(queryString)
		os.Exit(0)
	} else {
		_, sparql := interpreterInput()
		if sparql != "" {
			query := handleInterpreterInput(sparql)
			runQuery(query)
			os.Exit(0)
		}
	}
	flag.Parse()
	if vers {
		fmt.Fprintf(os.Stderr, "%s (%s)\n", version(), spargo.DefaultAgent)
		os.Exit(0)
	} else if flag.NFlag() == 0 {
		fmt.Fprintln(os.Stderr, "Usage:  spargo {options}              ")
		fmt.Fprintln(os.Stderr, "               OPTIONAL: [-sparql] ...")
		fmt.Fprintln(os.Stderr, "               OPTIONAL: [-query]  ...")
		fmt.Fprintln(os.Stderr, "               OPTIONAL: [-version]   ")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Output: [JSON]   {url}")
		fmt.Fprintf(os.Stderr, "Output: [STRING] '%s (%s) ...'\n\n", version(), spargo.DefaultAgent)
		flag.Usage()
		os.Exit(0)
	} else {
		fmt.Println("Welcome to spargo: arg handling is not yet implemented. Take a look at the README.md for examples on how to used spargo with piped input...")
		fmt.Println("\nDebug, inputs:\n")
		fmt.Printf("   * SPARQL: '%s' \n", endpoint)
		fmt.Printf("   * Query: '%s' \n", query)
		fmt.Println("")
		fmt.Println("Take a look at the README.md for examples on how to used spargo with piped input...")
	}
}
