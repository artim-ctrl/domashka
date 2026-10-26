package health

const statusOK = "ok"

type Response struct {
	Status string `json:"status"`
}
