/* @flow */
export type Corpus = "UNIVERSAL" | "WEB" | "IMAGES" | "LOCAL" | "NEWS" | "PRODUCTS" | "VIDEO";

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

