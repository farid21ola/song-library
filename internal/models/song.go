package models

type SongInfo struct {
	ReleaseDate string `json:"release_date,omitempty" example:"16.07.2006"`
	Text        string `json:"text,omitempty" example:"Ooh baby, don't you know I suffer?\\nOoh baby, canyou hear me moan?\\nYou caught me under false pretenses\\nHow long before you let me go?\\n\\nOoh\\nYou set my soul alight\\nOoh\\nYou set my soul alight"`
	Link        string `json:"link,omitempty" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
}

type Song struct {
	Artist      string `json:"group" validate:"required" example:"Muse"`
	Title       string `json:"song" validate:"required" example:"Supermassive Black Hole"`
	ReleaseDate string `json:"release_date,omitempty" example:"16.07.2006"`
	Text        string `json:"text,omitempty" example:"Ooh baby, don't you know..."`
	Link        string `json:"link,omitempty" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
}
