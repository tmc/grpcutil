/* @flow */
export type SearchRequestCorpus = "UNIVERSAL" | "WEB" | "IMAGES" | "LOCAL" | "NEWS" | "PRODUCTS" | "VIDEO";

export type SearchRequest = {
  query?: string,
  page_number?: number,
  result_per_page?: number,
  corpus?: "UNIVERSAL" | "WEB" | "IMAGES" | "LOCAL" | "NEWS" | "PRODUCTS" | "VIDEO"
};

export type SearchResponse = {
  results?: Array<string>,
  num_results?: number,
  original_request?: SearchRequest
};

