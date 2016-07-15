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


