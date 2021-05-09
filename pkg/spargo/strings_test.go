package spargo

var testQuery = `select distinct ?s ?label where {
	?s <http://the-fr.org/prop/format-registry/formatType> <http://the-fr.org/def/format-registry/RasterImage> .
	?s <http://www.w3.org/2000/01/rdf-schema#label> ?label .
} limit 10`

var testString = `{
   "head":{
      "vars":[
         "s",
         "label"
      ]
   },
   "results":{
      "bindings":[
         {
            "s":{
               "type":"uri",
               "value":"http://the-fr.org/id/file-format/25"
            },
            "label":{
               "type":"literal",
               "value":"OS/2 Bitmap",
               "xml:lang":"en"
            }
         },
         {
            "s":{
               "type":"uri",
               "value":"http://the-fr.org/id/file-format/28"
            },
            "label":{
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
