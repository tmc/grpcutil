// Code generated by protoc-gen-tstypes. DO NOT EDIT.

declare namespace simple {

    export enum SearchRequest_Corpus {
        UNIVERSAL = 0,
        WEB = 1,
        IMAGES = 2,
        LOCAL = 3,
        NEWS = 4,
        PRODUCTS = 5,
        VIDEO = 6,
    }
    export interface SearchRequest_XyzEntry {
        key?: string;
        value?: number;
    }

    // SearchRequest is an example type representing a search query.
    export interface SearchRequest {
        query?: string;
        page_number?: number;
        // Number of results per page.
        result_per_page?: number; // Should never be zero.
        corpus?: SearchRequest_Corpus;
        sent_at?: google.protobuf.Timestamp;
        xyz?: { [key: string]: number };
        zytes?: Uint8Array;
        example_required: number;
    }

    export interface SearchResponse {
        results: Array<string>;
        num_results: number;
        original_request: SearchRequest;
        next_results_uri?: string;
    }

}

