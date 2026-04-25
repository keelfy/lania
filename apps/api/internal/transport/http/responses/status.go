package responses

type StatusRes struct {
	API            string `json:"api"`
	Database       string `json:"database"`
	Cache          string `json:"cache"`
	Authentication string `json:"authentication"`
}
