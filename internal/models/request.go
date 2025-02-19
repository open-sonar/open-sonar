package models

// user's query request
type ChatRequest struct {
    Query      string `json:"query"`       
    NeedSearch bool   `json:"need_search"` 
    Pages      int    `json:"pages"`      
    Retries    int    `json:"retries"`    
}
