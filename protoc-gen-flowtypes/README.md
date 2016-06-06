protoc-gen-flowtypes
====================

Generate flowtype type definitions for proto3 messages and enums.

```sh
$ cat simple.proto
```
```proto
syntax = "proto3";

message SearchRequest {
  string Query = 1;
  int32 limit = 2;
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
  repeated string Results = 1;
  int32 NumResults = 2;
  SearchRequest OriginalRequest = 3;
}
```
```sh
$ protoc -I. --flowtypes_out=. simple.proto
$ cat simple_types.js
```
```js
/* @flow */
export type simpleCorpus = "UNIVERSAL" | "WEB" | "IMAGES" | "LOCAL" | "NEWS" | "PRODUCTS" | "VIDEO";

export type SearchRequest = {
  Query?: string,
  limit?: number,
  corpus?: Corpus
};

export type SearchResponse = {
  Results?: []string,
  NumResults?: number,
  OriginalRequest?: SearchRequest
};
```
