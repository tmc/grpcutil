protoc-gen-flowtypes
====================

Generate flowtype type definitions for proto3 messages and enums.

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
$ protoc -I. --flowtypes_out=. simple.proto
$ cat simpleTypes.js
```
```js
/* @flow */
export type SearchRequestCorpus = "UNIVERSAL" | "WEB" | "IMAGES" | "LOCAL" | "NEWS" | "PRODUCTS" | "VIDEO";

export type SearchRequest = {
  query?: string,
  page_number?: number,
  result_per_page?: number,
  corpus?: SearchRequestCorpus
};

export type SearchResponse = {
  results?: string[],
  num_results?: number,
  original_request?: SearchRequest
};
```
