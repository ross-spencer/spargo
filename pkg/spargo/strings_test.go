package spargo

var testQuery = `select distinct ?format ?label where {
	?format <http://the-fr.org/prop/format-registry/formatType> <http://the-fr.org/def/format-registry/RasterImage> .
	?format <http://www.w3.org/2000/01/rdf-schema#label> ?label .
} limit 10`

var testString = `{
   "head":{
      "vars":[
         "format",
         "label"
      ]
   },
   "results":{
      "bindings":[
         {
            "format":{
               "type":"uri",
               "value":"http://the-fr.org/id/file-format/25"
            },
            "label":{
               "datatype" : "http://example.com/DataTypes#unicode",
               "type":"literal",
               "value":"OS/2 Bitmap",
               "xml:lang":"en"
            }
         },
         {
            "format":{
               "type":"uri",
               "value":"http://the-fr.org/id/file-format/28"
            },
            "label":{
               "datatype" : "http://example.com/DataTypes#unicode",
               "type":"literal",
               "value":"CALS Compressed Bitmap",
               "xml:lang":"en"
            }
         }
      ]
   }
}`

// testEmptyResult is covered implicitly in the code when we compare the
// result of error conditions. We just want to make sure that for a
// properly empty result, e.g. no SPARQL results, we get something back
// that is properly well-formed.
var testEmptyResult = `{
  "head": null,
  "results": {
    "bindings": null
  }
}`
