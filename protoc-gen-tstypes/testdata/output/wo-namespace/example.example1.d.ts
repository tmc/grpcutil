// Code generated by protoc-gen-tstypes. DO NOT EDIT.

export enum SearchRequest_Corpus {
    UNIVERSAL = "UNIVERSAL",
    WEB = "WEB",
    IMAGES = "IMAGES",
    LOCAL = "LOCAL",
    NEWS = "NEWS",
    PRODUCTS = "PRODUCTS",
    VIDEO = "VIDEO",
}
export interface SearchRequest_XyzEntry {
    key?: string;
    value?: number;
}

export interface SearchRequest {
    query?: string;
    page_number?: number;
    result_per_page?: number;
    corpus?: SearchRequest_Corpus;
    sent_at?: google.protobuf.Timestamp;
    xyz?: { [key: string]: number };
    zytes?: Uint8Array;
}

export interface SearchResponse {
    results?: Array<string>;
    num_results?: number;
    original_request?: SearchRequest;
}

