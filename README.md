# spargo

SPARQL helper library for Golang

[![Build Status](https://travis-ci.org/ross-spencer/spargo.svg?branch=master)](https://travis-ci.org/ross-spencer/spargo)
[![GoDoc](https://godoc.org/github.com/ross-spencer/spargo?status.svg)](https://godoc.org/github.com/ross-spencer/spargo)
[![Go Report Card](https://goreportcard.com/badge/github.com/ross-spencer/spargo/pkg/spargo)](https://goreportcard.com/report/github.com/ross-spencer/spargo/pkg/spargo)

## .sparql File Format

The .sparql file format can be written as follows, note `#!spargo` is a magic
number when used like this.

```
#!spargo

ENDPOINT=...

# Comment
{sparql query}
```

## spargo interpreter

Borrowing from the above, with spargo reachable via a path such as
`/usr/bin/spargo` then a .sparql file can be configured like an interpretable
script, so:

```
#!/usr/bin/spargo

ENDPOINT=...

# Comment
{sparql query}
```

Lets call it `mysparql.sparql`. It can be run with executable permissions:

```
$ ./mysparql.sparql
```

And a JSON response will be returned to the caller.

## spargo Command

The `spargo` command primarily supports piped input, and there are some example
queries that can help demonstrate that.

In the spargo cmd folder, one can do the following:

```
github.com/ross-spencer/spargo/cmd/spargo$ cat examples/5-describe-wikidata.sparql | ./spargo

Connecting to: https://query.wikidata.org/sparql

Query: #! spargo
# Describe JPEG2000 in Wikidata database.
describe wd:Q931783

...{result}...
```

## spargo Package

The important part of this repository is the `spargo` package. To use it we
can do something like as follows once it is imported.

```golang
package ...

import (
	...
	...

	"github.com/ross-spencer/spargo/pkg/spargo"
)

func ...() {
	sparqlMe := spargo.SPARQLClient{}
	sparqlMe.ClientInit(url, queryString)
	res := sparqlMe.SPARQLGo().Human
}
```

And the results will be available in the `res` variable to be consumed by your
application.

## License

Apache License 2.0. More info [here](LICENSE).
