package spargo

/*
Basic SPARQL results will be returned as follows:

	{
	   "head":{
	      "vars":[
	         "item",
	         "itemLabel"
	      ]
	   },
	   "results":{
	      "bindings":[
	         {
	            "predicate_one":{
	               "type":"uri",
	               "value":"http://www.wikidata.org/entity/Q28114535"
	            },
	            "predicate_two":{
	               "xml:lang":"en",
	               "type":"literal",
	               "value":"Mr. White"
	            }
	         },
	         {
	            "predicate_one":{
	               "type":"uri",
	               "value":"http://www.wikidata.org/entity/Q28665865"
	            },
	            "predicate_two":{
	               "xml:lang":"en",
	               "type":"literal",
	               "value":"Мyka"
	            }
	         }
	      ]
	   }
	}

*/

type item struct {
	Lang     string `json:"xml:lang"` // Populated if requested in query.
	Type     string // Can be "uri", "literal"
	Value    string
	DataType string
}

type binding struct {
	Bindings []map[string]item
}

// SPARQLResult packages a SPARQL response from an endpoint.
type SPARQLResult struct {
	Head    map[string]interface{}
	Results binding
	Human   string
}
