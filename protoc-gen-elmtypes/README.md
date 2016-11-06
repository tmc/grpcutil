protoc-gen-elmtypes
====================

Generate elm type definitions for proto3 messages and enums.

Contributions welcome.

```sh
$ cat simple.proto
```
```proto
syntax = "proto3";

message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
  enum Corpus {
    UNIVERSAL = 0;
    WEB = 1;
    IMAGES = 2;
    LOCAL = 3;
    NEWS = 4;
    PRODUCTS = 5;
    VIDEO = 6;
  }
  Corpus corpus = 4;
}

message SearchResponse {
  repeated string results = 1;
  int32 num_results = 2;
  SearchRequest original_request = 3;
}
```

# Example use
```sh
$ protoc -I. --elmtypes_out=. simple.proto
```
This generates Simple.elm.

## [Simple.elm](Simple.elm)
```elm
-- this is a generated file
module Simple exposing (..)
import Json.Encode as JE
import Json.Decode exposing (..)

type SearchRequestCorpus = UNIVERSAL | WEB | IMAGES | LOCAL | NEWS | PRODUCTS | VIDEO

type alias SearchRequest = {
  query: Maybe String,
  page_number: Maybe Int,
  result_per_page: Maybe Int,
  corpus: Maybe SearchRequestCorpus
}

type alias SearchResponse = {
  results: Maybe List String,
  num_results: Maybe Int,
  original_request: Maybe SearchRequest
}



(maybe ("SearchRequestCorpus" := UNIVERSAL | WEB | IMAGES | LOCAL | NEWS | PRODUCTS | VIDEO))

searchRequest = object4 SearchRequest (maybe ("query" := string)) (maybe ("page_number" := int)) (maybe ("result_per_page" := int)) (maybe ("corpus" := SearchRequestCorpus))

searchResponse = object3 SearchResponse (maybe ("results" := (list string))) (maybe ("num_results" := int)) (maybe ("original_request" := SearchRequest))
````
