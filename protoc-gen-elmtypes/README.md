protoc-gen-elmtypes
====================

Generate elm type definitions for proto3 messages and enums.

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
```sh
$ protoc -I. --elmtypes_out=. simple.proto
$ cat simple_types.elm
```
```elm
-- this is a generated file
module Simple exposing (..)

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


