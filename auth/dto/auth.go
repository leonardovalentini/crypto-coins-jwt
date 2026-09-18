package dto

type Response struct {
	Message string `json:"message" xml:"message"`
}

type ResponseCreated struct {
	Id int `json:"id" xml:"id"`
}

type VerifyResponse struct {
	Id string `json:"id" xml:"id"`
}
